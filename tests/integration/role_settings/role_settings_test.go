package role_settings_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/fastschema/fastschema"
	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/entity"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/entdbadapter"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/fastschema/fastschema/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testEnv holds the test environment with app, server, and tokens
type testEnv struct {
	t          *testing.T
	app        *fastschema.App
	server     *restfulresolver.Server
	adminToken string
	userToken  string
}

// cleanup closes the database connection
func (te *testEnv) cleanup() {
	te.app.DB().Close()
}

// clearEnvs clears environment variables that might interfere with tests
func clearEnvs(t *testing.T) {
	t.Helper()
	envVars := []string{
		"APP_KEY", "APP_PORT", "APP_BASE_URL", "APP_DASH_URL",
		"APP_API_BASE_NAME", "APP_DASH_BASE_NAME",
		"DB_DRIVER", "DB_NAME", "DB_HOST", "DB_PORT", "DB_USER", "DB_PASS",
		"STORAGE", "AUTH", "MAIL", "ROLE_PERMISSION_SETTINGS",
	}
	for _, env := range envVars {
		t.Setenv(env, "")
	}
}

// createTestEnv creates a complete test environment with app, server, admin user, and regular user
func createTestEnv(t *testing.T, permSettings *fs.RolePermissionSettingsConfig) *testEnv {
	t.Helper()
	clearEnvs(t)

	// Create schema builder and database
	sb := utils.Must(schema.NewBuilderFromDir(t.TempDir(), fs.SystemSchemaTypes...))
	entDB := utils.Must(entdbadapter.NewTestClient(
		utils.Must(os.MkdirTemp("", "migrations")), sb, nil,
	))

	// Create app with optional permission settings
	config := &fs.Config{
		HideResourcesInfo:      true,
		Dir:                    t.TempDir(),
		DB:                     entDB,
		RolePermissionSettings: permSettings,
	}
	app, err := fastschema.New(config)
	require.NoError(t, err)

	// Initialize resources and create server
	resources := app.Resources()
	require.NoError(t, resources.Init())
	server := restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: resources,
		Logger:          logger.CreateMockLogger(true),
	}).Server()

	te := &testEnv{t: t, app: app, server: server}

	// Setup admin user
	te.adminToken = te.setupAdmin()

	// Create and login regular user
	te.userToken = te.createAndLoginUser("testuser", "test@local.ltd", "123")

	return te
}

// setupAdmin sets up the app with admin user and returns admin token
func (te *testEnv) setupAdmin() string {
	setupToken := utils.Must(te.app.GetSetupToken(context.Background()))
	req := httptest.NewRequest("POST", "/api/setup", bytes.NewReader([]byte(`{
		"token":"`+setupToken+`",
		"username":"admin",
		"email":"admin@local.ltd",
		"password":"123"
	}`)))
	resp := utils.Must(te.server.Test(req))
	defer resp.Body.Close()
	require.Equal(te.t, 200, resp.StatusCode, "Setup should succeed")

	return te.login("admin", "123")
}

// createAndLoginUser creates a user and returns their token
func (te *testEnv) createAndLoginUser(username, email, password string) string {
	ctx := context.Background()
	_, err := db.Create[*fs.User](ctx, te.app.DB(), fs.Map{
		"username": username,
		"email":    email,
		"password": password,
		"active":   true,
		"roles":    []*entity.Entity{entity.New(fs.RoleUser.ID)},
	})
	require.NoError(te.t, err)

	return te.login(username, password)
}

// login authenticates a user and returns their token
func (te *testEnv) login(username, password string) string {
	req := httptest.NewRequest("POST", "/api/auth/local/login", bytes.NewReader([]byte(`{
		"login":"`+username+`",
		"password":"`+password+`"
	}`)))
	resp := utils.Must(te.server.Test(req))
	defer resp.Body.Close()
	require.Equal(te.t, 200, resp.StatusCode, "Login should succeed")

	response := utils.Must(utils.ReadCloserToString(resp.Body))
	token := strings.Split(response, `"token":"`)[1]
	return strings.Split(token, `"`)[0]
}

// addDBPermission adds a permission to the database and reloads cache
func (te *testEnv) addDBPermission(resource, value string, roleID uint64) {
	ctx := context.Background()
	permissionModel := utils.Must(te.app.DB().Model("permission"))
	_, err := permissionModel.CreateFromJSON(ctx, `{
		"resource": "`+resource+`",
		"value": "`+value+`",
		"role_id": `+utils.If(roleID == 0, "2", string(rune('0'+roleID)))+`
	}`)
	require.NoError(te.t, err)
	require.NoError(te.t, te.app.UpdateCache(ctx))
}

// get makes a GET request with optional auth token
func (te *testEnv) get(path, token string) *http.Response {
	req := httptest.NewRequest("GET", path, nil)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return utils.Must(te.server.Test(req))
}

func TestExportEndpointViaHTTP(t *testing.T) {
	te := createTestEnv(t, nil)
	defer te.cleanup()

	// Add a permission so export has data
	te.addDBPermission("api.test.resource", "allow", fs.RoleUser.ID)

	resp := te.get("/api/role/export", te.adminToken)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var apiResp struct {
		Data fs.RolePermissionSettingsConfig `json:"data"`
	}
	respBody := utils.Must(utils.ReadCloserToString(resp.Body))
	require.NoError(t, json.Unmarshal([]byte(respBody), &apiResp))

	// Verify User role has the permission we added
	var userConfig *fs.RoleConfig
	for _, rc := range apiResp.Data.Roles {
		if rc.Name == "User" {
			userConfig = rc
			break
		}
	}
	require.NotNil(t, userConfig, "User role should be in export")
	assert.GreaterOrEqual(t, len(userConfig.Permissions), 1)
}

func TestEnvConfigDeniesWhenDBAllows(t *testing.T) {
	// Env config: DENY User role access to api.role.list
	permSettings := &fs.RolePermissionSettingsConfig{
		Roles: []*fs.RoleConfig{{
			Name:        "User",
			Permissions: []*fs.Permission{{Resource: "api.role.list", Value: "deny"}},
		}},
	}

	te := createTestEnv(t, permSettings)
	defer te.cleanup()

	// Add DB permission that ALLOWS (env should override to deny)
	te.addDBPermission("api.role.list", "allow", fs.RoleUser.ID)

	resp := te.get("/api/role", te.userToken)
	defer resp.Body.Close()

	assert.Equal(t, 403, resp.StatusCode, "Should be denied - env config overrides DB allow")
}

func TestEnvConfigAllowsWhenDBDenies(t *testing.T) {
	// Env config: ALLOW User role access to api.role.list
	permSettings := &fs.RolePermissionSettingsConfig{
		Roles: []*fs.RoleConfig{{
			Name:        "User",
			Permissions: []*fs.Permission{{Resource: "api.role.list", Value: "allow"}},
		}},
	}

	te := createTestEnv(t, permSettings)
	defer te.cleanup()

	// Add DB permission that DENIES (env should override to allow)
	te.addDBPermission("api.role.list", "deny", fs.RoleUser.ID)

	resp := te.get("/api/role", te.userToken)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode, "Should be allowed - env config overrides DB deny")
}

func TestExportOutputIsValidForImport(t *testing.T) {
	te := createTestEnv(t, nil)
	defer te.cleanup()

	resp := te.get("/api/role/export", te.adminToken)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)

	var apiResp struct {
		Data fs.RolePermissionSettingsConfig `json:"data"`
	}
	respBody := utils.Must(utils.ReadCloserToString(resp.Body))
	require.NoError(t, json.Unmarshal([]byte(respBody), &apiResp))

	assert.NoError(t, apiResp.Data.Validate(), "Exported config should be valid for import")
}

func TestInvalidConfigValidation(t *testing.T) {
	tests := []struct {
		name   string
		config *fs.RolePermissionSettingsConfig
	}{
		{
			name: "empty role name",
			config: &fs.RolePermissionSettingsConfig{
				Roles: []*fs.RoleConfig{{Name: "", Permissions: []*fs.Permission{{Resource: "test", Value: "allow"}}}},
			},
		},
		{
			name: "empty resource",
			config: &fs.RolePermissionSettingsConfig{
				Roles: []*fs.RoleConfig{{Name: "User", Permissions: []*fs.Permission{{Resource: "", Value: "allow"}}}},
			},
		},
		{
			name: "empty value",
			config: &fs.RolePermissionSettingsConfig{
				Roles: []*fs.RoleConfig{{Name: "User", Permissions: []*fs.Permission{{Resource: "test", Value: ""}}}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Error(t, tt.config.Validate())
		})
	}
}

func TestEnvParsingIntegration(t *testing.T) {
	jsonConfig := `{
		"roles": [{
			"name": "User",
			"permissions": [
				{"resource": "content.blog.list", "value": "allow"},
				{"resource": "content.blog.create", "value": "deny", "modifier": "let x = 1"}
			]
		}]
	}`

	var config fs.RolePermissionSettingsConfig
	require.NoError(t, json.Unmarshal([]byte(jsonConfig), &config))
	require.NoError(t, config.Validate())

	assert.Len(t, config.Roles, 1)
	assert.Equal(t, "User", config.Roles[0].Name)
	assert.Len(t, config.Roles[0].Permissions, 2)
	assert.Equal(t, "let x = 1", config.Roles[0].Permissions[1].Modifier)
}
