package auth

import (
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
)

type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type LocalAuthProvider struct {
	db     func() db.Client
	appKey func() string
}

func NewLocalAuthProvider(config Config, redirectURL string) (fs.AuthProvider, error) {
	return &LocalAuthProvider{}, nil
}

func (la *LocalAuthProvider) Init(db func() db.Client, appKey func() string) {
	la.db = db
	la.appKey = appKey
}

func (la *LocalAuthProvider) Name() string {
	return "local"
}

func (la *LocalAuthProvider) Login(c fs.Context) (_ any, err error) {
	loginEntity, err := c.Payload()
	if err != nil {
		return nil, err
	}

	loginData, err := schema.BindEntity[*LoginData](loginEntity)
	if err != nil {
		return nil, err
	}

	if loginData == nil || loginData.Login == "" || loginData.Password == "" {
		return nil, errors.BadRequest("login and password are required")
	}

	user, err := db.Builder[*fs.User](la.db()).
		Where(db.Or(
			db.EQ("username", loginData.Login),
			db.EQ("email", loginData.Login),
		)).
		Select(
			"id",
			"username",
			"email",
			"password",
			"provider",
			"provider_id",
			"provider_username",
			"active",
			"roles",
			schema.FieldCreatedAt,
			schema.FieldUpdatedAt,
			schema.FieldDeletedAt,
		).
		First(c)
		// First(c.Context())
	if err != nil && !db.IsNotFound(err) {
		return nil, err
	}

	if user == nil {
		return nil, errors.Unauthorized("invalid login or password")
	}

	if err := utils.CheckHash(loginData.Password, user.Password); err != nil {
		return nil, errors.Unauthorized("invalid login or password")
	}

	if !user.Active {
		return nil, errors.Unauthorized("user is not active")
	}

	jwtToken, exp, err := user.JwtClaim(la.appKey())
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: jwtToken, Expires: exp}, nil
}

func (la *LocalAuthProvider) Callback(c fs.Context) (_ *fs.User, err error) {
	return nil, nil
}
