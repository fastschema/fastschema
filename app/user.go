package app

import (
	"time"

	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	jwt "github.com/golang-jwt/jwt/v4"
)

type User struct {
	ID       uint64 `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Password string `json:"password,omitempty"`

	Provider         string `json:"provider,omitempty"`
	ProviderID       string `json:"provider_id,omitempty"`
	ProviderUsername string `json:"provider_username,omitempty"`

	RoleIDs []uint64       `json:"role_ids,omitempty"`
	Roles   []*Role        `json:"roles,omitempty"`
	Active  bool           `json:"active,omitempty"`
	Entity  *schema.Entity `json:"entity,omitempty"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

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

func (u *User) JwtClaim(exp time.Time, key string, jwtHeaders ...map[string]any) (string, error) {
	u.RoleIDs = make([]uint64, 0)

	for _, role := range u.Roles {
		if role.ID > 0 {
			u.RoleIDs = append(u.RoleIDs, role.ID)
		}
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if len(jwtHeaders) > 0 {
		for k, v := range jwtHeaders[0] {
			token.Header[k] = v
		}
	}

	return token.SignedString([]byte(key))
}
