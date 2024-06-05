package authservice

// import (
// 	"bytes"
// 	"encoding/json"
// 	"fmt"
// 	"io/ioutil"
// 	"net/http"

// 	"github.com/fastschema/fastschema/fs"
// 	userservice "github.com/fastschema/fastschema/services/user"

// 	"golang.org/x/oauth2"
// )

// // const (
// // 	GITHUB_ACCESS_TOKEN_URL = "https://github.com/login/oauth/access_token"
// // 	GITHUB_USER_URL         = "https://api.github.com/user"
// // )

// type GithubAccessTokenResponse struct {
// 	Scope       string `json:"scope"`
// 	TokenType   string `json:"token_type"`
// 	AccessToken string `json:"access_token"`
// }

// type GithubUserResponse struct {
// 	Login     string `json:"login"`
// 	ID        int    `json:"id"`
// 	AvatarURL string `json:"avatar_url"`
// 	Name      string `json:"name"`
// 	Blog      string `json:"blog"`
// 	Email     string `json:"email"`
// 	Bio       string `json:"bio"`
// }

// func (as *AuthService) Name() string {
// 	return "github"
// }

// func (as *AuthService) GetGithubAccessToken(code string) (string, error) {
// 	requestBody := map[string]string{
// 		"code":          code,
// 		"client_id":     as.OAuthGithub.config.ClientID,
// 		"client_secret": as.OAuthGithub.config.ClientSecret,
// 	}
// 	requestJSON, _ := json.Marshal(requestBody)
// 	req, err := http.NewRequest(
// 		"POST",
// 		as.AuthConfigs.Providers.Github.AccessTokenURL,
// 		bytes.NewBuffer(requestJSON),
// 	)

// 	if err != nil {
// 		return "", err
// 	}

// 	req.Header.Set("Content-Type", "application/json")
// 	req.Header.Set("Accept", "application/json")

// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		return "", err
// 	}

// 	var accessTokenResponse GithubAccessTokenResponse
// 	if err := json.Unmarshal(body, &accessTokenResponse); err != nil {
// 		return "", err
// 	}

// 	return accessTokenResponse.AccessToken, nil
// }

// func (as *AuthService) GetGithubUser(accessToken string) (*GithubUserResponse, error) {
// 	req, err := http.NewRequest(
// 		"GET",
// 		as.AuthConfigs.Providers.Github.UserURL,
// 		nil,
// 	)

// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
// 	resp, err := http.DefaultClient.Do(req)

// 	if err != nil {
// 		return nil, err
// 	}

// 	body, err := ioutil.ReadAll(resp.Body)

// 	if err != nil {
// 		return nil, err
// 	}

// 	userResponse := &GithubUserResponse{}
// 	if err := json.Unmarshal(body, userResponse); err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	return userResponse, nil
// }

// func (as *AuthService) GetGithubUserFromAccessCode(code string) (*GithubUserResponse, error) {
// 	accessToken, err := as.GetGithubAccessToken(code)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return as.GetGithubUser(accessToken)
// }

// func (as *AuthService) LoginGithub(c fs.Context, _ any) (err error) {
// 	url := as.OAuthGithub.config.AuthCodeURL(
// 		as.AuthConfigs.Providers.Github.StateCode,
// 		oauth2.AccessTypeOffline,
// 		oauth2.SetAuthURLParam("scope", "user:email"),
// 	)
// 	return c.Redirect(url)
// }

// func (as *AuthService) CallbackGithub(c fs.Context, _ any) (u *userservice.LoginResponse, err error) {
// 	if c.Arg("code") == "" {
// 		return nil, fmt.Errorf("code is empty")
// 	}

// 	if c.Arg("state") != as.AuthConfigs.Providers.Github.StateCode {
// 		return nil, fmt.Errorf("invalid oauth Github state")
// 	}

// 	githubUser, err := as.GetGithubUserFromAccessCode(c.Arg("code"))

// 	if err != nil {
// 		return nil, err
// 	}
// 	providerUsers := ProviderUsers{
// 		Github: *githubUser,
// 	}
// 	return as.processLoginResponse(c, providerUsers, Github)
// }
