package plugin

import "github.com/UiPath/uipathcli/auth"

// AuthResult provides authentication information provided by the configured
// authenticators. Typically, it contains the JWT bearer token for authenticating
// with the external APIs.
type AuthResult struct {
	Token *auth.AuthToken
}
