package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"math"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestOAuthAuthenticatorNotEnabled(t *testing.T) {
	config := map[string]interface{}{
		"clientId": "my-client-id",
		"scopes":   "OR.Users",
	}
	url, _ := url.Parse("http:/localhost")
	context := createAuthContext(*url, config, false, io.Discard)

	authenticator := NewOAuthAuthenticator(cache.NewFileCache(), *NewBrowserLauncher())
	result := authenticator.Auth(context)
	if result.Error != "" {
		t.Errorf("Expected no error when oauth flow is skipped, but got: %v", result.Error)
	}
	if result.Token != nil {
		t.Errorf("Expected auth token to be empty, but got: %v", result.Token)
	}
}

func TestOAuthAuthenticatorInvalidConfig(t *testing.T) {
	config := map[string]interface{}{
		"clientId":    1,
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	url, _ := url.Parse("http:/localhost")
	context := createAuthContext(*url, config, false, io.Discard)

	authenticator := NewOAuthAuthenticator(cache.NewFileCache(), *NewBrowserLauncher())
	result := authenticator.Auth(context)
	if result.Error != "Invalid oauth authenticator configuration: Invalid value for clientId: '1'" {
		t.Errorf("Expected error with invalid config, but got: %v", result.Error)
	}
}

func TestOAuthFlowIdentityFails(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: http.StatusBadRequest,
		ResponseBody:   "Invalid token request",
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityBaseUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "Error retrieving access token: Token service returned status code '400' and body 'Invalid token request'" {
		t.Errorf("Expected error when identity call fails, but got: %v", result.Error)
	}
}

func TestNonConfidentialOAuthFlowSuccessful(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityBaseUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	if result.Token.Type != "Bearer" {
		t.Errorf("Expected JWT bearer token, but got: %v", result.Token.Type)
	}
	if result.Token.Value != "my-access-token" {
		t.Errorf("Expected value for JWT bearer token, but got: %v", result.Token.Value)
	}
}

func TestConfidentialOAuthFlowSuccessful(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityBaseUrl := identityServerFake.Start(t, true)

	context := createConfidentialAuthContext(identityBaseUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)

	result := <-resultChannel
	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	if result.Token.Type != "Bearer" {
		t.Errorf("Expected JWT bearer token, but got: %v", result.Token.Type)
	}
	if result.Token.Value != "my-access-token" {
		t.Errorf("Expected value for JWT bearer token, but got: %v", result.Token.Value)
	}
}

func TestOAuthFlowIsCached(t *testing.T) {
	calls := 0
	identityServerFake := identityServerFake{
		ResponseHandler: func(data map[string]string) (int, string) {
			calls++
			if calls > 1 {
				return http.StatusInternalServerError, "There was an error"
			}
			body := `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`
			return http.StatusOK, body
		},
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityBaseUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)
	<-resultChannel

	authenticator := NewOAuthAuthenticator(cache.NewFileCache(), *NewBrowserLauncher())
	result := authenticator.Auth(context)

	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	if result.Token.Type != "Bearer" {
		t.Errorf("Expected JWT bearer token, but got: %v", result.Token.Type)
	}
	if result.Token.Value != "my-access-token" {
		t.Errorf("Expected value for JWT bearer token, but got: %v", result.Token.Value)
	}
	if calls != 1 {
		t.Errorf("Expected only one call to identity server, but got: %d", calls)
	}
}

func TestOAuthFlowUsesRefreshToken(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseHandler: func(data map[string]string) (int, string) {
			refreshToken := data["refresh_token"]
			if refreshToken == "my-refresh-token" {
				body := `{"access_token": "my-renewed-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`
				return http.StatusOK, body
			}
			body := `{"access_token": "my-access-token", "expires_in": 10, "token_type": "Bearer", "scope": "OR.Users", "refresh_token": "my-refresh-token"}`
			return http.StatusOK, body
		},
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityBaseUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)
	<-resultChannel

	authenticator := NewOAuthAuthenticator(cache.NewFileCache(), *NewBrowserLauncher())
	result := authenticator.Auth(context)

	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	if result.Token.Type != "Bearer" {
		t.Errorf("Expected JWT bearer token, but got: %v", result.Token.Type)
	}
	if result.Token.Value != "my-renewed-access-token" {
		t.Errorf("Expected renewed value for JWT bearer token, but got: %v", result.Token.Value)
	}
	scopes := loginUrl.Query()["scope"][0]
	if scopes != "OR.Users offline_access" {
		t.Errorf("Expected offline_access scope, but got: %v", scopes)
	}
}

func TestOAuthFlowDoesNotUseRefreshTokenWhenDisabled(t *testing.T) {
	refreshTokenCalled := false
	identityServerFake := identityServerFake{
		ResponseHandler: func(data map[string]string) (int, string) {
			if data["refresh_token"] != "" {
				fmt.Println(data)
				refreshTokenCalled = true
			}
			body := `{"access_token": "my-access-token", "expires_in": 10, "token_type": "Bearer", "scope": "OR.Users", "refresh_token": "my-refresh-token"}`
			return http.StatusOK, body
		},
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	config := map[string]interface{}{
		"clientId":      random(),
		"redirectUri":   "http://localhost:0",
		"scopes":        "OR.Users",
		"offlineAccess": false,
	}
	context := createAuthContext(identityBaseUrl, config, false, io.Discard)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)
	<-resultChannel

	loginUrl, resultChannel = callAuthenticator(context)
	performLogin(loginUrl, t)
	result := <-resultChannel

	if result.Error != "" {
		t.Errorf("Expected no error when performing oauth flow, but got: %v", result.Error)
	}
	if refreshTokenCalled {
		t.Errorf("Expected to not use refresh token, but it did")
	}
	scopes := loginUrl.Query()["scope"][0]
	if scopes != "OR.Users" {
		t.Errorf("Expected no offline_access scope, but got: %v", scopes)
	}
}

func TestOAuthFlowIgnoresInvalidRefreshToken(t *testing.T) {
	calls := 0
	calledWithRefreshToken := false
	identityServerFake := identityServerFake{
		ResponseHandler: func(data map[string]string) (int, string) {
			calls++
			refreshToken := data["refresh_token"]
			if refreshToken == "my-refresh-token" {
				calledWithRefreshToken = true
				return http.StatusBadRequest, "Refresh token is invalid"
			}
			if calls == 1 {
				body := `{"access_token": "my-access-token", "expires_in": 10, "token_type": "Bearer", "scope": "OR.Users", "refresh_token": "my-refresh-token"}`
				return http.StatusOK, body
			}
			body := `{"access_token": "my-second-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users", "refresh_token": "my-refresh-token"}`
			return http.StatusOK, body
		},
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityBaseUrl)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)
	<-resultChannel

	loginUrl, resultChannel = callAuthenticator(context)
	performLogin(loginUrl, t)
	result := <-resultChannel

	if result.Error != "" {
		t.Errorf("Expected no error with invalid refresh token, but got: %v", result.Error)
	}
	if result.Token.Value != "my-second-access-token" {
		t.Errorf("Expected new JWT bearer token, but got: %v", result.Token.Value)
	}
	if !calledWithRefreshToken {
		t.Errorf("Expected to use refresh token, but did not")
	}
}

func TestProvidesCorrectPkceCodes(t *testing.T) {
	identityFake := identityServerFake{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityUrl := identityFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityUrl)
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
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users"}`,
	}
	identityBaseUrl := identityServerFake.Start(t, false)

	context := createNonConfidentialAuthContext(identityBaseUrl)
	loginUrl, _ := callAuthenticator(context)
	responseBody := performLogin(loginUrl, t)

	if !strings.Contains(responseBody, "Successfully logged in") {
		t.Errorf("Expected successfully logged in page, but got: %v", responseBody)
	}
}

func TestInvalidStateShowsErrorMessage(t *testing.T) {
	identityUrl, _ := url.Parse("http://localhost")
	context := createNonConfidentialAuthContext(*identityUrl)
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
	ctx := createNonConfidentialAuthContext(*identityUrl)
	loginUrl, _ := callAuthenticator(ctx)

	redirectUri := loginUrl.Query().Get("redirect_uri")
	state := loginUrl.Query().Get("state")
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, redirectUri+"?code=&state="+state, nil)
	if err != nil {
		t.Fatalf("Unexpected error calling login url: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("Unexpected error calling login url: %v", err)
	}
	defer func() { _ = response.Body.Close() }()

	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Login url response body cannot be read: %v", err)
	}

	if string(responseBody) != "Could not find query string 'code' in redirect_url" {
		t.Errorf("Expected error message that state does not match, but got: %v", string(responseBody))
	}
}

func TestOAuthFlowRedactsClientSecretAndRefreshToken(t *testing.T) {
	identityServerFake := identityServerFake{
		ResponseStatus: http.StatusOK,
		ResponseBody:   `{"access_token": "my-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Users", "refresh_token": "my-refresh-token"}`,
	}
	identityBaseUrl := identityServerFake.Start(t, true)

	var writer bytes.Buffer

	config := map[string]interface{}{
		"clientId":     random(),
		"clientSecret": random(),
		"redirectUri":  "http://localhost:0",
		"scopes":       "OR.Users",
	}
	context := createAuthContext(identityBaseUrl, config, true, &writer)
	loginUrl, resultChannel := callAuthenticator(context)
	performLogin(loginUrl, t)
	<-resultChannel

	output := writer.String()
	if !strings.Contains(output, "client_secret=%2A%2Aredacted%2A%2A") {
		t.Errorf("Expected client_secret to be redacted, but got: %v", output)
	}
	if !strings.Contains(output, `{"access_token":"my-access-token","expires_in":3600,"token_type":"Bearer","scope":"OR.Users","refresh_token":"**redacted**"}`) {
		t.Errorf("Expected refresh_token to be redacted, but got: %v", output)
	}
}

func callAuthenticator(context AuthenticatorContext) (url.URL, chan AuthenticatorResult) {
	loginChan := make(chan string)
	authenticator := NewOAuthAuthenticator(cache.NewFileCache(), BrowserLauncher{
		Exec: process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
			switch runtime.GOOS {
			case "windows":
				loginChan <- args[1]
			default:
				loginChan <- args[0]
			}
		}),
	})

	resultChannel := make(chan AuthenticatorResult)
	go func(context AuthenticatorContext) {
		result := authenticator.Auth(context)
		resultChannel <- result
	}(context)

	loginUrl := <-loginChan
	url, _ := url.Parse(loginUrl)
	return *url, resultChannel
}

func createConfidentialAuthContext(baseUrl url.URL) AuthenticatorContext {
	config := map[string]interface{}{
		"clientId":     random(),
		"clientSecret": random(),
		"redirectUri":  "http://localhost:0",
		"scopes":       "OR.Users",
	}
	return createAuthContext(baseUrl, config, false, io.Discard)
}

func createNonConfidentialAuthContext(baseUrl url.URL) AuthenticatorContext {
	config := map[string]interface{}{
		"clientId":    random(),
		"redirectUri": "http://localhost:0",
		"scopes":      "OR.Users",
	}
	return createAuthContext(baseUrl, config, false, io.Discard)
}

func createAuthContext(baseUrl url.URL, config map[string]interface{}, debug bool, writer io.Writer) AuthenticatorContext {
	identityUrl := createIdentityUrl(baseUrl.Host)
	request := NewAuthenticatorRequest(fmt.Sprintf("%s://%s", baseUrl.Scheme, baseUrl.Host), map[string]string{})
	context := NewAuthenticatorContext(config, identityUrl, "d7b087788be2154da3ad9d6bc14588f4", false, debug, *request, log.NewDebugLogger(writer))
	return *context
}

func createIdentityUrl(hostName string) url.URL {
	if hostName == "" {
		hostName = "localhost"
	}
	identityUrl, _ := url.Parse(fmt.Sprintf("http://%s/identity_", hostName))
	return *identityUrl
}

func performLogin(loginUrl url.URL, t *testing.T) string {
	redirectUri := loginUrl.Query().Get("redirect_uri")
	state := loginUrl.Query().Get("state")
	request, err := http.NewRequestWithContext(t.Context(), http.MethodGet, redirectUri+"?code=testcode&state="+state, nil)
	if err != nil {
		t.Fatalf("Unexpected error calling login url: %v", err)
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("Unexpected error calling login url: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("Login url response body cannot be read: %v", err)
	}
	return string(data)
}

func random() string {
	value, err := rand.Int(rand.Reader, big.NewInt(math.MaxInt64))
	if err != nil {
		panic(fmt.Errorf("Unexpected error generating new client id: %w", err))
	}
	return value.String()
}

type identityServerFake struct {
	ResponseStatus  int
	ResponseBody    string
	ResponseHandler func(map[string]string) (int, string)
	codeChallenge   *string
}

func (i *identityServerFake) Start(t *testing.T, confidential bool) url.URL {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.String() == "/identity_/connect/token" {
			i.handleIdentityTokenRequest(confidential, r, w)
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

func (i *identityServerFake) handleIdentityTokenRequest(confidential bool, request *http.Request, response http.ResponseWriter) {
	body, _ := io.ReadAll(request.Body)
	requestBody := string(body)
	data, _ := url.ParseQuery(requestBody)

	if i.ResponseHandler != nil {
		responseStatus, responseBody := i.ResponseHandler(i.convertToMap(data))
		response.WriteHeader(responseStatus)
		_, _ = response.Write([]byte(responseBody))
		return
	}

	if i.isEmpty(data, "client_id") {
		i.writeValidationErrorResponse(response, "client_id is missing")
	} else if confidential && i.isEmpty(data, "client_secret") {
		i.writeValidationErrorResponse(response, "client_secret is missing")
	} else if i.isEmpty(data, "code") {
		i.writeValidationErrorResponse(response, "code is missing")
	} else if i.isEmpty(data, "code_verifier") {
		i.writeValidationErrorResponse(response, "code_verifier is missing")
	} else if i.isEmpty(data, "redirect_uri") {
		i.writeValidationErrorResponse(response, "redirect_uri is missing")
	} else if i.isEmpty(data, "grant_type") || data["grant_type"][0] != "authorization_code" {
		i.writeValidationErrorResponse(response, "Invalid grant_type")
	} else if i.codeChallenge != nil && !i.validPkce(data["code_verifier"][0], *i.codeChallenge) {
		i.writeValidationErrorResponse(response, "Invalid pkce")
	} else {
		response.WriteHeader(i.ResponseStatus)
		_, _ = response.Write([]byte(i.ResponseBody))
	}
}

func (i *identityServerFake) convertToMap(values url.Values) map[string]string {
	result := map[string]string{}
	for key, value := range values {
		if len(value) > 0 {
			result[key] = value[0]
		}
	}
	return result
}

func (i *identityServerFake) isEmpty(values url.Values, key string) bool {
	return len(values[key]) != 1 || values[key][0] == ""
}

func (i *identityServerFake) validPkce(codeVerifier string, expectedCodeChallenge string) bool {
	hash := sha256.Sum256([]byte(codeVerifier))
	codeChallenge := base64.StdEncoding.WithPadding(base64.NoPadding).EncodeToString(hash[:])
	codeChallenge = strings.ReplaceAll(codeChallenge, "+", "-")
	codeChallenge = strings.ReplaceAll(codeChallenge, "/", "_")
	return codeChallenge == expectedCodeChallenge
}

func (i *identityServerFake) writeValidationErrorResponse(response http.ResponseWriter, message string) {
	response.WriteHeader(http.StatusBadRequest)
	_, _ = response.Write([]byte(message))
}
