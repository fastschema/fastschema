package userservice

import (
	"time"

	"github.com/fastschema/fastschema/app"
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
	DB() app.DBClient
	Key() string
}

type UserService struct {
	DB     func() app.DBClient
	AppKey func() string
}

func New(app AppLike) *UserService {
	return &UserService{
		DB:     app.DB,
		AppKey: app.Key,
	}
}

func (u *UserService) Login(c app.Context, loginData *LoginData) (*LoginResponse, error) {
	userModel, err := u.DB().Model("user")
	if err != nil {
		return nil, err
	}

	userEntity, err := userModel.Query(app.Or(
		app.EQ("username", loginData.Login),
		app.EQ("email", loginData.Login),
	)).Select(
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
	).First()
	if err != nil && !app.IsNotFound(err) {
		return nil, err
	}

	if userEntity == nil {
		return nil, errors.Unauthorized("invalid login or password")
	}

	if err := utils.CheckHash(loginData.Password, userEntity.GetString("password", "")); err != nil {
		return nil, errors.Unauthorized("invalid login or password")
	}

	user := app.EntityToUser(userEntity)
	if !user.Active {
		return nil, errors.Unauthorized("user is not active")
	}

	jwtToken, exp, err := user.JwtClaim(u.AppKey())
	if err != nil {
		return nil, err
	}

	return &LoginResponse{Token: jwtToken, Expires: exp}, nil
}

func (u *UserService) Logout(c app.Context, _ *any) (*any, error) {
	return nil, nil
}

func (u *UserService) Me(c app.Context, _ *any) (*app.User, error) {
	user := c.User()

	if user == nil {
		return nil, errors.Unauthorized()
	}

	return user, nil
}
