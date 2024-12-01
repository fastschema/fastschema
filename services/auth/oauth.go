package authservice

import (
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
)

type LoginResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

func (as *AuthService) Login(c fs.Context, _ any) (_ any, err error) {
	provider := as.GetAuthProvider(c.Arg("provider"))
	if provider == nil {
		return nil, errors.NotFound("invalid auth provider")
	}

	return provider.Login(c)
}

func (as *AuthService) Callback(c fs.Context, _ any) (u *LoginResponse, err error) {
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

	return as.createUser(c, user)
}

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

func (as *AuthService) createUser(c fs.Context, providerUser *fs.User) (*LoginResponse, error) {
	userExisted, err := db.Builder[*fs.User](as.DB()).
		Where(db.And(
			db.EQ("provider", providerUser.Provider),
			db.EQ("provider_id", providerUser.ProviderID),
		)).
		Only(c)

	if err != nil {
		// There is an error other than not found error
		if !db.IsNotFound(err) {
			return nil, err
		}

		// The error is not found, create a new user
		if userExisted, err = db.Create[*fs.User](c, as.DB(), fs.Map{
			"provider":          providerUser.Provider,
			"provider_id":       providerUser.ProviderID,
			"provider_username": providerUser.ProviderUsername,
			"username":          providerUser.Username,
			"email":             providerUser.Email,
			"active":            true,
			"roles":             []*entity.Entity{entity.New(fs.RoleUser.ID)},
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

	return &LoginResponse{Token: jwtToken, Expires: exp}, nil
}
