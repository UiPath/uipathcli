package auth

// The AuthenticatorResult indicates if the authentication was successful
// and returns the authentication credentials.
type AuthenticatorResult struct {
	Error error
	Token *AuthToken
}

func AuthenticatorError(err error) *AuthenticatorResult {
	return &AuthenticatorResult{err, nil}
}

func AuthenticatorSuccess(token *AuthToken) *AuthenticatorResult {
	return &AuthenticatorResult{nil, token}
}
