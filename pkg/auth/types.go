package auth

import (
	"time"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
)

type LoginData struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token   string    `json:"token"`
	Expires time.Time `json:"expires"`
}

type Register struct {
	Username        string `json:"username"`
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

func (d *Register) Entity(activationMethod, provider string) *entity.Entity {
	return entity.New().
		Set("username", d.Username).
		Set("email", d.Email).
		Set("password", d.Password).
		Set("active", activationMethod == "auto").
		Set("provider", provider).
		Set("roles", []*entity.Entity{
			entity.New(fs.RoleUser.ID),
		})
}

type Recovery struct {
	Email string `json:"email"`
}

type Confirmation struct {
	Token string `json:"token"`
}

type ResetPassword struct {
	*Confirmation

	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type Activation struct {
	Activation string `json:"activation"` // auto, manual, email, activated
}
