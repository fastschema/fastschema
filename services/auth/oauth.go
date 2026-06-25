package authservice

import (
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/errors"
)

func (as *AuthService) Me(c fs.Context, _ any) (*fs.User, error) {
	if c.User() == nil {
		return nil, errors.Unauthorized()
	}

	user, err := db.Builder[*fs.User](as.DB()).Where(db.EQ("id", c.User().ID)).Only(c)
	if err != nil {
		if db.IsNotFound(err) {
			return nil, errors.Unauthorized()
		}
		return nil, err
	}

	return user, nil
}

func (as *AuthService) Login(c fs.Context, _ any) (_ any, err error) {
	provider := as.GetAuthProvider(c.Arg("provider"))
	if provider == nil {
		return nil, errors.NotFound("invalid auth provider")
	}

	return provider.Login(c)
}

func (as *AuthService) Callback(c fs.Context, _ any) (u *fs.JWTTokens, err error) {
	provider := as.GetAuthProvider(c.Arg("provider"))
	if provider == nil {
		return nil, errors.NotFound("invalid auth provider")
	}

	user, err := provider.Callback(c)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.Unauthorized("invalid user")
	}

	return as.createLoginResponse(c, user)
}

func (as *AuthService) VerifyIDToken(c fs.Context, payload fs.IDToken) (u *fs.JWTTokens, err error) {
	provider := as.GetAuthProvider(c.Arg("provider"))
	if provider == nil {
		return nil, errors.NotFound("invalid auth provider")
	}

	user, err := provider.VerifyIDToken(c, payload)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, errors.Unauthorized("invalid user")
	}

	return as.createLoginResponse(c, user)
}

func (as *AuthService) createLoginResponse(c fs.Context, providerUser *fs.User) (*fs.JWTTokens, error) {
	loginUser, err := db.Builder[*fs.User](as.DB()).
		Where(db.And(
			db.EQ("provider", providerUser.Provider),
			db.EQ("provider_id", providerUser.ProviderID),
		)).
		Select("id", "username", "email", "provider", "provider_id", "provider_username", "active", "roles").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		return nil, err
	}

	if loginUser == nil {
		if loginUser, err = as.createUser(c, providerUser); err != nil {
			return nil, err
		}
	}

	return as.GenerateJWTTokens(c, loginUser)
}

func (as *AuthService) createUser(c fs.Context, providerUser *fs.User) (*fs.User, error) {
	// Check for existing user with same email but different provider
	duplicateEmailUser, err := db.Builder[*fs.User](as.DB()).
		Where(db.And(
			db.EQ("email", providerUser.Email),
			db.NEQ("provider", providerUser.Provider),
		)).
		Select("id", "email", "provider").
		First(c)
	if err != nil && !db.IsNotFound(err) {
		return nil, err
	}

	if duplicateEmailUser != nil {
		return nil, errors.Unauthorized(auth.MSG_EXISTING_USER_WITH_EMAIL)
	}

	// Resolve the User role by name from cache; role name is the stable identifier across deployments.
	// Fail closed: refuse to create an orphan user if the role is missing from cache.
	userRole := as.RoleByName(fs.RoleUser.Name)
	if userRole == nil {
		c.Logger().Errorf("role '%s' not found in cache; cannot create new social user", fs.RoleUser.Name)
		return nil, errors.InternalServerError("user role not available")
	}

	userEntity := fs.Map{
		"provider":          providerUser.Provider,
		"provider_id":       providerUser.ProviderID,
		"provider_username": providerUser.ProviderUsername,
		"username":          strings.TrimSpace(providerUser.Username),
		"email":             strings.TrimSpace(providerUser.Email),
		"active":            true,
		"roles":             []*entity.Entity{entity.New(userRole.ID)},
	}

	if providerUser.FirstName != "" {
		userEntity["first_name"] = providerUser.FirstName
	}

	if providerUser.LastName != "" {
		userEntity["last_name"] = providerUser.LastName
	}

	if providerUser.ProviderProfileImage != "" {
		userEntity["provider_profile_image"] = providerUser.ProviderProfileImage
	}

	newUser, err := db.Create[*fs.User](c, as.DB(), userEntity)
	if err != nil {
		return nil, err
	}

	// Carry the resolved role object so GenerateJWTTokens encodes the real role ID in the token.
	newUser.Roles = []*fs.Role{userRole}

	return newUser, nil
}
