package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const oauthGoogleURLAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

type GoogleUserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

type GoogleAuthProvider struct {
	oauth *oauth2.Config
}

func NewGoogleAuthProvider(
	config map[string]string,
	redirectURL string,
) (fs.AuthProvider, error) {
	if config["client_id"] == "" || config["client_secret"] == "" {
		return nil, fmt.Errorf("github client id or secret is not set")
	}

	return &GithubAuthProvider{
		oauth: &oauth2.Config{
			ClientID:     config["client_id"],
			ClientSecret: config["client_secret"],
			Endpoint:     google.Endpoint,
			RedirectURL:  redirectURL,
			Scopes: []string{
				"https://www.googleapis.com/auth/userinfo.email",
				"https://www.googleapis.com/auth/userinfo.profile",
			},
		},
	}, nil
}

func (as *GoogleAuthProvider) Name() string {
	return "google"
}

func (as *GoogleAuthProvider) Login(c fs.Context, _ any) (_ any, err error) {
	state := utils.RandomString(16) // should replace with a cookie from context
	url := as.oauth.AuthCodeURL(state)
	return nil, c.Redirect(url)
}

func (as *GoogleAuthProvider) Callback(c fs.Context) (_ *fs.User, err error) {
	// should check c.Arg("state") for invalid oauth Github state
	if c.Arg("code") == "" {
		return nil, fmt.Errorf("code is empty")
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
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}

	response, err := http.Get(oauthGoogleURLAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}

	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}

	userResponse := &GoogleUserResponse{}
	if err := json.Unmarshal(body, userResponse); err != nil {
		return nil, err
	}

	return userResponse, nil
}
