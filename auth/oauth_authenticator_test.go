package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/UiPath/uipathcli/cache"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestOAuthAuthenticatorNotEnabled(t *testing.T) {
	config := map[string]interface{}{
		"clientId": "my-client-id",
		"scopes":   "OR.Users",
	}
	request := NewAuthenticatorRequest("http:/localhost", map[string]string{})
	context := NewAuthenticatorContext("login", config, false, false, *request)

	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
	}
	result := authenticator.Auth(*context)
	if result.Error != "" {
		t.Errorf("Expected no error when oauth flow is skipped, but got: %v", result.Error)
	}
	if len(result.RequestHeader) != 0 {
		t.Errorf("Expected request header to be empty, but got: %v", result.RequestHeader)
	}
}

func TestOAuthAuthenticatorPreservesExistingHeaders(t *testing.T) {
	config := map[string]interface{}{
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	headers := map[string]string{
		"my-header": "my-value",
	}
	request := NewAuthenticatorRequest("http:/localhost", headers)
	context := NewAuthenticatorContext("login", config, false, false, *request)

	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
	}
	result := authenticator.Auth(*context)
	if result.Error != "" {
		t.Errorf("Expected no error when oauth flow is skipped, but got: %v", result.Error)
	}
	if result.RequestHeader["my-header"] != "my-value" {
		t.Errorf("Request header should not be changed, but got: %v", result.RequestHeader)
	}
}

func TestOAuthAuthenticatorInvalidRequestUrl(t *testing.T) {
	config := map[string]interface{}{
		"clientId":    "my-client-id",
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	request := NewAuthenticatorRequest("://invalid", map[string]string{})
	context := NewAuthenticatorContext("login", config, false, false, *request)

	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
	}
	result := authenticator.Auth(*context)
	if result.Error != `Invalid request url '://invalid': parse "://invalid": missing protocol scheme` {
		t.Errorf("Expected error with invalid request url, but got: %v", result.Error)
	}
}

func TestOAuthAuthenticatorInvalidIdentityUrl(t *testing.T) {
	config := map[string]interface{}{
		"clientId":    "my-client-id",
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	request := NewAuthenticatorRequest("INVALID-URL", map[string]string{})
	context := NewAuthenticatorContext("login", config, false, false, *request)

	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
	}
	result := authenticator.Auth(*context)
	if result.Error != `Invalid identity url 'INVALID-URL': parse ":///identity_": missing protocol scheme` {
		t.Errorf("Expected error with invalid request url, but got: %v", result.Error)
	}
}

func TestOAuthAuthenticatorInvalidConfig(t *testing.T) {
	config := map[string]interface{}{
		"clientId":    1,
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	request := NewAuthenticatorRequest("http:/localhost", map[string]string{})
	context := NewAuthenticatorContext("login", config, false, false, *request)

	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
	}
	result := authenticator.Auth(*context)
	if result.Error != "Invalid oauth authenticator configuration: Invalid value for clientId: '1'" {
		t.Errorf("Expected error with invalid config, but got: %v", result.Error)
	}
}

func TestOAuthFlowIdentityFails(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: 400,
		ResponseBody:   "Invalid token request",
	}
	identityUrl := identityServerFake.Start(t)

	context := createAuthContext(identityUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "Error retrieving access token: Token service returned status code '400' and body 'Invalid token request'" {
		t.Errorf("Expected error when identity call fails, but got: %v", result.Error)
	}
}

func TestOAuthFlowSuccessful(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: 200,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityUrl := identityServerFake.Start(t)

	context := createAuthContext(identityUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	authorizationHeader := result.RequestHeader["Authorization"]
	if authorizationHeader != "Bearer my-access-token" {
		t.Errorf("Expected JWT bearer token in authorization header, but got: %v", authorizationHeader)
	}
}

func TestOAuthFlowWithCustomIdentityUri(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: 200,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityUrl := identityServerFake.Start(t)
	config := map[string]interface{}{
		"clientId":    strconv.Itoa(rand.Int()),
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
		"uri":         identityUrl.String() + "/identity_",
	}
	request := NewAuthenticatorRequest("no-url", map[string]string{})
	context := NewAuthenticatorContext("login", config, false, false, *request)

	loginUrl, resultChannel := callAuthenticator(*context)
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	authorizationHeader := result.RequestHeader["Authorization"]
	if authorizationHeader != "Bearer my-access-token" {
		t.Errorf("Expected JWT bearer token in authorization header, but got: %v", authorizationHeader)
	}
}

func TestOAuthFlowIsCached(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: 200,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityUrl := identityServerFake.Start(t)

	context := createAuthContext(identityUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)
	<-resultChannel

	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
	}
	result := authenticator.Auth(context)

	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	authorizationHeader := result.RequestHeader["Authorization"]
	if authorizationHeader != "Bearer my-access-token" {
		t.Errorf("Expected JWT bearer token in authorization header, but got: %v", authorizationHeader)
	}
}

func TestProvidesCorrectPkceCodes(t *testing.T) {
	identityFake := identityServerFake{
		ResponseStatus: 200,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityUrl := identityFake.Start(t)

	context := createAuthContext(identityUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	identityFake.VerifyCodeChallenge(loginUrl.Query().Get("code_challenge"))
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
}

func TestShowsSuccessfullyLoggedInPage(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: 200,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityUrl := identityServerFake.Start(t)

	context := createAuthContext(identityUrl)
	loginUrl, _ := callAuthenticator(context)
	responseBody := performLogin(loginUrl, t)

	if !strings.Contains(responseBody, "Successfully logged in") {
		t.Errorf("Expected successfully logged in page, but got: %v", responseBody)
	}
}

func TestInvalidStateShowsErrorMessage(t *testing.T) {
	identityUrl, _ := url.Parse("http://localhost")
	context := createAuthContext(*identityUrl)
	loginUrl, _ := callAuthenticator(context)

	queryString := loginUrl.Query()
	queryString.Set("state", "invalid")
	loginUrl.RawQuery = queryString.Encode()
	responseBody := performLogin(loginUrl, t)

	if responseBody != "The query string 'state' in the redirect_url did not match" {
		t.Errorf("Expected error message that state does not match, but got: %v", responseBody)
	}
}

func TestMissingCodeShowsErrorMessage(t *testing.T) {
	identityUrl, _ := url.Parse("http://localhost")
	context := createAuthContext(*identityUrl)
	loginUrl, _ := callAuthenticator(context)

	redirectUri := loginUrl.Query().Get("redirect_uri")
	state := loginUrl.Query().Get("state")
	response, err := http.Get(redirectUri + "?code=&state=" + state)
	if err != nil {
		t.Fatalf("Unexpected error calling login url: %v", err)
	}
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Login url response body cannot be read: %v", err)
	}

	if string(responseBody) != "Could not find query string 'code' in redirect_url" {
		t.Errorf("Expected error message that state does not match, but got: %v", string(responseBody))
	}
}

func callAuthenticator(context AuthenticatorContext) (url.URL, chan AuthenticatorResult) {
	loginChan := make(chan string)
	authenticator := OAuthAuthenticator{
		Cache: cache.FileCache{},
		BrowserLauncher: NoOpBrowserLauncher{
			loginUrlChannel: loginChan,
		},
	}

	resultChannel := make(chan AuthenticatorResult)
	go func(context AuthenticatorContext) {
		result := authenticator.Auth(context)
		resultChannel <- result
	}(context)

	loginUrl := <-loginChan
	url, _ := url.Parse(loginUrl)
	return *url, resultChannel
}

func createAuthContext(identityUrl url.URL) AuthenticatorContext {
	config := map[string]interface{}{
		"clientId":    strconv.Itoa(rand.Int()),
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	request := NewAuthenticatorRequest(fmt.Sprintf("%s://%s", identityUrl.Scheme, identityUrl.Host), map[string]string{})
	context := NewAuthenticatorContext("login", config, false, false, *request)
	return *context
}

func performLogin(loginUrl url.URL, t *testing.T) string {
	redirectUri := loginUrl.Query().Get("redirect_uri")
	state := loginUrl.Query().Get("state")
	response, err := http.Get(redirectUri + "?code=testcode&state=" + state)
	if err != nil {
		t.Fatalf("Unexpected error calling login url: %v", err)
	}
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Login url response body cannot be read: %v", err)
	}
	return string(data)
}

type identityServerFake struct {
	ResponseStatus int
	ResponseBody   string
	codeChallenge  *string
}

func (i *identityServerFake) Start(t *testing.T) url.URL {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/identity_/connect/token" {
			i.handleIdentityTokenRequest(r, w)
			return
		}
	}))
	t.Cleanup(func() { srv.Close() })
	uri, _ := url.Parse(srv.URL)
	return *uri
}

func (i *identityServerFake) VerifyCodeChallenge(codeChallenge string) {
	i.codeChallenge = &codeChallenge
}

func (i *identityServerFake) handleIdentityTokenRequest(request *http.Request, response http.ResponseWriter) {
	body, _ := io.ReadAll(request.Body)
	requestBody := string(body)
	data, _ := url.ParseQuery(requestBody)

	if len(data["client_id"]) != 1 || data["client_id"][0] == "" {
		i.writeValidationErrorResponse(response, "client_id is missing")
	} else if len(data["code"]) != 1 || data["code"][0] == "" {
		i.writeValidationErrorResponse(response, "code is missing")
	} else if len(data["code_verifier"]) != 1 || data["code_verifier"][0] == "" {
		i.writeValidationErrorResponse(response, "code_verifier is missing")
	} else if len(data["redirect_uri"]) != 1 || data["redirect_uri"][0] == "" {
		i.writeValidationErrorResponse(response, "redirect_uri is missing")
	} else if len(data["grant_type"]) != 1 || data["grant_type"][0] != "authorization_code" {
		i.writeValidationErrorResponse(response, "Invalid grant_type")
	} else if i.codeChallenge != nil && !i.validPkce(data["code_verifier"][0], *i.codeChallenge) {
		i.writeValidationErrorResponse(response, "Invalid pkce")
	} else {
		response.WriteHeader(i.ResponseStatus)
		response.Write([]byte(i.ResponseBody))
	}
}

func (i identityServerFake) validPkce(codeVerifier string, expectedCodeChallenge string) bool {
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	codeChallenge = strings.ReplaceAll(codeChallenge, "+", "-")
	codeChallenge = strings.ReplaceAll(codeChallenge, "/", "_")
	return codeChallenge == expectedCodeChallenge
}

func (i identityServerFake) writeValidationErrorResponse(response http.ResponseWriter, message string) {
	response.WriteHeader(400)
	response.Write([]byte(message))
}

type NoOpBrowserLauncher struct {
	loginUrlChannel chan string
}

func (l NoOpBrowserLauncher) OpenBrowser(url string) error {
	l.loginUrlChannel <- url
	return nil
}
