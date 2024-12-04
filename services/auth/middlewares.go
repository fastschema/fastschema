package authservice

import (
	"fmt"
	"strings"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/errors"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/golang-jwt/jwt/v4"
)

func (as *AuthService) ParseUser(c fs.Context) error {
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
			user.Roles = as.GetRolesFromIDs(user.RoleIDs)
			c.Local("user", user)
		}
	}

	return c.Next()
}

func (as *AuthService) Authorize(c fs.Context) error {
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

	// If the resource id is "api.realtime.content"
	// Then add the schema name and event name to the id: api.realtime.content.category.create
	if resourceID == "api.realtime.content" {
		resourceID = fmt.Sprintf("api.realtime.content.%s.%s", c.Arg("schema"), c.Arg("event", "*"))
	}

	user := c.User()
	if user == nil {
		user = &fs.User{
			ID:       0,
			Username: "",
			Roles:    as.GetRolesFromIDs([]uint64{fs.RoleGuest.ID}),
		}
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

	if as.AuthUserCan(c, user, resourceID) {
		return nil
	}

	return utils.If(
		user.ID > 0,
		errors.Forbidden("Forbidden"),
		errors.Unauthorized("Unauthorized"),
	)
}
