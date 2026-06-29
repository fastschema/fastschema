package auth_test

import (
	"bytes"
	"context"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/fastschema/fastschema/db"
	"github.com/fastschema/fastschema/fs"
	"github.com/fastschema/fastschema/pkg/auth"
	"github.com/fastschema/fastschema/pkg/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func runPolicy(p *fs.RegistrationPolicy, in *fs.RegistrationInput) error {
	return auth.BuiltinPolicyValidator(p)(context.Background(), in)
}

func TestBuiltinPolicy_AllowlistAndBlocklist(t *testing.T) {
	t.Run("allowlist rejects non-listed domain", func(t *testing.T) {
		p := &fs.RegistrationPolicy{AllowedEmailDomains: []string{"acme.com"}}
		err := runPolicy(p, &fs.RegistrationInput{Email: "a@evil.com"})
		require.Error(t, err)
	})
	t.Run("allowlist passes listed domain (case-insensitive)", func(t *testing.T) {
		p := &fs.RegistrationPolicy{AllowedEmailDomains: []string{"ACME.com"}}
		require.NoError(t, runPolicy(p, &fs.RegistrationInput{Email: "a@acme.com"}))
	})
	t.Run("blocklist rejects listed domain", func(t *testing.T) {
		p := &fs.RegistrationPolicy{BlockedEmailDomains: []string{"spam.com"}}
		require.Error(t, runPolicy(p, &fs.RegistrationInput{Email: "a@spam.com"}))
	})
	t.Run("empty policy blocks nothing", func(t *testing.T) {
		require.NoError(t, runPolicy(&fs.RegistrationPolicy{}, &fs.RegistrationInput{Email: "a@anything.io"}))
	})
}

func TestBuiltinPolicy_ReservedUsername(t *testing.T) {
	p := &fs.RegistrationPolicy{ReservedUsernames: []string{"admin", "root"}}
	require.Error(t, runPolicy(p, &fs.RegistrationInput{Email: "a@x.com", Username: "Admin"}))
	require.NoError(t, runPolicy(p, &fs.RegistrationInput{Email: "a@x.com", Username: "alice"}))
	// empty username is not checked
	require.NoError(t, runPolicy(p, &fs.RegistrationInput{Email: "a@x.com"}))
}

func TestBuiltinPolicy_NormalizeEmailMutates(t *testing.T) {
	p := &fs.RegistrationPolicy{NormalizeEmail: true}
	in := &fs.RegistrationInput{Email: "User@GMAIL.COM"}
	require.NoError(t, runPolicy(p, in))
	// local part casing preserved, domain lowercased
	assert.Equal(t, "User@gmail.com", in.Email)
}

func TestNormalizeEmail(t *testing.T) {
	// domain lowercased, local part casing preserved
	assert.Equal(t, "Foo@gmail.com", auth.NormalizeEmail("Foo@GMAIL.com"))
	// IDN domain converted to punycode (ASCII)
	assert.Equal(t, "a@xn--bcher-kva.de", auth.NormalizeEmail("a@bücher.de"))
	// surrounding whitespace trimmed
	assert.Equal(t, "a@gmail.com", auth.NormalizeEmail("  a@Gmail.com  "))
	// no domain → trimmed input returned
	assert.Equal(t, "notanemail", auth.NormalizeEmail("  notanemail  "))
}

func TestRunPreUserRegisterHooks_OrderAndShortCircuit(t *testing.T) {
	var calls []string
	h1 := func(_ context.Context, in *fs.RegistrationInput) error {
		calls = append(calls, "h1")
		in.Email = "normalized@x.com"
		return nil
	}
	boom := errors.New("rejected")
	h2 := func(_ context.Context, in *fs.RegistrationInput) error {
		calls = append(calls, "h2")
		return boom
	}
	h3 := func(_ context.Context, in *fs.RegistrationInput) error {
		calls = append(calls, "h3")
		return nil
	}

	in := &fs.RegistrationInput{Email: "raw@x.com"}
	err := fs.RunPreUserRegisterHooks(context.Background(), []fs.PreUserRegisterHook{h1, h2, h3}, in)

	require.ErrorIs(t, err, boom)
	assert.Equal(t, []string{"h1", "h2"}, calls)  // h3 not reached (short-circuit)
	assert.Equal(t, "normalized@x.com", in.Email) // h1 mutation propagated
}

// TestLocalRegister_HookWiring proves the local Register path fires the
// PreUserRegister chain (reject aborts) and applies hook mutations to the
// persisted entity.
func TestLocalRegister_HookReject(t *testing.T) {
	config := &testAppConfig{
		activation: "manual",
		createData: true,
		preUserRegister: func(_ context.Context, in *fs.RegistrationInput) error {
			return auth.ERR_SAVE_USER // any error aborts; use an existing one
		},
	}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post("/user/register", provider.Register, &fs.Meta{Public: true}))

	body := []byte(`{"username":"rejectme","email":"reject@local.ltd","password":"p","confirm_password":"p"}`)
	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	assert.GreaterOrEqual(t, resp.StatusCode, 400)

	// No user row created for the rejected email.
	count := utils.Must(db.Builder[*fs.User](config.db).
		Where(db.EQ("email", "reject@local.ltd")).Count(context.Background()))
	assert.Equal(t, 0, count)
}

func TestLocalRegister_HookNormalizesEmail(t *testing.T) {
	config := &testAppConfig{
		activation:         "manual",
		createData:         true,
		registrationPolicy: &fs.RegistrationPolicy{NormalizeEmail: true},
		preUserRegister:    auth.BuiltinPolicyValidator(&fs.RegistrationPolicy{NormalizeEmail: true}),
	}
	provider := createLocalAuthProvider(config)
	server := createServer(t, fs.Post("/user/register", provider.Register, &fs.Meta{Public: true}))

	body := []byte(`{"username":"normuser","email":"Norm@LOCAL.LTD","password":"p","confirm_password":"p"}`)
	req := httptest.NewRequest("POST", "/user/register", bytes.NewReader(body))
	resp, _ := server.Test(req)
	defer func() { assert.NoError(t, resp.Body.Close()) }()
	require.Equal(t, 200, resp.StatusCode)

	// Domain lowercased, local part preserved.
	user := utils.Must(db.Builder[*fs.User](config.db).
		Where(db.EQ("username", "normuser")).First(context.Background()))
	assert.Equal(t, "Norm@local.ltd", user.Email)
}

func TestRunPreUserRegisterHooks_EmptyAndNil(t *testing.T) {
	require.NoError(t, fs.RunPreUserRegisterHooks(context.Background(), nil, &fs.RegistrationInput{}))
	require.NoError(t, fs.RunPreUserRegisterHooks(
		context.Background(),
		[]fs.PreUserRegisterHook{nil},
		&fs.RegistrationInput{},
	))
}
