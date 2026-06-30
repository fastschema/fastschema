package authservice

import (
	"github.com/fastschema/fastschema/fs"
)

// methodsDTO is the public, whitelisted shape of the enabled login methods. It
// is an explicit allowlist: ONLY provider names and the otp flag are exposed.
// The raw AuthConfig.Providers map holds client_id/client_secret, so the config
// object must NEVER be serialized here. The method list is public knowledge
// (it is rendered on the login screen); the secrets are not.
type methodsDTO struct {
	Providers []string `json:"providers"`
	OTP       bool     `json:"otp"`
}

// AuthMethods returns the enabled login methods so the dash can render the
// correct buttons. Public and always available (not gated by the cli flag).
func (as *AuthService) AuthMethods(c fs.Context, _ any) (*methodsDTO, error) {
	dto := &methodsDTO{Providers: []string{}}

	cfg := as.AppConfig()
	if cfg == nil || cfg.AuthConfig == nil {
		return dto, nil
	}

	dto.Providers = append(dto.Providers, cfg.AuthConfig.EnabledProviders...)
	dto.OTP = cfg.AuthConfig.OTP != nil && cfg.AuthConfig.OTP.Enabled

	return dto, nil
}
