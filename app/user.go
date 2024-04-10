package app

import (
	"time"

	"github.com/fastschema/fastschema/pkg/utils"
	jwt "github.com/golang-jwt/jwt/v4"
)

// User is a struct that contains user data
type User struct {
	ID       uint64 `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`

	Provider         string `json:"provider,omitempty"`
	ProviderID       string `json:"provider_id,omitempty"`
	ProviderUsername string `json:"provider_username,omitempty"`

	RoleIDs []uint64 `json:"role_ids,omitempty"`
	Roles   []*Role  `json:"roles,omitempty"`
	Active  bool     `json:"active,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// UserJwtClaims is a struct that contains the user jwt claims
type UserJwtClaims struct {
	jwt.RegisteredClaims
	User *User `json:"user"`
}

func (u *User) IsRoot() bool {
	if u == nil {
		return false
	}

	if u.Roles == nil {
		return false
	}

	for _, role := range u.Roles {
		if role.Root {
			return true
		}
	}

	return false
}

// JwtClaim generates a jwt claim
func (u *User) JwtClaim(key string, exps ...time.Time) (string, time.Time, error) {
	u.RoleIDs = make([]uint64, 0)

	for _, role := range u.Roles {
		if role.ID > 0 {
			u.RoleIDs = append(u.RoleIDs, role.ID)
		}
	}

	exp := time.Now().Add(time.Hour * 24 * 30)
	if len(exps) > 0 {
		exp = exps[0]
	}

	claims := &UserJwtClaims{
		User: &User{
			ID:       u.ID,
			Provider: u.Provider,
			Username: u.Username,
			Email:    u.Email,
			Active:   u.Active,
			RoleIDs:  u.RoleIDs,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    utils.Env("APP_NAME"),
			ExpiresAt: &jwt.NumericDate{Time: exp},
		},
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := jwtToken.SignedString([]byte(key))

	return signedString, exp, err
}
