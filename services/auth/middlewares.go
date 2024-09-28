package authservice

import (
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/golang-jwt/jwt/v4"
)

func (as *AuthService) ParseUser(c fs.Context) (err error) {
	authToken := c.AuthToken()
	jwtToken, err := jwt.ParseWithClaims(
		authToken,
		&fs.UserJwtClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(as.AppKey()), nil
		},
	)

	if err == nil {
		if claims, ok := jwtToken.Claims.(*fs.UserJwtClaims); ok && jwtToken.Valid {
			user := claims.User
			if user.Roles, err = as.GetRolesFromIDs(user.RoleIDs); err != nil {
				c.Logger().Errorf("Cannot get user roles: %v", err)
				return errors.InternalServerError("Cannot get user roles")
			}
			c.Value("user", user)
		}
	}

	return c.Next()
}

func (as *AuthService) Authorize(c fs.Context) error {
	roles, err := as.Roles()
	if err != nil {
		c.Logger().Errorf("Cannot get system roles: %v", err)
		return errors.InternalServerError("Cannot get system roles")
	}

	resource := c.Resource()
	if resource == nil {
		return errors.NotFound("Resource not found")
	}

	resourceID := resource.ID()

	// If the resource id has prefix with "content.", for example: content.create
	// Then add the schema name to the id: content.category.create
	// Do this to clarify the content schema because the content service is dynamic.
	if strings.HasPrefix(resourceID, "api.content.") {
		resourceID = fmt.Sprintf("api.content.%s.%s", c.Arg("schema"), resourceID[12:])
	}

	if resourceID == "api.realtime.content" {
		resourceID = fmt.Sprintf("api.realtime.content.%s.%s", c.Arg("schema"), c.Arg("event", "*"))
	}

	user := c.User()
	if user == nil {
		user = fs.GuestUser
	}

	// Allow root user to access all routes.
	if user.IsRoot() {
		return nil
	}

	// Allow white listed routes.
	if resource.IsPublic() {
		return nil
	}

	// Disallow inactive users.
	if user.ID > 0 && !user.Active {
		return errors.Forbidden("User is inactive")
	}

	// Check for all user roles for this action.
	// If any role has permission value allow, then allow.
	for _, role := range user.Roles {
		permission := as.GetPermission(roles, role.ID, resourceID)

		// if permission value is allow, then allow
		if permission.Value == fs.PermissionTypeAllow.String() {
			return nil
		}
	}

	return utils.If(
		user.ID > 0,
		errors.Forbidden("Forbidden"),
		errors.Unauthorized("Unauthorized"),
	)
}
