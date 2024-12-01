package fs

import (
	"time"

	"github.com/fastschema/fastschema/pkg/utils"
	jwt "github.com/golang-jwt/jwt/v4"
)

// User is a struct that contains user data
type User struct {
	_        any    `json:"-" fs:"label_field=username"`
	ID       uint64 `json:"id,omitempty"`
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty" fs:"optional"`
	Password string `json:"password,omitempty" fs:"optional" fs.setter:"$args.Exist && $args.Value != '' ? $hash($args.Value) : $undefined" fs.getter:"$context.Value('keeppassword') == 'true' ? $args.Value : $undefined"`

	Active           bool   `json:"active,omitempty" fs:"optional"`
	Provider         string `json:"provider,omitempty" fs:"optional"`
	ProviderID       string `json:"provider_id,omitempty" fs:"optional"`
	ProviderUsername string `json:"provider_username,omitempty" fs:"optional"`

	RoleIDs []uint64 `json:"role_ids,omitempty"`
	Roles   []*Role  `json:"roles,omitempty" fs.relation:"{'type':'m2m','schema':'role','field':'users','owner':false}"`
	Files   []*File  `json:"files,omitempty" fs.relation:"{'type':'o2m','schema':'file','field':'user','owner':true}"`

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
