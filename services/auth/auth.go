package authservice

import (
	// "errors"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	userservice "github.com/fastschema/fastschema/services/user"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type AppLike interface {
	DB() db.Client
	Key() string
}

type AuthProvider struct {
	config *oauth2.Config
}

const (
	Github = "github"
	Google = "google"
)

type ProviderUsers struct {
	Github GithubUserResponse `json:"github"`
	Google GoogleUserResponse `json:"google"`
}

type AuthService struct {
	DB          func() db.Client
	AppKey      func() string
	OAuthGithub AuthProvider
	OAuthGoogle AuthProvider
}

func New(app AppLike) *AuthService {
	authService := &AuthService{
		DB:     app.DB,
		AppKey: app.Key,
	}

	if utils.Env("GITHUB_PROVIDER_ENABLED") == "true" {
		if utils.Env("GITHUB_CLIENT_ID") == "" || utils.Env("GITHUB_CLIENT_SECRET") == "" {
			panic("Github client id or secret is not set")
		}
		authService.OAuthGithub = AuthProvider{
			config: &oauth2.Config{
				ClientID:     utils.Env("GITHUB_CLIENT_ID"),
				ClientSecret: utils.Env("GITHUB_CLIENT_SECRET"),
				RedirectURL:  utils.Env("APP_BASE_URL") + "/api/auth/github/callback",
				Endpoint:     github.Endpoint,
			},
		}
	}

	if utils.Env("GOOGLE_PROVIDER_ENABLED") == "true" {
		if utils.Env("GOOGLE_CLIENT_ID") == "" || utils.Env("GOOGLE_CLIENT_SECRET") == "" {
			panic("Google client id or secret is not set")
		}
		authService.OAuthGoogle = AuthProvider{
			config: &oauth2.Config{
				ClientID:     utils.Env("GOOGLE_CLIENT_ID"),
				ClientSecret: utils.Env("GOOGLE_CLIENT_SECRET"),
				Endpoint:     google.Endpoint,
				RedirectURL:  utils.Env("APP_BASE_URL") + "/api/auth/google/callback",
				Scopes: []string{
					"https://www.googleapis.com/auth/userinfo.email",
					"https://www.googleapis.com/auth/userinfo.profile",
				},
			},
		}
	}

	return authService
}

func (as *AuthService) Login(c fs.Context, _ any) (nil, err error) {

	switch c.Arg("provider") {
	case Github:
		return nil, as.LoginGithub(c, nil)
	case Google:
		return nil, as.LoginGoogle(c, nil)
	default:
		return nil, errors.New("invalid provider")
	}

}

func (as *AuthService) Callback(c fs.Context, _ any) (u *userservice.LoginResponse, err error) {
	if c.Arg("provider") == Github {
		return as.CallbackGithub(c, nil)
	}
	if c.Arg("provider") == Google {
		return as.CallbackGoogle(c, nil)
	}
	return nil, errors.New("invalid provider")
}

func (as *AuthService) processLoginResponse(c fs.Context, providerUsers ProviderUsers, provider string) (*userservice.LoginResponse, error) {
	var query *db.Predicate
	if provider == Github {
		query = db.EQ("username", providerUsers.Github.Login)
	} else if provider == Google {
		query = db.EQ("email", providerUsers.Google.Email)
	}

	userExisted, _ := db.Query[*fs.User](as.DB()).Where(query).First(c.Context())

	if userExisted != nil {
		if !userExisted.Active {
			return nil, errors.Unauthorized("user is not active")
		}
		jwtToken, exp, err := userExisted.JwtClaim(as.AppKey())
		if err != nil {
			return nil, err
		}

		return &userservice.LoginResponse{Token: jwtToken, Expires: exp}, nil
	}

	userRole, err := db.Query[*fs.Role](as.DB()).Where(db.EQ("name", "User")).First(c.Context())
	if err != nil {
		e := utils.If(db.IsNotFound(err), errors.NotFound, errors.InternalServerError)
		return nil, e(err.Error())
	}

	var userSaved *fs.User
	if provider == Github {
		userSaved, err = db.Create[*fs.User](c.Context(), as.DB(), schema.NewEntityFromMap(map[string]any{
			"username": providerUsers.Github.Login,
			"email":    providerUsers.Github.Email,
			"active":   true,
			"provider": Github,
			// "provider_id":       providerUsers.Github.ID,
			"provider_username": providerUsers.Github.Login,
			"roles": []*schema.Entity{
				schema.NewEntity(userRole.ID),
			},
		}))
	} else if provider == Google {
		userSaved, err = db.Create[*fs.User](c.Context(), as.DB(), schema.NewEntityFromMap(map[string]any{
			"username":    "", // need to fix to create user without username
			"email":       providerUsers.Google.Email,
			"active":      true,
			"provider":    Google,
			"provider_id": providerUsers.Google.ID,
			// "provider_username": providerUsers.Google.Name,
			"roles": []*schema.Entity{
				schema.NewEntity(userRole.ID),
			},
		}))
	}

	if err != nil {
		return nil, err
	}
	jwtToken, exp, err := userSaved.JwtClaim(as.AppKey())
	if err != nil {
		return nil, err
	}
	return &userservice.LoginResponse{Token: jwtToken, Expires: exp}, nil
}
