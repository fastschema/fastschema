package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const ProviderGoogle = "google"

type GoogleUserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GoogleAuthProvider struct {
	oauth       *oauth2.Config
	userInfoURL string
}

func NewGoogleAuthProvider(config fs.Map, redirectURL string) (fs.AuthProvider, error) {
	clientID := fs.MapValue(config, "client_id", "")
	clientSecret := fs.MapValue(config, "client_secret", "")
	if clientID == "" || clientSecret == "" {
		return nil, fmt.Errorf("google client id or secret is not set")
	}

	googleAuthProvider := &GoogleAuthProvider{
		userInfoURL: fs.MapValue(config, "user_info_url", ""),
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},
	}

	if config["access_token_url"] != "" {
		googleAuthProvider.oauth.Endpoint.TokenURL = fs.MapValue(config, "access_token_url", "")
	}

	if googleAuthProvider.userInfoURL == "" {
		googleAuthProvider.userInfoURL = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="
	}

	return googleAuthProvider, nil
}

func (as *GoogleAuthProvider) Name() string {
	return ProviderGoogle
}

func (as *GoogleAuthProvider) Login(c fs.Context) (_ any, err error) {
	state := utils.RandomString(16)
	url := as.oauth.AuthCodeURL(state)
	return nil, c.Redirect(url)
}

func (as *GoogleAuthProvider) Callback(c fs.Context) (_ *fs.User, err error) {
	// should check c.Arg("state") for invalid oauth Google state
	if c.Arg("code") == "" {
		return nil, fmt.Errorf("callback code is empty")
	}

	googleUser, err := as.getUser(c.Arg("code"))
	if err != nil {
		return nil, err
	}

	return &fs.User{
		Provider:         as.Name(),
		ProviderID:       googleUser.ID,
		ProviderUsername: googleUser.Email,

		Username: strings.Split(googleUser.Email, "@gmail")[0],
		Email:    googleUser.Email,
		Active:   true,
		RoleIDs:  []uint64{fs.RoleUser.ID},
		Roles:    []*fs.Role{fs.RoleUser},
	}, nil
}

func (as *GoogleAuthProvider) getUser(code string) (*GoogleUserResponse, error) {
	token, err := as.oauth.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("google auth code exchange error: %s", err.Error())
	}

	userResponse, err := utils.SendRequest[GoogleUserResponse](
		"GET",
		as.userInfoURL+token.AccessToken,
		map[string]string{},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &userResponse, nil
}
