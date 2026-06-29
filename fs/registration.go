package fs

import "context"

// RegistrationInput carries the self-service signup data passed to PreUserRegister
// hooks. It is shared by the local (email/password) and OAuth registration paths.
//
// Hooks receive a pointer and MAY mutate Email/Username (e.g. normalization);
// the caller applies the mutated values to the persisted entity. Returning a
// non-nil error aborts registration before the user row is created.
type RegistrationInput struct {
	Email      string         // signup email (mutable)
	Username   string         // signup username (mutable)
	Provider   string         // "local" or the OAuth provider name
	ProviderID string         // OAuth provider subject id (empty for local)
	Profile    map[string]any // optional raw provider profile (OAuth)
	IsOAuth    bool           // true for social-login registration
}

// PreUserRegisterHook runs just before a self-service user row is created.
// It does NOT fire for admin-created users (content API).
type PreUserRegisterHook func(ctx context.Context, in *RegistrationInput) error

// RunPreUserRegisterHooks executes the chain in order, short-circuiting on the
// first error. nil hooks are skipped. This is the single execution path shared
// by both registration call sites (local + OAuth).
func RunPreUserRegisterHooks(
	ctx context.Context,
	hooks []PreUserRegisterHook,
	in *RegistrationInput,
) error {
	for _, h := range hooks {
		if h == nil {
			continue
		}
		if err := h(ctx, in); err != nil {
			return err
		}
	}
	return nil
}
