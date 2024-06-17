package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

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
	oauth          *oauth2.Config
	accessTokenURL string
	userInfoURL    string
}

func NewGithubAuthProvider(config Config, redirectURL string) (fs.AuthProvider, error) {
	if config["client_id"] == "" || config["client_secret"] == "" {
		return nil, fmt.Errorf("github client id or secret is not set")
	}

	githubAuthProvider := &GithubAuthProvider{
		accessTokenURL: config["access_token_url"],
		userInfoURL:    config["user_info_url"],
		oauth: &oauth2.Config{
			ClientID:     config["client_id"],
			ClientSecret: config["client_secret"],
			RedirectURL:  redirectURL,
			Endpoint:     github.Endpoint,
		},
	}

	if githubAuthProvider.accessTokenURL == "" {
		githubAuthProvider.accessTokenURL = "https://github.com/login/oauth/access_token"
	}

	if githubAuthProvider.userInfoURL == "" {
		githubAuthProvider.userInfoURL = "https://api.github.com/user"
	}

	return githubAuthProvider, nil
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
		return nil, fmt.Errorf("github auth: callback code is empty")
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

	userResponse, err := utils.SendRequest[GithubUserResponse](
		"GET",
		ga.userInfoURL,
		Config{"Authorization": fmt.Sprintf("token %s", accessToken)},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &userResponse, nil
}

func (ga *GithubAuthProvider) getAccessToken(code string) (string, error) {
	requestBody, _ := json.Marshal(Config{
		"code":          code,
		"client_id":     ga.oauth.ClientID,
		"client_secret": ga.oauth.ClientSecret,
	})

	accessTokenResponse, err := utils.SendRequest[GithubAccessTokenResponse](
		"POST",
		ga.accessTokenURL,
		Config{
			"Content-Type": "application/json",
			"Accept":       "application/json",
		},
		bytes.NewBuffer(requestBody),
	)
	if err != nil {
		return "", err
	}

	return accessTokenResponse.AccessToken, nil
}
