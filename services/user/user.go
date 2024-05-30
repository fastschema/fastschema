package userservice

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

type AppLike interface {
	DB() db.Client
	Key() string
}

type UserService struct {
	DB     func() db.Client
	AppKey func() string
}

func New(app AppLike) *UserService {
	return &UserService{
		DB:     app.DB,
		AppKey: app.Key,
	}
}

func (u *UserService) Login(c fs.Context, loginData *LoginData) (*LoginResponse, error) {
	user, err := db.Query[*fs.User](u.DB()).
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
		First(c.Context())
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

	jwtToken, exp, err := user.JwtClaim(u.AppKey())
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: jwtToken, Expires: exp}, nil
}

func (u *UserService) Logout(c fs.Context, _ any) (*any, error) {
	return nil, nil
}

func (u *UserService) Me(c fs.Context, _ any) (*fs.User, error) {
	user := c.User()

	if user == nil {
		return nil, errors.Unauthorized()
	}

	return user, nil
}
