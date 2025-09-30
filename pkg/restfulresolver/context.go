package restfulresolver

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
)

type Context struct {
	ctx         *fiber.Ctx           `json:"-"`
	fasthttpCtx *fasthttp.RequestCtx `json:"-"`

	args     map[string]string
	resource *fs.Resource
	result   *fs.Result
	entity   *entity.Entity
	logger   logger.Logger
}

func (c *Context) Result(results ...*fs.Result) *fs.Result {
	if len(results) > 0 {
		c.result = results[0]
	}

	return c.result
}

func (c *Context) SetArg(name, value string) string {
	c.args[name] = value
	return value
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

func (c *Context) Context() *fasthttp.RequestCtx {
	return c.ctx.Context()
}

func (c *Context) Deadline() (deadline time.Time, ok bool) {
	return c.fasthttpCtx.Deadline()
}

func (c *Context) Done() <-chan struct{} {
	return c.fasthttpCtx.Done()
}

func (c *Context) Err() error {
	return c.fasthttpCtx.Err()
}

func (c *Context) Value(key any) any {
	return c.fasthttpCtx.Value(key)
}

func (c *Context) IP() string {
	return c.ctx.IP()
}

func (c *Context) Body() ([]byte, error) {
	return c.ctx.Body(), nil
}

func (c *Context) Payload() (*entity.Entity, error) {
	if c.entity != nil {
		return c.entity, nil
	}

	body := c.ctx.Body()
	if len(body) == 0 {
		return nil, nil
	}

	c.entity = entity.New()
	if err := c.entity.UnmarshalJSON(body); err != nil {
		return nil, err
	}

	return c.entity, nil
}

func (c *Context) Resource() *fs.Resource {
	return c.resource
}

func (c *Context) AuthToken() string {
	// get token from header Authorization
	bearer := c.Header("Authorization")
	if len(bearer) >= 7 && bearer[:7] == "Bearer " {
		bearer = bearer[7:]
	}

	// get token from cookie
	if bearer == "" {
		bearer = c.Cookie("token")
	}

	// get token from websocket subprotocol
	if bearer == "" {
		subProtocol := c.Header("Sec-WebSocket-Protocol")
		if len(subProtocol) >= 14 && subProtocol[:14] == "Authorization," {
			bearer = strings.TrimSpace(subProtocol[14:])
		}
	}

	return bearer
}

func (c *Context) TraceID() string {
	traceID := c.ctx.Locals(fs.TraceID)
	if traceID == nil {
		return ""
	}

	return fmt.Sprintf("%v", c.ctx.Locals(fs.TraceID))
}

func (c *Context) Hostname() string {
	return c.ctx.Hostname()
}

func (c *Context) Base() string {
	return c.ctx.Protocol() + "://" + c.ctx.Hostname()
}

func (c *Context) Method() string {
	return c.ctx.Method()
}

func (c *Context) RouteName() string {
	return c.ctx.Route().Name
}

func (c *Context) OriginalURL() string {
	return c.ctx.OriginalURL()
}

func (c *Context) Path() string {
	return c.ctx.Path()
}

func (c *Context) Response() *Response {
	return &Response{c.ctx.Response()}
}

func (c *Context) Status(v int) *Context {
	c.ctx.Status(v)
	return c
}

func (c *Context) Local(key string, value ...any) (val any) {
	return c.ctx.Locals(key, value...)
}

func (c *Context) Logger() logger.Logger {
	return c.logger.WithContext(logger.LogContext{fs.TraceID: c.TraceID()}, 0)
}

func (c *Context) WSClient() fs.WSClient {
	return nil
}

func (c *Context) User() *fs.User {
	if user, ok := c.ctx.Locals("user").(*fs.User); ok {
		return user
	}

	return nil
}

func (c *Context) JSON(v any) error {
	return c.ctx.JSON(v)
}

func (c *Context) Header(key string, vals ...string) string {
	if len(vals) > 0 {
		c.ctx.Set(key, vals[0])
		return vals[0]
	}

	return c.ctx.Get(key)
}

func (c *Context) Cookie(name string, values ...*Cookie) string {
	cookieValue := c.ctx.Cookies(name)
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
		c.ctx.Cookie(&cookie)
		cookieValue = v.Value
	}

	return cookieValue
}

func (c *Context) Next() error {
	return c.ctx.Next()
}

func (c *Context) Send(data []byte) error {
	return c.ctx.Send(data)
}

func (c *Context) Redirect(path string) error {
	return c.ctx.Redirect(path)
}

func (c *Context) Bind(v any) error {
	// if there is no content type header, we assume it's JSON
	if c.ctx.Get("Content-Type") == "" {
		c.ctx.Set("Content-Type", "application/json")
		c.ctx.Request().Header.Set("Content-Type", "application/json")
	}

	return c.ctx.BodyParser(v)
}

func (c *Context) Files() ([]*fs.File, error) {
	form, err := c.ctx.MultipartForm()
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
