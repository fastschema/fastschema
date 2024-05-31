package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const GithubLoginOauthURL = "https://github.com/login/oauth/access_token"
const GithubUserURL = "https://api.github.com/user"

type GithubAccessTokenResponse struct {
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	AccessToken string `json:"access_token"`
}

type GithubUserResponse struct {
	Login     string `json:"login"`
	ID        int    `json:"id"`
	AvatarURL string `json:"avatar_url"`
	Name      string `json:"name"`
	Blog      string `json:"blog"`
	Email     string `json:"email"`
	Bio       string `json:"bio"`
}

type GithubAuthProvider struct {
	oauth *oauth2.Config
}

func NewGithubAuthProvider(
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
			RedirectURL:  redirectURL,
			Endpoint:     github.Endpoint,
		},
	}, nil
}

func (ga *GithubAuthProvider) Name() string {
	return "github"
}

func (ga *GithubAuthProvider) Login(c fs.Context) (_ any, err error) {
	state := utils.RandomString(16) // should replace with a cookie from context
	rediectURL := ga.oauth.AuthCodeURL(
		state,
		oauth2.AccessTypeOffline,
		oauth2.SetAuthURLParam("scope", "user:email"),
	)
	return nil, c.Redirect(rediectURL)
}

func (ga *GithubAuthProvider) Callback(c fs.Context) (_ *fs.User, err error) {
	// should check c.Arg("state") for invalid oauth Github state
	if c.Arg("code") == "" {
		return nil, fmt.Errorf("code is empty")
	}

	githubUser, err := ga.getUser(c.Arg("code"))
	if err != nil {
		return nil, err
	}

	return &fs.User{
		Provider:         ga.Name(),
		ProviderID:       strconv.Itoa(githubUser.ID),
		ProviderUsername: githubUser.Login,

		Username: githubUser.Login,
		Email:    githubUser.Email,
		Active:   true,
		RoleIDs:  []uint64{fs.RoleUser.ID},
		Roles:    []*fs.Role{fs.RoleUser},
	}, nil
}

func (ga *GithubAuthProvider) getUser(code string) (*GithubUserResponse, error) {
	accessToken, err := ga.getAccessToken(code)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", GithubUserURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	userResponse := &GithubUserResponse{}
	if err := json.Unmarshal(body, userResponse); err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return userResponse, nil
}

func (ga *GithubAuthProvider) getAccessToken(code string) (string, error) {
	requestBody, _ := json.Marshal(map[string]string{
		"code":          code,
		"client_id":     ga.oauth.ClientID,
		"client_secret": ga.oauth.ClientSecret,
	})
	req, err := http.NewRequest("POST", GithubLoginOauthURL, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var accessTokenResponse GithubAccessTokenResponse
	if err := json.Unmarshal(body, &accessTokenResponse); err != nil {
		return "", err
	}

	return accessTokenResponse.AccessToken, nil
}
