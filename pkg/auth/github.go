package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const ProviderGithub = "github"

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

func NewGithubAuthProvider(config fs.Map, redirectURL string) (fs.AuthProvider, error) {
	clientID := fs.MapValue(config, "client_id", "")
	clientSecret := fs.MapValue(config, "client_secret", "")
	if clientID == "" ||
		clientSecret == "" {
		return nil, errors.New("github client id or secret is not set")
	}

	githubAuthProvider := &GithubAuthProvider{
		accessTokenURL: fs.MapValue(config, "access_token_url", ""),
		userInfoURL:    fs.MapValue(config, "user_info_url", ""),
		oauth: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
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
	return ProviderGithub
}

func (ga *GithubAuthProvider) WithResources(resource *fs.Resource) {
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
		return nil, errors.New("github auth: callback code is empty")
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

func (ga *GithubAuthProvider) Form(c fs.Context) (_ *fs.User, err error) {
	return nil, errors.New("not implemented")
}

func (ga *GithubAuthProvider) getUser(code string) (*GithubUserResponse, error) {
	accessToken, err := ga.getAccessToken(code)
	if err != nil {
		return nil, err
	}

	userResponse, err := utils.SendRequest[GithubUserResponse](
		"GET",
		ga.userInfoURL,
		map[string]string{"Authorization": "token " + accessToken},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &userResponse, nil
}

func (ga *GithubAuthProvider) getAccessToken(code string) (string, error) {
	requestBody, err := json.Marshal(map[string]string{
		"code":          code,
		"client_id":     ga.oauth.ClientID,
		"client_secret": ga.oauth.ClientSecret,
	})
	if err != nil {
		return "", fmt.Errorf("failed to marshal access token request body: %w", err)
	}

	accessTokenResponse, err := utils.SendRequest[GithubAccessTokenResponse](
		"POST",
		ga.accessTokenURL,
		map[string]string{
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
