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
)

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

func (a OAuthAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(nil)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid oauth authenticator configuration: %w", err))
	}
	token, err := a.retrieveToken(config.IdentityUri, *config, ctx.OperationId, ctx.Insecure)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving access token: %w", err))
	}
	return *AuthenticatorSuccess(NewBearerToken(token))
}

func (a OAuthAuthenticator) retrieveToken(identityBaseUri url.URL, config oauthAuthenticatorConfig, operationId string, insecure bool) (string, error) {
	cacheKey := fmt.Sprintf("oauthtoken|%s|%s|%s|%s", identityBaseUri.Scheme, identityBaseUri.Hostname(), config.ClientId, config.Scopes)
	token, _ := a.cache.Get(cacheKey)
	if token != "" {
		return token, nil
	}

	secretGenerator := newSecretGenerator()
	codeVerifier, codeChallenge := secretGenerator.GeneratePkce()
	state := secretGenerator.GenerateState()
	code, err := a.login(identityBaseUri, config, state, codeChallenge)
	if err != nil {
		return "", err
	}

	identityClient := newIdentityClient(a.cache)
	tokenRequest := newAuthorizationCodeTokenRequest(
		identityBaseUri,
		config.ClientId,
		code,
		codeVerifier,
		config.RedirectUrl.String(),
		operationId,
		insecure)
	tokenResponse, err := identityClient.GetToken(*tokenRequest)
	if err != nil {
		return "", err
	}
	a.cache.Set(cacheKey, tokenResponse.AccessToken, tokenResponse.ExpiresIn)
	return tokenResponse.AccessToken, nil
}

func (a OAuthAuthenticator) login(identityBaseUri url.URL, config oauthAuthenticatorConfig, state string, codeChallenge string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
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
	return ctx.Config["clientId"] != nil && ctx.Config["redirectUri"] != nil && ctx.Config["scopes"] != nil
}

func (a OAuthAuthenticator) getConfig(ctx AuthenticatorContext) (*oauthAuthenticatorConfig, error) {
	clientId, err := a.parseRequiredString(ctx.Config, "clientId")
	if err != nil {
		return nil, err
	}
	redirectUri, err := a.parseRequiredString(ctx.Config, "redirectUri")
	if err != nil {
		return nil, err
	}
	parsedRedirectUri, err := url.Parse(redirectUri)
	if err != nil {
		return nil, err
	}
	scopes, err := a.parseRequiredString(ctx.Config, "scopes")
	if err != nil {
		return nil, err
	}
	return newOAuthAuthenticatorConfig(clientId, *parsedRedirectUri, scopes, ctx.IdentityUri), nil
}

func (a OAuthAuthenticator) parseRequiredString(config map[string]interface{}, name string) (string, error) {
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

func NewOAuthAuthenticator(cache cache.Cache, browserLauncher BrowserLauncher) *OAuthAuthenticator {
	return &OAuthAuthenticator{cache, browserLauncher}
}
