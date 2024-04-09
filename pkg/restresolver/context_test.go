package restresolver_test

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/fastschema/fastschema/app"
	"github.com/fastschema/fastschema/pkg/restresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDContextKeyString(t *testing.T) {
	key := restresolver.RequestIDContextKey("test_key")
	str := key.String()

	assert.Equal(t, "test_key", str)
}

func TestContextResult(t *testing.T) {
	c := &restresolver.Context{}
	result := &app.Result{}
	c.Result(result)
	assert.Equal(t, result, c.Result())
}

func TestContextArgs(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Get("/test/:param", func(c *restresolver.Context) error {
		args := c.Args()
		assert.Equal(t, map[string]string{"param": "param", "param1": "value1", "param2": "5"}, args)
		assert.Equal(t, "value1", c.Arg("param1"))
		assert.Equal(t, "default_value", c.Arg("param3", "default_value"))
		assert.Equal(t, 5, c.ArgInt("param2"))
		assert.Equal(t, 15, c.ArgInt("param4", 15))
		return c.JSON(args)
	})

	req := httptest.NewRequest("GET", "/test/param?param1=value1&param2=5", nil)
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, `{"param":"param","param1":"value1","param2":"5"}`, utils.Must(utils.ReadCloserToString(resp.Body)))
}

func TestContextEntity(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Post("/test", func(c *restresolver.Context) error {
		entity, err := c.Entity()
		assert.NoError(t, err)
		assert.Equal(t, "value", entity.Get("key"))
		entity2, _ := c.Entity()
		assert.Equal(t, entity, entity2)
		return c.JSON(entity)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"key": "value"}`)))
	defer req.Body.Close()
	resp, err := server.Test(req)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestContextEntityError(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Post("/test", func(c *restresolver.Context) error {
		_, err := c.Entity()
		assert.Error(t, err)
		return err
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"key": "value"`)))
	defer req.Body.Close()
	resp, err := server.Test(req)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 500, resp.StatusCode)
}

func TestContextParse(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Post("/test", func(c *restresolver.Context) error {
		entity := map[string]any{}
		err := c.Parse(&entity)
		assert.NoError(t, err)
		assert.Equal(t, "value", entity["key"])
		return c.JSON(entity)
	})

	req := httptest.NewRequest("POST", "/test", bytes.NewReader([]byte(`{"key": "value"}`)))
	defer req.Body.Close()
	resp, err := server.Test(req)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestContextResource(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	resource := app.NewResource("test", func(c app.Context, _ *any) (*any, error) {
		return nil, nil
	})
	server.Get("/test", func(c *restresolver.Context) error {
		assert.NotNil(t, c.Resource())
		return nil
	}, resource)

	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := server.Test(req)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestContextAuthToken(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Get("/test", func(c *restresolver.Context) error {
		assert.Equal(t, "token", c.AuthToken())
		return c.Redirect("/redirect")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer token")
	resp, err := server.Test(req)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("Cookie", "token=token")
	resp, err = server.Test(req2)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 302, resp.StatusCode)
}

func TestContextMethods(t *testing.T) {
	resource := app.NewResource("test_resource", func(c app.Context, _ *any) (*any, error) {
		return nil, nil
	})
	server := restresolver.New(restresolver.Config{
		Logger: app.CreateMockLogger(true),
	})
	server.Get("/test", func(c *restresolver.Context) error {
		c.Value("test", "test_value")
		assert.Equal(t, "test_value", c.Value("test"))
		assert.Nil(t, c.User())

		c.Value("user", &app.User{})
		assert.NotNil(t, c.User())

		assert.Equal(t, "header-value", c.Header("custom-header"))

		assert.NotNil(t, c.ID())
		assert.NotNil(t, c.Response())
		assert.NotNil(t, c.Logger())
		assert.Equal(t, "GET", c.Method())
		assert.Equal(t, "example.com", c.Hostname())
		assert.Equal(t, "http://example.com", c.Base())
		assert.Equal(t, "/test", c.OriginalURL())
		assert.Equal(t, "/test", c.Path())
		assert.NotNil(t, "test_resource", c.RouteName())
		assert.NotNil(t, c.Context())
		c.Status(201)
		c.Header("response-header", "response-header-value")
		c.Cookie("testcookiename", &restresolver.Cookie{
			Name:  "testcookiename",
			Value: "testcookievalue",
		})

		return c.Send([]byte("send"))
	}, resource)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "cookie=cookie")
	req.URL.RawQuery = "query=query"
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Cookie", "cookie=cookie")
	req.Header.Set("custom-header", "header-value")
	req.URL.RawQuery = "query=query"
	resp, err := server.Test(req)
	defer closeResponse(t, resp)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
	cookies := resp.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "testcookiename", cookies[0].Name)
	assert.Equal(t, "testcookievalue", cookies[0].Value)
	assert.Equal(t, "response-header-value", resp.Header.Get("response-header"))
	assert.Equal(t, "send", utils.Must(utils.ReadCloserToString(resp.Body)))
}

func createTestImage(t *testing.T) string {
	tmpFilePath := t.TempDir() + "/image.png"
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	c := color.RGBA{255, 255, 255, 255}
	for y := 0; y < img.Bounds().Dy(); y++ {
		for x := 0; x < img.Bounds().Dx(); x++ {
			img.Set(x, y, c)
		}
	}

	f, err := os.Create(tmpFilePath)
	assert.NoError(t, err)
	defer f.Close()

	assert.NoError(t, png.Encode(f, img))
	return tmpFilePath
}

func TestContextFiles(t *testing.T) {
	server := restresolver.New(restresolver.Config{})
	server.Use(func(c *restresolver.Context) error {
		assert.NotNil(t, c.Response().Header("Content-Type"))
		c.Response().Header("custom-header", "custom-header-value")
		assert.NoError(t, c.Next())
		return nil
	})
	server.Post("/test", func(c *restresolver.Context) error {
		files, err := c.Files()

		assert.NoError(t, err)
		assert.Len(t, files, 1)
		assert.Equal(t, "image.png", files[0].Name)
		assert.Equal(t, "image/png", files[0].Type)
		return c.JSON(files)
	})

	filePath := createTestImage(t)
	body := new(bytes.Buffer)
	mw := multipart.NewWriter(body)
	file, err := os.Open(filePath)
	assert.NoError(t, err)

	w, err := mw.CreateFormFile("field", filePath)
	assert.NoError(t, err)
	_, err = io.Copy(w, file)
	assert.NoError(t, err)
	mw.Close()

	req := httptest.NewRequest("POST", "/test", body)
	req.Header.Add("Content-Type", mw.FormDataContentType())
	resp, err := server.Test(req)
	assert.NoError(t, err)
	defer closeResponse(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, "custom-header-value", resp.Header.Get("custom-header"))
}
