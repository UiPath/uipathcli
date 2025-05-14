package auth

type tokenResponseJson struct {
	AccessToken  string  `json:"access_token"`
	ExpiresIn    float32 `json:"expires_in"`
	TokenType    string  `json:"token_type"`
	Scope        string  `json:"scope"`
	RefreshToken *string `json:"refresh_token,omitempty"`
}

func (r tokenResponseJson) Redacted() tokenResponseJson {
	return tokenResponseJson{
		AccessToken:  r.AccessToken,
		ExpiresIn:    r.ExpiresIn,
		TokenType:    r.TokenType,
		Scope:        r.Scope,
		RefreshToken: r.redact(r.RefreshToken),
	}
}

func (r tokenResponseJson) redact(value *string) *string {
	if value == nil {
		return nil
	}
	result := redactedMessage
	return &result
}
