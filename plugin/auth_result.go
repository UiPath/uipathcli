package plugin

// AuthResult provides authentication information provided by the configured
// authenticators. Typically, it contains the Authorization HTTP header with
// a bearer token.
type AuthResult struct {
	Header map[string]string
}
