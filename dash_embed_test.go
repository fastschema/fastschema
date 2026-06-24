package fastschema

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
)

// TestDashEmbedServesAssets verifies the compiled-in dash bundle is served at
// the dash base path and that its HTML references the matching asset base, so a
// stale or wrongly-based bundle is caught before release. The Fiber mount mirrors
// the production static handler in pkg/restfulresolver/resource.go.
func TestDashEmbedServesAssets(t *testing.T) {
	app := fiber.New()
	app.Use("/dash", filesystem.New(filesystem.Config{
		Root:       http.FS(embedDashStatic),
		PathPrefix: "dash",
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/" || c.Path() == ""
		},
	}))

	// index.html is served at /dash/ and points assets at the /dash/ base.
	resp, err := app.Test(httptest.NewRequest(http.MethodGet, "/dash/", nil))
	if err != nil {
		t.Fatalf("request /dash/ failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("/dash/ status = %d, want 200", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "/dash/assets/") {
		t.Fatal("/dash/ index.html does not reference the /dash/assets/ base")
	}

	// A hashed JS asset (name varies per build) must serve 200.
	entries, err := embedDashStatic.ReadDir("dash/assets")
	if err != nil {
		t.Fatalf("read embedded dash/assets: %v", err)
	}
	var asset string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".js") {
			asset = e.Name()
			break
		}
	}
	if asset == "" {
		t.Fatal("no .js asset found in embedded dash/assets")
	}

	resp2, err := app.Test(httptest.NewRequest(http.MethodGet, "/dash/assets/"+asset, nil))
	if err != nil {
		t.Fatalf("request /dash/assets/%s failed: %v", asset, err)
	}
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("/dash/assets/%s status = %d, want 200", asset, resp2.StatusCode)
	}
}

// TestDashEmbedSPAFallbackServesIndex verifies that deep links into the dash
// single-page application resolve to index.html instead of 404. The dash is a
// client-routed SPA with a single HTML entry, so hard navigations such as the
// setup URL (/dash/setup/?token=...) have no matching file on disk and must be
// served the SPA shell. The Fiber mount mirrors the production static handler
// in pkg/restfulresolver/resource.go, including the NotFoundFile fallback.
func TestDashEmbedSPAFallbackServesIndex(t *testing.T) {
	app := fiber.New()
	app.Use("/dash", filesystem.New(filesystem.Config{
		Root:         http.FS(embedDashStatic),
		PathPrefix:   "dash",
		NotFoundFile: "dash/index.html",
		Next: func(c *fiber.Ctx) bool {
			return c.Path() == "/" || c.Path() == ""
		},
	}))

	// Deep SPA routes that have no matching file must serve the index shell
	// with a 200 so the client router can render them on hard navigation.
	for _, path := range []string{"/dash/setup/", "/dash/login", "/dash/content"} {
		resp, err := app.Test(httptest.NewRequest(http.MethodGet, path, nil))
		if err != nil {
			t.Fatalf("request %s failed: %v", path, err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s status = %d, want 200 (SPA fallback)", path, resp.StatusCode)
		}
		body, _ := io.ReadAll(resp.Body)
		if !strings.Contains(string(body), "/dash/assets/") {
			t.Fatalf("%s did not serve the SPA index shell", path)
		}
	}
}
