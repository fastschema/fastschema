package fs

type ProviderConfig map[string]string

type AuthConfig struct {
	EnabledProviders []string                  `json:"enabled_providers"`
	Providers        map[string]ProviderConfig `json:"providers"`
}

type AuthProvider interface {
	Name() string
	Login(Context) (any, error)
	Callback(Context) (*User, error)
}

type CreateAuthProviderFunc func(
	map[string]string,
	string,
) (AuthProvider, error)

func (ac *AuthConfig) Clone() *AuthConfig {
	if ac == nil {
		return nil
	}

	clone := &AuthConfig{
		EnabledProviders: make([]string, len(ac.EnabledProviders)),
		Providers:        make(map[string]ProviderConfig),
	}

	copy(clone.EnabledProviders, ac.EnabledProviders)

	for k, v := range ac.Providers {
		clone.Providers[k] = v
	}

	return clone
}
