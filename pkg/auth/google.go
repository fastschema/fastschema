package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"cloud.google.com/go/auth/credentials/idtoken"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/utils"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

const ProviderGoogle = "google"

type GoogleUser struct {
	Issuer    string    `json:"iss"`
	Audience  string    `json:"aud"`
	ExpiresAt time.Time `json:"exp"`
	IssuedAt  time.Time `json:"iat"`

	ID            string `json:"id"` // Using token Subject as ID
	Email         string `json:"email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Locale        string `json:"locale"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
	HD            string `json:"hd"`
}

func (gu *GoogleUser) ToFSUser() *fs.User {
	return &fs.User{
		Provider:             ProviderGoogle,
		ProviderID:           gu.ID,
		ProviderUsername:     gu.Email,
		ProviderProfileImage: gu.Picture,

		Username:  gu.Email,
		Email:     gu.Email,
		FirstName: gu.GivenName,
		LastName:  gu.FamilyName,
		Active:    true,
		RoleIDs:   []uint64{fs.RoleUser.ID},
		Roles:     []*fs.Role{fs.RoleUser},
	}
}

type GoogleAuthProvider struct {
	oauth       *oauth2.Config
	userInfoURL string
}

func NewGoogleAuthProvider(config fs.Map, redirectURL string) (fs.AuthProvider, error) {
	clientID := fs.MapValue(config, "client_id", "")
	clientSecret := fs.MapValue(config, "client_secret", "")
	if clientID == "" || clientSecret == "" {
		return nil, errors.New("google client id or secret is not set")
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

	accessTokenUrl := fs.MapValue(config, "access_token_url", "")
	if accessTokenUrl != "" {
		googleAuthProvider.oauth.Endpoint.TokenURL = accessTokenUrl
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
		return nil, errors.New("callback code is empty")
	}

	googleUser, err := as.getUser(c.Arg("code"))
	if err != nil {
		return nil, err
	}

	return googleUser.ToFSUser(), nil
}

func (as *GoogleAuthProvider) VerifyIDToken(c fs.Context, t fs.IDToken) (_ *fs.User, err error) {
	if t.IDToken == "" {
		return nil, errors.New("id token is required")
	}

	payload, err := idtoken.Validate(c, t.IDToken, as.oauth.ClientID)
	if err != nil {
		c.Logger().Errorf("invalid id token: %v", err)
		return nil, errors.New("invalid id token")
	}

	// Map standard fields
	googleUser := &GoogleUser{
		ID: payload.Subject,
		// Issuer:    payload.Issuer,
		// Audience:  payload.Audience,
		// ExpiresAt: time.Unix(payload.Expires, 0),
		// IssuedAt:  time.Unix(payload.IssuedAt, 0),
	}

	if c := payload.Claims; c != nil {
		// if v, ok := c["hd"].(string); ok {
		// 	user.HD = v
		// }

		// if v, ok := c["email_verified"].(bool); ok {
		// 	user.EmailVerified = v
		// }

		// if v, ok := c["name"].(string); ok {
		// 	user.Name = v
		// }

		if v, ok := c["email"].(string); ok {
			googleUser.Email = v
		}

		if v, ok := c["given_name"].(string); ok {
			googleUser.GivenName = v
		}

		if v, ok := c["family_name"].(string); ok {
			googleUser.FamilyName = v
		}

		if v, ok := c["picture"].(string); ok {
			googleUser.Picture = v
		}
	}

	return googleUser.ToFSUser(), nil
}

func (as *GoogleAuthProvider) getUser(code string) (*GoogleUser, error) {
	token, err := as.oauth.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("google auth code exchange error: %w", err)
	}

	userResponse, err := utils.SendRequest[GoogleUser](
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
