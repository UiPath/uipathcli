package auth

// The AuthenticatorResult indicates if the authentication was successful
// and returns the authentication credentials.
type AuthenticatorResult struct {
	Error         string                 `json:"error"`
	RequestHeader map[string]string      `json:"requestHeader"`
	Config        map[string]interface{} `json:"config"`
}

func AuthenticatorError(err error) *AuthenticatorResult {
	return &AuthenticatorResult{err.Error(), map[string]string{}, map[string]interface{}{}}
}

func AuthenticatorSuccess(requestHeader map[string]string, config map[string]interface{}) *AuthenticatorResult {
	return &AuthenticatorResult{"", requestHeader, config}
}
