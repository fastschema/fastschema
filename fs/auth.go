package fs

import (
	"fmt"
	"sort"
	"sync"
)

type IDToken struct {
	IDToken string `json:"id_token"`
}

type AuthProviderMaker func(Map, string) (AuthProvider, error)
type AuthProvider interface {
	Name() string
	Login(Context) (any, error)
	Callback(Context) (*User, error)
	VerifyIDToken(Context, IDToken) (*User, error)
}

var (
	authProviderMakersMu sync.RWMutex
	authProviderMakers   = make(map[string]AuthProviderMaker)
)

// RegisterAuthProviderMaker makes an auth provider factory available by the provided name.
// If RegisterAuthProviderMaker is called twice with the same name or if auth provider factory is nil, it panics.
func RegisterAuthProviderMaker(name string, fn AuthProviderMaker) {
	authProviderMakersMu.Lock()
	defer authProviderMakersMu.Unlock()
	if fn == nil {
		panic("auth: Register auth provider is nil")
	}
	if _, dup := authProviderMakers[name]; dup {
		panic("auth: Register called twice for auth provider " + name)
	}
	authProviderMakers[name] = fn
}

// CreateAuthProvider creates an auth provider by the provided name.
func CreateAuthProvider(name string, config Map, redirectURL string) (AuthProvider, error) {
	authProviderMakersMu.RLock()
	defer authProviderMakersMu.RUnlock()
	fn, ok := authProviderMakers[name]
	if !ok {
		return nil, fmt.Errorf("auth: unknown auth provider %q", name)
	}
	return fn(config, redirectURL)
}

// AuthProviders returns a sorted list of the names of the registered auth providers.
func AuthProviders() []string {
	authProviderMakersMu.RLock()
	defer authProviderMakersMu.RUnlock()
	list := make([]string, 0, len(authProviderMakers))
	for name := range authProviderMakers {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

type AuthConfig struct {
	EnabledProviders []string       `json:"enabled_providers"`
	Providers        map[string]Map `json:"providers"`
}

func (ac *AuthConfig) Clone() *AuthConfig {
	if ac == nil {
		return nil
	}

	clone := &AuthConfig{
		EnabledProviders: make([]string, len(ac.EnabledProviders)),
		Providers:        make(map[string]Map),
	}

	copy(clone.EnabledProviders, ac.EnabledProviders)

	for k, v := range ac.Providers {
		clone.Providers[k] = v
	}

	return clone
}
