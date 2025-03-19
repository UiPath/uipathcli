package plugin

import (
	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/utils/network"
)

// AuthResult provides authentication information provided by the configured
// authenticators. Typically, it contains the JWT bearer token for authenticating
// with the external APIs.
type AuthResult struct {
	Token *auth.AuthToken
}

func (a AuthResult) ToAuthorization() *network.Authorization {
	if a.Token == nil {
		return nil
	}
	return network.NewAuthorization(a.Token.Type, a.Token.Value)
}
