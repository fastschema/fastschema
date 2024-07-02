package restfulresolver_test

import (
	"fmt"
	"net"
	"net/http/httptest"
	"testing"

	fhws "github.com/fasthttp/websocket"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/gofiber/contrib/websocket"
	"github.com/stretchr/testify/assert"
)

func TestNewWSClient(t *testing.T) {
	conn := &websocket.Conn{}
	client := restfulresolver.NewWSClient(conn)
	assert.NotNil(t, client)
}

var adminUser = &fs.User{
	ID:       1,
	Username: "adminuser",
	Active:   true,
	Roles:    []*fs.Role{fs.RoleAdmin},
	RoleIDs:  []uint64{1},
}

func createTestApp(t *testing.T, wsResources []*fs.Resource, getHooks func() *fs.Hooks) *restfulresolver.Server {
	server := restfulresolver.New(restfulresolver.Config{
		Logger: logger.CreateMockLogger(),
	})

	server.Use(
		func(c *restfulresolver.Context) error {
			if c.Arg("nouser") == "" {
				c.Locals("user", adminUser)
			}
			return c.Next()
		},
	)

	api := server.Group("api", nil)
	restfulresolver.RegisterResourceRoutes(wsResources, api, getHooks)

	go func() {
		assert.NoError(t, server.Listen(":55555"))
	}()

	readyCh := make(chan struct{})

	go func() {
		for {
			conn, err := net.Dial("tcp", "localhost:55555")
			if err != nil {
				continue
			}

			if conn != nil {
				readyCh <- struct{}{}
				conn.Close()
				break
			}
		}
	}()

	<-readyCh

	return server
}

func TestWSClient(t *testing.T) {
	wsResources := []*fs.Resource{
		fs.WS("/realtime", func(c fs.Context, _ any) (any, error) {
			assert.NotNil(t, c.WSClient())

			if c.Arg("nouser") == "" {
				assert.Equal(t, c.User().Username, "adminuser")
			} else {
				assert.Nil(t, c.User())
			}

			assert.NotNil(t, c.Logger())

			return nil, nil
		}),
		fs.WS("/realtime/close", func(c fs.Context, _ any) (any, error) {
			client := c.WSClient()
			assert.True(t, client.IsCloseNormal(&fhws.CloseError{Code: 1000, Text: "test close"}))
			assert.True(t, client.IsCloseNormal(&fhws.CloseError{Code: 1001, Text: "test close"}))
			assert.False(t, client.IsCloseNormal(&fhws.CloseError{Code: 1006, Text: "test close"}))

			err := c.WSClient().Close("test close")
			assert.NoError(t, err)

			assert.Error(t, client.Close("test close"))

			return nil, nil
		}),
		fs.WS("/realtime/message/:name", func(c fs.Context, _ any) (any, error) {
			client := c.WSClient()
			argParam := c.Arg("name")
			argQuery := c.Arg("message")
			argDefault := c.Arg("default", "hi")
			argNonExist := c.Arg("nonexist")
			argInt1 := c.ArgInt("num1")
			argInt2 := c.ArgInt("num2")
			argInt3 := c.ArgInt("num3", 10)
			argInt := fmt.Sprintf("%d-%d-%d", argInt1, argInt2, argInt3)

			assert.NotNil(t, c.ID())

			for {
				_, msg, err := client.Read()
				if err != nil {
					break
				}

				msg = append(msg, []byte(" "+argParam+" "+argQuery+" "+argDefault+argNonExist+" "+argInt)...)
				if err = client.Write(msg, fs.WSMessageText); err != nil {
					break
				}
			}

			return nil, nil
		}),
	}

	server := createTestApp(t, wsResources, nil)
	defer func() {
		assert.NoError(t, server.Shutdown())
	}()

	req := httptest.NewRequest("GET", "http://localhost:55555/api/realtime", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Contains(t, resp.Status, "Upgrade Required")

	header := map[string][]string{}

	// Connect to websocket
	conn, resp, err := fhws.DefaultDialer.Dial("ws://localhost:55555/api/realtime", header)
	defer resp.Body.Close()
	defer conn.Close()
	assert.NoError(t, err)
	assert.Equal(t, 101, resp.StatusCode)
	assert.Equal(t, "websocket", resp.Header.Get("Upgrade"))

	// Context has no user
	conn, resp, err = fhws.DefaultDialer.Dial("ws://localhost:55555/api/realtime?nouser=true", header)
	defer resp.Body.Close()
	defer conn.Close()
	assert.NoError(t, err)

	// Server send close message
	conn, resp, err = fhws.DefaultDialer.Dial("ws://localhost:55555/api/realtime/close", header)
	defer resp.Body.Close()
	defer conn.Close()
	assert.NoError(t, err)
	assert.Equal(t, 101, resp.StatusCode)
	assert.Equal(t, "websocket", resp.Header.Get("Upgrade"))

	// Server send message
	wsURL := "ws://localhost:55555/api/realtime/message/john?message=hello&num1=5&num2=invalid"
	conn, resp, err = fhws.DefaultDialer.Dial(wsURL, header)
	defer resp.Body.Close()
	defer conn.Close()
	assert.NoError(t, err)
	assert.NoError(t, conn.WriteMessage(fhws.TextMessage, []byte("test message")))
	_, p, err := conn.ReadMessage()
	assert.NoError(t, err)
	assert.Equal(t, "test message john hello hi 5-0-10", string(p))

	// Client close
	conn, resp, err = fhws.DefaultDialer.Dial("ws://localhost:55555/api/realtime/close", header)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.NoError(t, conn.Close())
}

func TestWSHooksError(t *testing.T) {
	wsResources := []*fs.Resource{
		fs.WS("/realtime", func(c fs.Context, _ any) (any, error) {
			return nil, nil
		}),
	}

	getHooks := func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{
				func(c fs.Context) error {
					return fmt.Errorf("test error")
				},
			},
		}
	}

	server := createTestApp(t, wsResources, getHooks)
	defer func() {
		assert.NoError(t, server.Shutdown())
	}()

	conn, resp, err := fhws.DefaultDialer.Dial("ws://localhost:55555/api/realtime", nil)
	defer resp.Body.Close()
	defer conn.Close()
	assert.NoError(t, err)

	for {
		_, _, err := conn.ReadMessage()
		assert.Contains(t, err.Error(), "test error")
		break
	}
}

func TestWSHooksSuccess(t *testing.T) {
	wsResources := []*fs.Resource{
		fs.WS("/realtime", func(c fs.Context, _ any) (any, error) {
			assert.NotNil(t, c.User())
			return nil, nil
		}),
	}

	getHooks := func() *fs.Hooks {
		return &fs.Hooks{
			PreResolve: []fs.Middleware{
				func(c fs.Context) error {
					c.Value("user", adminUser)
					return nil
				},
			},
		}
	}

	server := createTestApp(t, wsResources, getHooks)
	defer func() {
		assert.NoError(t, server.Shutdown())
	}()

	conn, resp, err := fhws.DefaultDialer.Dial("ws://localhost:55555/api/realtime", nil)
	defer resp.Body.Close()
	defer conn.Close()
	assert.NoError(t, err)
}
