package authservice

import (
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/schema"
	userservice "github.com/fastschema/fastschema/services/user"
)

func (as *AuthService) Login(c fs.Context, _ any) (_ any, err error) {
	provider := as.GetAuthProvider(c.Arg("provider"))
	if provider == nil {
		return nil, errors.NotFound("invalid provider")
	}

	return provider.Login(c)
}

func (as *AuthService) Callback(c fs.Context, _ any) (u *userservice.LoginResponse, err error) {
	provider := as.GetAuthProvider(c.Arg("provider"))
	if provider == nil {
		return nil, errors.NotFound("invalid provider")
	}

	user, err := provider.Callback(c)
	if err != nil {
		return nil, errors.InternalServerError(err.Error())
	}

	return as.createUser(c, user)
}

func (as *AuthService) createUser(c fs.Context, providerUser *fs.User) (*userservice.LoginResponse, error) {
	userExisted, err := db.Query[*fs.User](as.DB()).
		Where(db.And(
			db.EQ("provider", providerUser.Provider),
			db.EQ("provider_id", providerUser.ProviderID),
		)).
		Only(c.Context())

	if err != nil {
		// There is an error other than not found error
		if !db.IsNotFound(err) {
			return nil, err
		}

		// The error is not found, create a new user
		if userExisted, err = db.Create[*fs.User](c.Context(), as.DB(), fs.Map{
			"provider":          providerUser.Provider,
			"provider_id":       providerUser.ProviderID,
			"provider_username": providerUser.ProviderUsername,
			"username":          providerUser.Username,
			"email":             providerUser.Email,
			"active":            true,
			"roles":             []*schema.Entity{schema.NewEntity(fs.RoleUser.ID)},
		}); err != nil {
			return nil, err
		}

		// Set the role of the user
		userExisted.RoleIDs = []uint64{fs.RoleUser.ID}
		userExisted.Roles = []*fs.Role{fs.RoleUser}
	}

	jwtToken, exp, err := userExisted.JwtClaim(as.AppKey())
	if err != nil {
		return nil, err
	}

	return &userservice.LoginResponse{Token: jwtToken, Expires: exp}, nil
}
