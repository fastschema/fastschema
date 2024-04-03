package restresolver

import (
	"time"

	"github.com/google/uuid"
)

func MiddlewareCookie(c *Context) error {
	if c.Cookie("UUID") == "" {
		exp := time.Now().Add(time.Hour * 100 * 365 * 24)
		c.Cookie("UUID", &Cookie{
			Name:     "UUID",
			Value:    uuid.NewString(),
			Expires:  exp,
			HTTPOnly: false,
			SameSite: "lax",
			Secure:   true,
		})
	}
	return c.Next()
}
