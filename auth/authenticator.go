package auth

type Authenticator interface {
	Auth(ctx AuthenticatorContext) AuthenticatorResult
	CanAuthenticate(ctx AuthenticatorContext) bool
}
