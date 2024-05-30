package roleservice

import (
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	jwt "github.com/golang-jwt/jwt/v4"
)

func (rs *RoleService) ParseUser(c fs.Context) error {
	authToken := c.AuthToken()
	jwtToken, err := jwt.ParseWithClaims(
		authToken,
		&fs.UserJwtClaims{},
		func(token *jwt.Token) (any, error) {
			return []byte(rs.AppKey()), nil
		},
	)

	if err == nil {
		if claims, ok := jwtToken.Claims.(*fs.UserJwtClaims); ok && jwtToken.Valid {
			user := claims.User
			user.Roles = rs.GetRolesFromIDs(user.RoleIDs)
			c.Value("user", user)
		}
	}

	return c.Next()
}

func (rs *RoleService) Authorize(c fs.Context) error {
	resource := c.Resource()

	if resource == nil {
		return errors.NotFound("Resource not found")
	}

	resourceID := resource.ID()
	// If the resource id has prefix with "content.", for example: content.create
	// Then add the schema name to the id: content.category.create
	// Do this to clarify the content schema because the content service is dynamic.
	if strings.HasPrefix(resourceID, "content.") {
		resourceID = fmt.Sprintf("content.%s.%s", c.Arg("schema"), resourceID[8:])
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
		permission := rs.GetPermission(role.ID, resourceID)

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
