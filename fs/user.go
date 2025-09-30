package fs

import (
	"time"

	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	jwt "github.com/golang-jwt/jwt/v4"
)

// User is a struct that contains user data
type User struct {
	_         any    `json:"-" fs:"label_field=username"`
	ID        uint64 `json:"id,omitempty"`
	Username  string `json:"username,omitempty" fs:"optional"`
	Email     string `json:"email,omitempty" fs:"optional"`
	FirstName string `json:"first_name,omitempty" fs:"optional"`
	LastName  string `json:"last_name,omitempty" fs:"optional"`
	Password  string `json:"password,omitempty" fs:"optional" fs.setter:"$args.Exist && $args.Value != '' ? $hash($args.Value) : $undefined" fs.getter:"$context.Value('keeppassword') == 'true' ? $args.Value : $undefined"`

	Active               bool   `json:"active,omitempty" fs:"optional"`
	Provider             string `json:"provider,omitempty" fs:"optional"`
	ProviderID           string `json:"provider_id,omitempty" fs:"optional"`
	ProviderUsername     string `json:"provider_username,omitempty" fs:"optional"`
	ProviderProfileImage string `json:"provider_profile_image,omitempty" fs:"optional"`

	RoleIDs []uint64 `json:"role_ids,omitempty"`
	Roles   []*Role  `json:"roles,omitempty" fs.relation:"{'type':'m2m','schema':'role','field':'users','owner':false}"`
	Files   []*File  `json:"files,omitempty" fs.relation:"{'type':'o2m','schema':'file','field':'user','owner':true}"`

	CreatedAt *time.Time `json:"created_at,omitempty"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

func (u User) Schema() *schema.Schema {
	return &schema.Schema{
		Fields: []*schema.Field{},
		DB: &schema.SchemaDB{
			Indexes: []*schema.SchemaDBIndex{
				// unique index on provider + provider_id
				{
					Name:    "idx_user_provider_provider_id",
					Unique:  true,
					Columns: []string{"provider", "provider_id"},
				},
				// unique index on username
				{
					Name:    "idx_user_username",
					Unique:  true,
					Columns: []string{"username"},
				},
				// unique index on email
				{
					Name:    "idx_user_email",
					Unique:  true,
					Columns: []string{"email"},
				},
			},
		},
	}
}

// UserJwtClaims is a struct that contains the user jwt claims
type UserJwtClaims struct {
	jwt.RegisteredClaims

	User *User `json:"user"`
}

type JwtCustomClaimsFunc func(Context, *UserJwtClaims) (jwt.Claims, error)

type UserJwtConfig struct {
	Key              string
	ExpiresAt        time.Time
	CustomClaimsFunc JwtCustomClaimsFunc
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
func (u *User) JwtClaim(c Context, config *UserJwtConfig) (string, time.Time, error) {
	if config == nil || config.Key == "" {
		return "", time.Time{}, errors.InternalServerError("jwt: missing secret key")
	}

	u.RoleIDs = make([]uint64, 0)

	for _, role := range u.Roles {
		if role.ID > 0 {
			u.RoleIDs = append(u.RoleIDs, role.ID)
		}
	}

	exp := config.ExpiresAt
	if config.ExpiresAt.IsZero() {
		exp = time.Now().Add(time.Hour * 24 * 30)
	}

	jwtClaims := &UserJwtClaims{
		User: &User{
			ID:                   u.ID,
			Provider:             u.Provider,
			ProviderProfileImage: u.ProviderProfileImage,
			Username:             u.Username,
			FirstName:            u.FirstName,
			LastName:             u.LastName,
			Email:                u.Email,
			Active:               u.Active,
			RoleIDs:              u.RoleIDs,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    utils.Env("APP_NAME"),
			ExpiresAt: &jwt.NumericDate{Time: exp},
		},
	}

	var claims jwt.Claims = jwtClaims
	if config.CustomClaimsFunc != nil {
		customClaims, err := config.CustomClaimsFunc(c, jwtClaims)
		if err != nil {
			return "", time.Time{}, err
		}

		if customClaims != nil {
			claims = customClaims
		}
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedString, err := jwtToken.SignedString([]byte(config.Key))

	return signedString, exp, err
}
