package restfulresolver

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	fhws "github.com/fasthttp/websocket"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

var closeNormals = []fs.WSCloseType{
	fs.WSCloseNormalClosure,
	fs.WSCloseGoingAway,
}

type WSClient struct {
	mu   sync.RWMutex
	conn *websocket.Conn
	id   string
}

func NewWSClient(conn *websocket.Conn) *WSClient {
	return &WSClient{
		conn: conn,
		id:   utils.RandomString(16),
	}
}

// ID returns the ID of the connection.
func (c *WSClient) ID() string {
	return c.id
}

// Close closes the connection without any error.
func (c *WSClient) Close(msgs ...string) error {
	return c.CloseWithCode(fs.WSCloseNormalClosure, append(msgs, "")[0])
}

// CloseWithCode closes the connection with the given code and message.
func (c *WSClient) CloseWithCode(code fs.WSCloseType, msg string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if e := c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code.Int(), msg),
	); e != nil {
		return fmt.Errorf("wsclient: failed to write close message: %w", e)
	}

	if closeError := c.conn.Close(); closeError != nil {
		return fmt.Errorf("wsclient: failed to close conn: %w", closeError)
	}

	return nil
}

// Read reads the next message from the WebSocket connection.
func (c *WSClient) Read() (messageType fs.WSMessageType, p []byte, err error) {
	mt, p, err := c.conn.ReadMessage()
	return fs.WSMessageType(mt), p, err
}

// Write writes the given data to the WebSocket connection.
// If no message types are provided, it defaults to TextMessage.
func (c *WSClient) Write(data []byte, messageTypes ...fs.WSMessageType) error {
	messageTypes = append(messageTypes, fs.WSMessageText)
	return c.conn.WriteMessage(messageTypes[0].Int(), data)
}

// IsCloseNormal checks if the close error is a normal closure.
func (c *WSClient) IsCloseNormal(err error) bool {
	var wce *fhws.CloseError
	if errors.As(err, &wce) && utils.Contains(closeNormals, fs.WSCloseTypeFromInt(wce.Code)) {
		return true
	}

	return false
}

type BaseContext = Context

type WSContext struct {
	*BaseContext

	wsClient *WSClient
}

func CreateWSContext(r *fs.Resource, c *fiber.Ctx, logger logger.Logger, wsClient *WSClient) *WSContext {
	ctx := &WSContext{
		BaseContext: &Context{
			Ctx:      c,
			args:     make(map[string]string),
			resource: r,
			logger:   logger,
		},
		wsClient: wsClient,
	}

	return ctx
}

func (c *WSContext) Arg(name string, defaults ...string) string {
	pv := c.wsClient.conn.Params(name)
	if pv != "" {
		return pv
	}

	qv := c.wsClient.conn.Query(name, "")
	if qv != "" {
		return qv
	}

	if len(defaults) > 0 {
		return defaults[0]
	}

	return ""
}

func (c *WSContext) ArgInt(name string, defaults ...int) int {
	v, err := strconv.Atoi(c.Arg(name))
	if err != nil && len(defaults) > 0 {
		return defaults[0]
	}

	return v
}

func (c *WSContext) Locals(key string, defaults ...any) any {
	return c.wsClient.conn.Locals(key, defaults...)
}

func (c *WSContext) Local(key string, value ...any) (val any) {
	return c.Locals(key, value...)
}

func (c *WSContext) TraceID() string {
	return fmt.Sprintf("%v", c.Locals(fs.TraceID))
}

func (c *WSContext) Logger() logger.Logger {
	return c.logger.WithContext(logger.LogContext{fs.TraceID: c.TraceID()}, 1)
}

func (c *WSContext) User() *fs.User {
	if user, ok := c.Locals("user").(*fs.User); ok {
		return user
	}

	return nil
}

func (c *WSContext) WSClient() fs.WSClient {
	return c.wsClient
}

func WSResourceHandler(r *fs.Resource, hooks *fs.Hooks, router *Router) {
	path := r.Meta().WS
	router.fiberGroup.Use(path, func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	router.fiberGroup.Get(path, func(ctx *fiber.Ctx) error {
		handler := websocket.New(func(conn *websocket.Conn) {
			client := NewWSClient(conn)
			c := CreateWSContext(r, ctx, router.logger, client)

			if hooks != nil {
				for _, hook := range hooks.PreResolve {
					if err := hook(c); err != nil {
						if closeErr := client.Close(err.Error()); closeErr != nil {
							router.logger.Error(closeErr)
						}
						return
					}
				}
			}

			if _, err := r.Handler()(c); err != nil {
				router.logger.Error(err)
			}
		}, websocket.Config{
			Subprotocols: []string{"Authorization"},
		})

		return handler(ctx)
	})
}
