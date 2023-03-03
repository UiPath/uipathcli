// Package auth implements multiple schemas for authentication CLI requests.
// - Supports JWT bearer token, OAuth flows and Personal access token out-of-the-box.
// - Provides interfaces to implement custom authenticators.
package auth

// Authenticator interface for providing auth credentials.
type Authenticator interface {
	Auth(ctx AuthenticatorContext) AuthenticatorResult
}
