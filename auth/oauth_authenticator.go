package auth

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/utils/network"
)

const RedirectUriVarName = "UIPATH_AUTH_REDIRECT_URI"

// The OAuthAuthenticator triggers the oauth authorization code flow with proof key for code exchange (PKCE).
//
// The user can login to the UiPath platform using the browser. In case the user interface is available,
// the browser is automatically launched and the oauth flow will be initiated.
// The CLI will open up a port on localhost waiting for the cloud.uipath.com platform to redirect back
// for handing over the authorization code which will be exchanged for a JWT bearer token using the
// token-endpoint from identity.
//
// There is no need to store any long-term credentials.
type OAuthAuthenticator struct {
	cache           cache.Cache
	browserLauncher BrowserLauncher
}

const refreshTokenDefaultExpiry = time.Duration(7) * 24 * time.Hour
const offlineAccessScope = "offline_access"

func (a OAuthAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(nil)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid oauth authenticator configuration: %w", err))
	}
	token, err := a.retrieveToken(*config, ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving access token: %w", err))
	}
	return *AuthenticatorSuccess(NewBearerToken(token))
}

func (a OAuthAuthenticator) retrieveToken(config oauthAuthenticatorConfig, ctx AuthenticatorContext) (string, error) {
	tokenResponse := a.getAccessTokenFromCache(config)
	if tokenResponse != nil {
		ctx.Logger.Log(fmt.Sprintf("Using existing access token from local cache which expires at %s\n", tokenResponse.ExpiresAt.UTC().Format(time.RFC3339)))
		return tokenResponse.AccessToken, nil
	}

	tokenResponse, err := a.renewAuthToken(config, ctx)
	if err != nil {
		ctx.Logger.LogError(fmt.Sprintf("Failed to renew auth token using refresh token: %v\n", err))
	} else if tokenResponse != nil {
		ctx.Logger.Log(fmt.Sprintf("Renewed access token using existing refresh token. New access token expires at %s\n", tokenResponse.ExpiresAt.UTC().Format(time.RFC3339)))
		return tokenResponse.AccessToken, nil
	}

	ctx.Logger.Log("No access token or refresh token available. Starting login flow...\n")

	secretGenerator := newSecretGenerator()
	codeVerifier, codeChallenge := secretGenerator.GeneratePkce()
	state := secretGenerator.GenerateState()
	code, err := a.login(config, state, codeChallenge)
	if err != nil {
		return "", err
	}

	identityClient := newIdentityClient(ctx.Logger)
	tokenRequest := newAuthorizationCodeTokenRequest(
		config.IdentityUri,
		config.ClientId,
		config.ClientSecret,
		code,
		codeVerifier,
		config.RedirectUrl.String(),
		a.networkSettings(ctx))
	tokenResponse, err = identityClient.GetToken(*tokenRequest)
	if err != nil {
		return "", err
	}

	a.updateTokenResponseCache(config, tokenResponse)
	return tokenResponse.AccessToken, nil
}

func (a OAuthAuthenticator) renewAuthToken(config oauthAuthenticatorConfig, ctx AuthenticatorContext) (*tokenResponse, error) {
	refreshTokenCacheKey := a.refreshTokenCacheKey(config)
	refreshToken, _ := a.cache.Get(refreshTokenCacheKey)
	if !config.OfflineAccess || refreshToken == "" {
		return nil, nil
	}

	refreshTokenRequest := newRefreshTokenRequest(
		config.IdentityUri,
		config.ClientId,
		config.ClientSecret,
		refreshToken,
		a.networkSettings(ctx))

	identityClient := newIdentityClient(ctx.Logger)
	tokenResponse, err := identityClient.GetToken(*refreshTokenRequest)
	if err != nil {
		return nil, err
	}

	a.updateTokenResponseCache(config, tokenResponse)
	return tokenResponse, nil
}

func (a OAuthAuthenticator) getAccessTokenFromCache(config oauthAuthenticatorConfig) *tokenResponse {
	cacheKey := a.accessTokenCacheKey(config)
	token, expiresAt := a.cache.Get(cacheKey)
	if token == "" {
		return nil
	}
	return newTokenResponse(token, expiresAt, nil)
}

func (a OAuthAuthenticator) updateTokenResponseCache(config oauthAuthenticatorConfig, tokenResponse *tokenResponse) {
	cacheKey := a.accessTokenCacheKey(config)
	a.cache.Set(cacheKey, tokenResponse.AccessToken, tokenResponse.ExpiresAt.Add(-TokenExpiryGracePeriod))

	if tokenResponse.RefreshToken != nil {
		refreshTokenCacheKey := a.refreshTokenCacheKey(config)
		a.cache.Set(refreshTokenCacheKey, *tokenResponse.RefreshToken, time.Now().UTC().Add(refreshTokenDefaultExpiry-TokenExpiryGracePeriod))
	}
}

func (a OAuthAuthenticator) accessTokenCacheKey(config oauthAuthenticatorConfig) string {
	identityBaseUri := config.IdentityUri
	return fmt.Sprintf("oauthaccesstoken|%s|%s|%s|%s|%s", identityBaseUri.Scheme, identityBaseUri.Hostname(), config.ClientId, config.ClientSecret, config.Scopes)
}

func (a OAuthAuthenticator) refreshTokenCacheKey(config oauthAuthenticatorConfig) string {
	identityBaseUri := config.IdentityUri
	return fmt.Sprintf("oauthrefreshtoken|%s|%s|%s|%s|%s", identityBaseUri.Scheme, identityBaseUri.Hostname(), config.ClientId, config.ClientSecret, config.Scopes)
}

func (a OAuthAuthenticator) login(config oauthAuthenticatorConfig, state string, codeChallenge string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(120)*time.Second)
	defer cancel()

	var code string
	var err error

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code = query.Get("code")
		if code == "" {
			err = errors.New("Could not find query string 'code' in redirect_url")
			a.writeErrorPage(w, err)
		} else if query.Get("state") != state {
			err = errors.New("The query string 'state' in the redirect_url did not match")
			a.writeErrorPage(w, err)
		} else {
			a.writeHtmlPage(w, LOGGED_IN_PAGE_HTML)
		}
		cancel()
	})
	listener, err := net.Listen("tcp", config.RedirectUrl.Host)
	if err != nil {
		return "", fmt.Errorf("Error starting listener on address %s and wait for oauth redirect: %w", config.RedirectUrl.Host, err)
	}
	defer func() { _ = listener.Close() }()

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 30 * time.Second,
	}
	defer func() { _ = server.Close() }()

	go func(listener net.Listener) {
		listenErr := server.Serve(listener)
		if listenErr != nil {
			err = fmt.Errorf("Error starting server on address %s and wait for oauth redirect: %w", config.RedirectUrl.Host, listenErr)
		}
		cancel()
	}(listener)

	port := listener.Addr().(*net.TCPAddr).Port
	identityBaseUri := config.IdentityUri
	redirectUri := fmt.Sprintf("%s://%s:%d", config.RedirectUrl.Scheme, config.RedirectUrl.Hostname(), port)
	loginUrl := fmt.Sprintf("%s/connect/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&code_challenge=%s&code_challenge_method=S256&state=%s",
		identityBaseUri.String(),
		url.QueryEscape(config.ClientId),
		url.QueryEscape(redirectUri),
		url.QueryEscape(config.Scopes),
		url.QueryEscape(codeChallenge),
		url.QueryEscape(state))

	go func(url string) {
		err := a.browserLauncher.Open(url)
		if err != nil {
			a.showBrowserLink(url)
		}
	}(loginUrl)

	<-ctx.Done()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", errors.New("OAuth Login expired")
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

func (a OAuthAuthenticator) enabled(ctx AuthenticatorContext) bool {
	clientIdSet := os.Getenv(ClientIdEnvVarName) != "" || ctx.Config["clientId"] != nil
	redirectUriSet := os.Getenv(RedirectUriVarName) != "" || ctx.Config["redirectUri"] != nil
	scopesSet := os.Getenv(ScopesEnvVarName) != "" || ctx.Config["scopes"] != nil
	return clientIdSet && redirectUriSet && scopesSet
}

func (a OAuthAuthenticator) getConfig(ctx AuthenticatorContext) (*oauthAuthenticatorConfig, error) {
	clientId, err := a.parseRequiredString(ctx.Config, "clientId", ClientIdEnvVarName)
	if err != nil {
		return nil, err
	}
	clientSecret, _ := a.parseString(ctx.Config, "clientSecret", ClientSecretEnvVarName)
	redirectUri, err := a.parseRequiredString(ctx.Config, "redirectUri", RedirectUriVarName)
	if err != nil {
		return nil, err
	}
	parsedRedirectUri, err := url.Parse(redirectUri)
	if err != nil {
		return nil, err
	}
	scopes, err := a.parseRequiredString(ctx.Config, "scopes", ScopesEnvVarName)
	if err != nil {
		return nil, err
	}
	offlineAccess, err := a.parseBool(ctx.Config, "offlineAccess", true)
	if err != nil {
		return nil, err
	}
	if offlineAccess {
		scopes = scopes + " " + offlineAccessScope
	}
	return newOAuthAuthenticatorConfig(clientId, clientSecret, *parsedRedirectUri, scopes, ctx.IdentityUri, offlineAccess), nil
}

func (a OAuthAuthenticator) parseString(config map[string]interface{}, name string, envVarName string) (string, error) {
	envVarValue := os.Getenv(envVarName)
	if envVarValue != "" {
		return envVarValue, nil
	}
	value := config[name]
	result, valid := value.(string)
	if value != nil && !valid {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func (a OAuthAuthenticator) parseBool(config map[string]interface{}, name string, defaultValue bool) (bool, error) {
	value := config[name]
	if value == nil {
		return defaultValue, nil
	}

	result, valid := value.(bool)
	if value != nil && !valid {
		return false, fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func (a OAuthAuthenticator) parseRequiredString(config map[string]interface{}, name string, envVarName string) (string, error) {
	envVarValue := os.Getenv(envVarName)
	if envVarValue != "" {
		return envVarValue, nil
	}
	value := config[name]
	result, valid := value.(string)
	if !valid || result == "" {
		return "", fmt.Errorf("Invalid value for %s: '%v'", name, value)
	}
	return result, nil
}

func (a OAuthAuthenticator) showBrowserLink(url string) {
	fmt.Fprintln(os.Stderr, "Go to URL and perform login:\n"+url)
}

func (a OAuthAuthenticator) writeErrorPage(w http.ResponseWriter, err error) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(err.Error()))
}

func (a OAuthAuthenticator) writeHtmlPage(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html")
	_, _ = w.Write([]byte(html))
}

func (a OAuthAuthenticator) networkSettings(ctx AuthenticatorContext) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		ctx.Debug,
		ctx.OperationId,
		map[string]string{},
		GetTokenTimeout,
		GetTokenMaxAttempts,
		ctx.Insecure,
	)
}

func NewOAuthAuthenticator(cache cache.Cache, browserLauncher BrowserLauncher) *OAuthAuthenticator {
	return &OAuthAuthenticator{cache, browserLauncher}
}
