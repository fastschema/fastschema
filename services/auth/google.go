package authservice

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fastschema/fastschema/fs"
	userservice "github.com/fastschema/fastschema/services/user"
)

const oauthGoogleUrlAPI = "https://www.googleapis.com/oauth2/v2/userinfo?access_token="

type GoogleUserResponse struct {
	ID      string `json:"id"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (as *AuthService) GetGoogleUserFromAccessCode(code string) (*GoogleUserResponse, error) {
	token, err := as.OAuthGoogle.config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(oauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}

	userResponse := &GoogleUserResponse{}
	if err := json.Unmarshal(body, userResponse); err != nil {
		return nil, err
	}

	return userResponse, nil
}

func (as *AuthService) LoginGoogle(c fs.Context, _ any) (err error) {
	url := as.OAuthGoogle.config.AuthCodeURL("randomstate")
	fmt.Println("url", url)
	return c.Redirect(url)
}

func (as *AuthService) CallbackGoogle(c fs.Context, _ any) (u *userservice.LoginResponse, err error) {
	if c.Arg("state") != "randomstate" {
		return nil, fmt.Errorf("invalid oauth google state")
	}

	if c.Arg("code") == "" {
		return nil, fmt.Errorf("code is empty")
	}

	googleUser, err := as.GetGoogleUserFromAccessCode(c.Arg("code"))

	fmt.Println("googleUser", googleUser)

	if err != nil {
		return nil, err
	}
	providerUsers := ProviderUsers{
		Google: *googleUser,
	}
	return as.processLoginResponse(c, providerUsers, Google)
}
