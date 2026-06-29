package auth_test

import (
	"bytes"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/logger"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/restfulresolver"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createAuthedServer builds a server that injects the given user into context
// (simulating an authenticated request) before the resources run.
func createAuthedServer(t *testing.T, user *fs.User, resources ...*fs.Resource) *restfulresolver.Server {
	rm := fs.NewResourcesManager()
	rm.Middlewares = append(rm.Middlewares, func(c fs.Context) error {
		if user != nil {
			c.Local("user", user)
		}
		return c.Next()
	})
	for _, r := range resources {
		rm.Add(r)
	}
	require.NoError(t, rm.Init())
	return restfulresolver.NewRestfulResolver(&restfulresolver.ResolverConfig{
		ResourceManager: rm,
		Logger:          logger.CreateMockLogger(true),
	}).Server()
}

func tokenFromConfirmMail(t *testing.T, mailer *MockMailer, newEmail string) string {
	t.Helper()
	for _, m := range mailer.GetSentMails() {
		if len(m.To) == 1 && m.To[0] == newEmail && strings.Contains(m.Body, "token=") {
			raw := m.Body[strings.Index(m.Body, "token=")+len("token="):]
			raw = strings.Fields(raw)[0] // token is the last token in the body
			tok, err := url.QueryUnescape(raw)
			require.NoError(t, err)
			return tok
		}
	}
	t.Fatalf("confirm mail to %s not found", newEmail)
	return ""
}

func TestChangeEmail_WrongPassword(t *testing.T) {
	config := &testAppConfig{activation: "manual", createData: true, mailer: &MockMailer{}}
	provider := createLocalAuthProvider(config)
	authUser := &fs.User{ID: config.user02ID}
	server := createAuthedServer(t, authUser,
		fs.Post("email/change", provider.ChangeEmail, &fs.Meta{Public: true}))

	body := []byte(`{"new_email":"changed@site.local","current_password":"wrong"}`)
	req := httptest.NewRequest("POST", "/email/change", bytes.NewReader(body))
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 422, resp.StatusCode)
	assert.Contains(t, utils.Must(utils.ReadCloserToString(resp.Body)), auth.MSG_INCORRECT_PASSWORD)
}

func TestChangeEmail_Unauthenticated(t *testing.T) {
	config := &testAppConfig{activation: "manual", createData: true, mailer: &MockMailer{}}
	provider := createLocalAuthProvider(config)
	server := createAuthedServer(t, nil, // no user injected
		fs.Post("email/change", provider.ChangeEmail, &fs.Meta{Public: true}))

	body := []byte(`{"new_email":"changed@site.local","current_password":"user02"}`)
	req := httptest.NewRequest("POST", "/email/change", bytes.NewReader(body))
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.Equal(t, 401, resp.StatusCode)
}

func TestEmailChange_FullFlowAndSingleUse(t *testing.T) {
	mailer := &MockMailer{}
	config := &testAppConfig{activation: "manual", createData: true, mailer: mailer}
	provider := createLocalAuthProvider(config)
	authUser := &fs.User{ID: config.user02ID}
	server := createAuthedServer(t, authUser,
		fs.Post("email/change", provider.ChangeEmail, &fs.Meta{Public: true}),
		fs.Post("email/confirm", provider.ConfirmEmailChange, &fs.Meta{Public: true}))

	const newEmail = "newuser02@site.local"

	// 1. Request the change.
	req := httptest.NewRequest("POST", "/email/change",
		bytes.NewReader([]byte(`{"new_email":"`+newEmail+`","current_password":"user02"}`)))
	resp, _ := server.Test(req)
	require.Equal(t, 200, resp.StatusCode)
	assert.NoError(t, resp.Body.Close())

	// 2. Dual email: notify old + confirm new (sent asynchronously).
	require.Eventually(t, func() bool {
		return len(mailer.GetSentMails()) == 2
	}, 2*time.Second, 10*time.Millisecond)
	mails := mailer.GetSentMails()
	require.Len(t, mails, 2)
	var toOld, toNew bool
	for _, m := range mails {
		if m.To[0] == "user02@site.local" {
			toOld = true
		}
		if m.To[0] == newEmail {
			toNew = true
		}
	}
	assert.True(t, toOld, "notify mail to old address")
	assert.True(t, toNew, "confirm mail to new address")

	// 3. Confirm with the token from the new-address mail.
	token := tokenFromConfirmMail(t, mailer, newEmail)
	confirmReq := httptest.NewRequest("POST", "/email/confirm",
		bytes.NewReader([]byte(`{"token":"`+token+`"}`)))
	confirmResp, _ := server.Test(confirmReq)
	require.Equal(t, 200, confirmResp.StatusCode)
	assert.NoError(t, confirmResp.Body.Close())

	// 4. Email committed.
	user := utils.Must(db.Builder[*fs.User](config.db).
		Where(db.EQ("id", config.user02ID)).First(req.Context()))
	assert.Equal(t, newEmail, user.Email)

	// 5. Single-use: re-confirming the same token fails.
	confirmReq2 := httptest.NewRequest("POST", "/email/confirm",
		bytes.NewReader([]byte(`{"token":"`+token+`"}`)))
	confirmResp2, _ := server.Test(confirmReq2)
	defer func() { assert.NoError(t, confirmResp2.Body.Close()) }()
	assert.GreaterOrEqual(t, confirmResp2.StatusCode, 400)
}

func TestConfirmEmailChange_InvalidToken(t *testing.T) {
	config := &testAppConfig{activation: "manual", createData: true, mailer: &MockMailer{}}
	provider := createLocalAuthProvider(config)
	server := createAuthedServer(t, nil,
		fs.Post("email/confirm", provider.ConfirmEmailChange, &fs.Meta{Public: true}))

	req := httptest.NewRequest("POST", "/email/confirm",
		bytes.NewReader([]byte(`{"token":"garbage"}`)))
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.GreaterOrEqual(t, resp.StatusCode, 400)
}
