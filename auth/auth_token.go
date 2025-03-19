package auth

// AuthToken for authenticating with external APIs. Tokens are typically
// JWT bearer tokens.
type AuthToken struct {
	Type  string
	Value string
}

func NewBearerToken(value string) *AuthToken {
	return &AuthToken{"Bearer", value}
}
