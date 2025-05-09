package auth

import "net/url"

type tokenRequestForm struct {
	GrantType    string
	Scopes       string
	ClientId     string
	ClientSecret string
	Code         string
	CodeVerifier string
	RedirectUri  string
	Properties   map[string]string
	RefreshToken string
}

func (r tokenRequestForm) Encode() string {
	form := r.createForm(
		r.GrantType,
		r.Scopes,
		r.ClientId,
		r.ClientSecret,
		r.Code,
		r.CodeVerifier,
		r.RedirectUri,
		r.RefreshToken,
		r.Properties,
	)
	return form.Encode()
}

func (r tokenRequestForm) Redacted() string {
	form := r.createForm(
		r.GrantType,
		r.Scopes,
		r.ClientId,
		r.redact(r.ClientSecret),
		r.Code,
		r.CodeVerifier,
		r.RedirectUri,
		r.redact(r.RefreshToken),
		r.Properties,
	)
	return form.Encode()
}

func (r tokenRequestForm) createForm(
	grantType string,
	scopes string,
	clientId string,
	clientSecret string,
	code string,
	codeVerifier string,
	redirectUri string,
	refreshToken string,
	properties map[string]string,
) url.Values {
	form := url.Values{}
	form.Add("grant_type", grantType)
	form.Add("scope", scopes)
	form.Add("client_id", clientId)
	form.Add("client_secret", clientSecret)
	form.Add("code", code)
	form.Add("code_verifier", codeVerifier)
	form.Add("redirect_uri", redirectUri)
	form.Add("refresh_token", refreshToken)
	for key, value := range properties {
		form.Add(key, value)
	}
	return form
}

func (r tokenRequestForm) redact(value string) string {
	if value == "" {
		return ""
	}
	return redactedMessage
}
