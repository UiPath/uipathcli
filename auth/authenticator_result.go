package auth

// The AuthenticatorResult indicates if the authentication was successful
// and returns the authentication credentials.
type AuthenticatorResult struct {
	Error string
	Token *AuthToken
}

func AuthenticatorError(err error) *AuthenticatorResult {
	return &AuthenticatorResult{err.Error(), nil}
}

func AuthenticatorSuccess(token *AuthToken) *AuthenticatorResult {
	return &AuthenticatorResult{"", token}
}
