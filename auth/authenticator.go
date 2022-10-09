package auth

type Authenticator interface {
	Auth(ctx AuthenticatorContext) AuthenticatorResult
}
