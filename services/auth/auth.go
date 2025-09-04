package authservice

import (
	"strings"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/expr"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
)

var providerNames = strings.Join(fs.AuthProviders(), ", ")
var providerArgs = fs.Args{
	"provider": fs.Arg{
		Required:    true,
		Type:        fs.TypeString,
		Description: "The auth provider name. Available providers: " + providerNames,
		Example:     "google",
	},
}

type AppLike interface {
	DB() db.Client
	Key() string
	GetAuthProvider(string) fs.AuthProvider
	Roles() []*fs.Role
}

type AuthService struct {
	DB              func() db.Client
	AppKey          func() string
	GetAuthProvider func(string) fs.AuthProvider
	Roles           func() []*fs.Role
}

func New(app AppLike) *AuthService {
	return &AuthService{
		DB:              app.DB,
		AppKey:          app.Key,
		GetAuthProvider: app.GetAuthProvider,
		Roles:           app.Roles,
	}
}

func (as *AuthService) CreateResource(api *fs.Resource, authProviders map[string]fs.AuthProvider) {
	localAuthProvider := as.
		GetAuthProvider(auth.ProviderLocal).(*auth.LocalProvider)

	authGroup := api.Group("auth").
		Add(fs.Get("me", as.Me, &fs.Meta{Public: true}))
	authGroup.
		Group(auth.ProviderLocal).
		Add(
			fs.Post("login", localAuthProvider.LocalLogin, &fs.Meta{Public: true}),
			fs.Post("register", localAuthProvider.Register, &fs.Meta{Public: true}),
			fs.Post("activate", localAuthProvider.Activate, &fs.Meta{Public: true}),
			fs.Post("activate/send", localAuthProvider.SendActivationLink, &fs.Meta{Public: true}),
			fs.Post("recover", localAuthProvider.Recover, &fs.Meta{Public: true}),
			fs.Post("recover/check", localAuthProvider.RecoverCheck, &fs.Meta{Public: true}),
			fs.Post("recover/reset", localAuthProvider.ResetPassword, &fs.Meta{Public: true}),
		)

	if len(authProviders) > 1 {
		authGroup.Group("provider", &fs.Meta{
			Prefix: "/:provider",
			Args:   providerArgs,
		}).Add(
			fs.NewResource("login", as.Login, &fs.Meta{
				Public: true,
				Get:    "/login",
			}),
			fs.NewResource("callback", as.Callback, &fs.Meta{
				Public: true,
				Get:    "/callback",
				Post:   "/callback",
			}),
			fs.NewResource("verify_idtoken", as.VerifyIDToken, &fs.Meta{
				Public: true,
				Post:   "/verify_idtoken",
			}),
		)
	}
}

func (as *AuthService) AuthUserCan(c fs.Context, user *fs.User, resourceID string) bool {
	exprConfig := expr.Config{
		DB: func() expr.DBLike {
			return as.DB()
		},
	}

	// Check for all user roles for this action.
	// If any role has permission value allow, then allow.
	for _, role := range user.Roles {
		if err := role.Check(c, exprConfig); err != nil {
			c.Logger().Error(err)
			continue
		}

		permission := as.GetPermission(role.ID, resourceID)
		if permission == nil {
			continue
		}

		if err := permission.Check(c, exprConfig); err == nil {
			return true
		} else {
			c.Logger().Error(err)
		}
	}

	return false
}

func (as *AuthService) GetRolesFromIDs(ids []uint64) []*fs.Role {
	result := []*fs.Role{}
	roles := as.Roles()

	for _, role := range roles {
		for _, id := range ids {
			if role.ID == id {
				result = append(result, role)
			}
		}
	}

	return result
}

func (as *AuthService) GetPermission(roleID uint64, resource string) *fs.Permission {
	matchedRole := &fs.Role{
		ID:          roleID,
		Permissions: []*fs.Permission{},
	}

	for _, role := range as.Roles() {
		if role.ID == roleID {
			matchedRole = role
		}
	}

	for _, permission := range matchedRole.Permissions {
		allowWildcard := strings.HasSuffix(permission.Resource, ".*") &&
			strings.HasPrefix(resource, permission.Resource[:len(permission.Resource)-2])
		allowExact := permission.Resource == resource

		if allowWildcard || allowExact {
			return permission
		}
	}

	return nil
}
