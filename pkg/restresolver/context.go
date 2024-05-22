package restresolver

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/schema"

	// "github.com/fastschema/fastschema/app/server"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

const requestID = "request_id"

type RequestIDContextKey string

func (c RequestIDContextKey) String() string {
	return string(c)
}

var (
	ContextKeyRequestID = RequestIDContextKey(requestID)
)

type Context struct {
	*fiber.Ctx
	args     map[string]string
	resource *fs.Resource
	result   *fs.Result
	entity   *schema.Entity
	logger   logger.Logger
}

func (c *Context) Result(results ...*fs.Result) *fs.Result {
	if len(results) > 0 {
		c.result = results[0]
	}

	return c.result
}

func (c *Context) Args() map[string]string {
	return c.args
}

func (c *Context) Arg(name string, defaults ...string) string {
	v, ok := c.args[name]
	if !ok && len(defaults) > 0 {
		return defaults[0]
	}

	return v
}

func (c *Context) ArgInt(name string, defaults ...int) int {
	v, err := strconv.Atoi(c.Arg(name))
	if err != nil && len(defaults) > 0 {
		return defaults[0]
	}

	return v
}

func (c *Context) Entity() (*schema.Entity, error) {
	if c.entity != nil {
		return c.entity, nil
	}

	c.entity = schema.NewEntity()
	if err := c.entity.UnmarshalJSON(c.Ctx.Body()); err != nil {
		return nil, err
	}

	return c.entity, nil
}

func (c *Context) Resource() *fs.Resource {
	return c.resource
}

func (c *Context) AuthToken() string {
	bearer := c.Header("Authorization")

	if len(bearer) >= 7 && bearer[:7] == "Bearer " {
		bearer = bearer[7:]
	}

	if bearer == "" {
		bearer = c.Cookie("token")
	}

	return bearer
}

func (c *Context) ID() string {
	return fmt.Sprintf("%v", c.Locals(requestID))
}

func (c *Context) Hostname() string {
	return c.Ctx.Hostname()
}

func (c *Context) Base() string {
	return c.Ctx.Protocol() + "://" + c.Ctx.Hostname()
}

func (c *Context) Method() string {
	return c.Ctx.Method()
}

func (c *Context) RouteName() string {
	return c.Ctx.Route().Name
}

func (c *Context) OriginalURL() string {
	return c.Ctx.OriginalURL()
}

func (c *Context) Path() string {
	return c.Ctx.Path()
}

func (c *Context) Response() *Response {
	return &Response{c.Ctx.Response()}
}

func (c *Context) Context() context.Context {
	return context.WithValue(c.Ctx.Context(), ContextKeyRequestID, c.ID())
}

func (c *Context) Status(v int) *Context {
	c.Ctx.Status(v)
	return c
}

func (c *Context) Value(key string, value ...any) (val any) {
	return c.Ctx.Locals(key, value...)
}

func (c *Context) Logger() logger.Logger {
	return c.logger.WithContext(logger.LogContext{requestID: c.ID()}, 0)
}

func (c *Context) User() *fs.User {
	if user, ok := c.Locals("user").(*fs.User); ok {
		return user
	}

	return nil
}

func (c *Context) JSON(v any) error {
	return c.Ctx.JSON(v)
}

func (c *Context) Header(key string, vals ...string) string {
	if len(vals) > 0 {
		c.Ctx.Set(key, vals[0])
		return vals[0]
	}

	return c.Ctx.Get(key)
}

func (c *Context) Cookie(name string, values ...*Cookie) string {
	cookieValue := c.Ctx.Cookies(name)
	if len(values) > 0 {
		v := values[0]
		cookie := fiber.Cookie{
			Name:     name,
			Value:    v.Value,
			Path:     v.Path,
			Domain:   v.Domain,
			Expires:  v.Expires,
			Secure:   v.Secure,
			HTTPOnly: v.HTTPOnly,
			SameSite: v.SameSite,
		}
		c.Ctx.Cookie(&cookie)
		cookieValue = v.Value
	}

	return cookieValue
}

func (c *Context) Next() error {
	return c.Ctx.Next()
}

func (c *Context) Send(data []byte) error {
	return c.Ctx.Send(data)
}

func (c *Context) Redirect(path string) error {
	return c.Ctx.Redirect(path)
}

func (c *Context) Parse(v any) error {
	// if there is no content type header, we assume it's JSON
	if c.Ctx.Get("Content-Type") == "" {
		c.Ctx.Set("Content-Type", "application/json")
		c.Ctx.Request().Header.Set("Content-Type", "application/json")
	}

	return c.Ctx.BodyParser(v)
}

func (c *Context) Files() ([]*fs.File, error) {
	form, err := c.MultipartForm()
	if err != nil {
		return nil, err
	}

	files := make([]*fs.File, 0)

	for _, fileHeaders := range form.File {
		for _, fileHeader := range fileHeaders {
			f, err := fileHeader.Open()
			if err != nil {
				return nil, err
			}

			m := make([]byte, 512)

			if _, err := f.Read(m); err != nil {
				return nil, err
			}

			if _, err := f.Seek(0, 0); err != nil {
				return nil, err
			}

			files = append(files, &fs.File{
				Name:   fileHeader.Filename,
				Size:   uint64(fileHeader.Size),
				Type:   http.DetectContentType(m),
				Reader: f,
			})
		}
	}

	return files, nil
}

type Response struct {
	*fasthttp.Response
}

func (r *Response) Header(key string, vals ...string) string {
	if len(vals) > 0 {
		r.Response.Header.Add(key, vals[0])
		return vals[0]
	}

	return string(r.Response.Header.Peek(key))
}
