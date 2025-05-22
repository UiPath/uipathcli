package test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

type httpFake struct {
	TokenResponse      *ResponseData
	OAuthTokenResponse *ResponseData
	Responses          map[string]ResponseData
	ResponseHandler    func(RequestData) ResponseData
}

func (f httpFake) Handle(request RequestData) *ResponseData {
	if request.URL.String() == "/identity_/connect/token" && f.TokenResponse != nil {
		return f.handleIdentityTokenCredentialsRequest(request)
	}
	if request.URL.String() == "/identity_/connect/token" && f.OAuthTokenResponse != nil {
		return f.handleIdentityTokenOAuthRequest(request)
	}

	requestUrl := request.URL.Path
	query, _ := url.QueryUnescape(request.URL.RawQuery)
	if query != "" {
		requestUrl += "?" + query
	}
	response, found := f.Responses[requestUrl]
	if !found {
		response, found = f.Responses["*"]
	}
	if found {
		return &response
	}

	if f.ResponseHandler != nil {
		response := f.ResponseHandler(request)
		return &response
	}
	return nil
}

func (f httpFake) SendOAuthResponse(uri string) error {
	queryString, _ := url.ParseQuery(uri)
	baseRedirectUri := queryString["redirect_uri"][0]
	state := queryString["state"][0]
	redirectUri := fmt.Sprintf("%s?code=mycode&state=%s", baseRedirectUri, state)
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, redirectUri, nil)
	if err != nil {
		return fmt.Errorf("Error sending oauth browser request: %w", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return fmt.Errorf("Error sending oauth browser request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Invalid response status %d when sending oauth browser request: %w", response.StatusCode, err)
	}
	return nil
}

func (f httpFake) handleIdentityTokenCredentialsRequest(request RequestData) *ResponseData {
	data, _ := url.ParseQuery(string(request.Body))
	if f.isEmpty(data, "client_id") {
		return &ResponseData{http.StatusBadRequest, "client_id is missing"}
	}
	if f.isEmpty(data, "client_secret") {
		return &ResponseData{http.StatusBadRequest, "client_secret is missing"}
	}
	if f.isEmpty(data, "grant_type") || data["grant_type"][0] != "client_credentials" {
		return &ResponseData{http.StatusBadRequest, "Invalid grant_type"}
	}
	return f.TokenResponse
}

func (f httpFake) handleIdentityTokenOAuthRequest(request RequestData) *ResponseData {
	data, _ := url.ParseQuery(string(request.Body))
	if f.isEmpty(data, "client_id") {
		return &ResponseData{http.StatusBadRequest, "client_id is missing"}
	}
	if f.isEmpty(data, "code") {
		return &ResponseData{http.StatusBadRequest, "code is missing"}
	}
	if f.isEmpty(data, "code_verifier") {
		return &ResponseData{http.StatusBadRequest, "code_verifier is missing"}
	}
	if f.isEmpty(data, "redirect_uri") {
		return &ResponseData{http.StatusBadRequest, "redirect_uri is missing"}
	}
	if f.isEmpty(data, "grant_type") || data["grant_type"][0] != "authorization_code" {
		return &ResponseData{http.StatusBadRequest, "Invalid grant_type"}
	}
	return f.OAuthTokenResponse
}

func (f httpFake) isEmpty(values url.Values, key string) bool {
	return len(values[key]) != 1 || values[key][0] == ""
}
