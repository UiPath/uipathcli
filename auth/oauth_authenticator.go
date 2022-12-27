package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/UiPath/uipathcli/cache"
)

type OAuthAuthenticator struct {
	Cache cache.Cache
}

func (a OAuthAuthenticator) Auth(ctx AuthenticatorContext) AuthenticatorResult {
	if !a.enabled(ctx) {
		return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
	}
	config, err := a.getConfig(ctx)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid oauth authenticator configuration: %v", err))
	}
	requestUrl, err := url.Parse(ctx.Request.URL)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid request url '%s': %v", ctx.Request.URL, err))
	}
	identityBaseUri, err := url.Parse(fmt.Sprintf("%s://%s/identity_", requestUrl.Scheme, requestUrl.Host))
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Invalid identity url '%s': %v", ctx.Request.URL, err))
	}
	token, err := a.retrieveToken(*identityBaseUri, *config, ctx.Insecure)
	if err != nil {
		return *AuthenticatorError(fmt.Errorf("Error retrieving access token: %v", err))
	}
	ctx.Request.Header["Authorization"] = "Bearer " + token
	return *AuthenticatorSuccess(ctx.Request.Header, ctx.Config)
}

func (a OAuthAuthenticator) retrieveToken(identityBaseUri url.URL, config OAuthAuthenticatorConfig, insecure bool) (string, error) {
	cacheKey := fmt.Sprintf("oauthtoken|%s|%s|%s|%s", identityBaseUri.Scheme, identityBaseUri.Hostname(), config.ClientId, config.Scopes)
	token, _ := a.Cache.Get(cacheKey)
	if token != "" {
		return token, nil
	}

	secretGenerator := SecretGenerator{}
	codeVerifier, codeChallenge := secretGenerator.GeneratePkce()
	state := secretGenerator.GenerateState()
	code, err := a.login(identityBaseUri, config, state, codeChallenge)
	if err != nil {
		return "", err
	}

	identityClient := identityClient(a)
	tokenRequest := newAuthorizationCodeTokenRequest(
		identityBaseUri,
		config.ClientId,
		code,
		codeVerifier,
		config.RedirectUrl.String(),
		insecure)
	tokenResponse, err := identityClient.GetToken(*tokenRequest)
	if err != nil {
		return "", err
	}
	a.Cache.Set(cacheKey, tokenResponse.AccessToken, tokenResponse.ExpiresIn)
	return tokenResponse.AccessToken, nil
}

func (a OAuthAuthenticator) login(identityBaseUri url.URL, config OAuthAuthenticatorConfig, state string, codeChallenge string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	var code string
	var err error

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		code = query.Get("code")
		if code == "" {
			err = fmt.Errorf("Could not find query string 'code' in redirect_url")
			a.writeErrorPage(w, err)
		} else if query.Get("state") != state {
			err = fmt.Errorf("The query string 'state' in the redirect_url did not match")
			a.writeErrorPage(w, err)
		} else {
			a.writeHtmlPage(w, LOGGED_IN_PAGE_HTML)
		}
		cancel()
	})
	port := config.RedirectUrl.Port()
	server := &http.Server{
		Addr:    ":" + port,
		Handler: mux,
	}
	go func() {
		listenErr := server.ListenAndServe()
		if listenErr != nil {
			err = fmt.Errorf("Error starting server on port %s and wait for oauth redirect: %v", port, listenErr)
		}
		cancel()
	}()

	loginUrl := fmt.Sprintf("%s/connect/authorize?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&code_challenge=%s&code_challenge_method=S256&state=%s",
		identityBaseUri.String(),
		url.QueryEscape(config.ClientId),
		url.QueryEscape(config.RedirectUrl.String()),
		url.QueryEscape(config.Scopes),
		url.QueryEscape(codeChallenge),
		url.QueryEscape(state))

	go func(url string) {
		a.openBrowser(url)
	}(loginUrl)

	<-ctx.Done()

	if errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return "", fmt.Errorf("OAuth Login expired")
	}
	if err != nil {
		return "", err
	}
	return code, nil
}

func (a OAuthAuthenticator) enabled(ctx AuthenticatorContext) bool {
	return ctx.Config["clientId"] != nil && ctx.Config["redirectUri"] != nil && ctx.Config["scopes"] != nil
}

func (a OAuthAuthenticator) getConfig(ctx AuthenticatorContext) (*OAuthAuthenticatorConfig, error) {
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
	return NewOAuthAuthenticatorConfig(clientId, *parsedRedirectUri, scopes), nil
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

func (a OAuthAuthenticator) openBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		a.showBrowserLink(url)
		return fmt.Errorf("Platform not supported: %s", runtime.GOOS)
	}

	err := cmd.Start()
	if err != nil {
		a.showBrowserLink(url)
		return err
	}
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			a.showBrowserLink(url)
		}
		return err
	case <-time.After(5 * time.Second):
		return nil
	}
}

func (a OAuthAuthenticator) writeErrorPage(w http.ResponseWriter, err error) {
	w.Header().Add("content-type", "text/html")
	w.Write([]byte(err.Error()))
}

func (a OAuthAuthenticator) writeHtmlPage(w http.ResponseWriter, html string) {
	w.Header().Add("content-type", "text/html")
	w.Write([]byte(html))
}
