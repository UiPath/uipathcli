package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/cache"
	"github.com/UiPath/uipathcli/utils/network"
)

type IdentityClient struct {
	cache    cache.Cache
	baseUri  string
	debug    bool
	settings network.HttpClientSettings
	logger   log.Logger
}

func (c IdentityClient) GetToken(tokenRequest TokenRequest) (*TokenResponse, error) {
	form := url.Values{}
	form.Add("grant_type", tokenRequest.GrantType)
	form.Add("scope", tokenRequest.Scopes)
	form.Add("client_id", tokenRequest.ClientId)
	form.Add("client_secret", tokenRequest.ClientSecret)
	form.Add("code", tokenRequest.Code)
	form.Add("code_verifier", tokenRequest.CodeVerifier)
	form.Add("redirect_uri", tokenRequest.RedirectUri)
	for key, value := range tokenRequest.Properties {
		form.Add(key, value)
	}
	baseUri, err := url.Parse(c.baseUri)
	if err != nil {
		return nil, fmt.Errorf("Invalid uri '%s' for identity server: %v", c.baseUri, err)
	}

	cacheKey := c.cacheKey(*baseUri, tokenRequest)
	token, expiresIn := c.cache.Get(cacheKey)
	if token != "" {
		return NewTokenResponse(token, expiresIn), nil
	}

	response, err := c.retrieveToken(*baseUri, form)
	if err != nil {
		return nil, err
	}
	c.cache.Set(cacheKey, response.AccessToken, response.ExpiresIn)
	return response, nil
}

func (c IdentityClient) retrieveToken(baseUri url.URL, form url.Values) (*TokenResponse, error) {
	uri := baseUri.JoinPath("/connect/token")
	header := http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	request := network.NewHttpPostRequest(uri.String(), nil, header, strings.NewReader(form.Encode()), -1)

	client := network.NewHttpClient(c.logger, c.debug, c.settings)
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Token service returned status code '%v' and body '%v'", response.StatusCode, string(bytes))
	}

	var result tokenResponseJson
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, fmt.Errorf("Error parsing json response: %w", err)
	}
	return NewTokenResponse(result.AccessToken, result.ExpiresIn), nil
}

func (c IdentityClient) cacheKey(baseUri url.URL, tokenRequest TokenRequest) string {
	return fmt.Sprintf("token|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		baseUri.Scheme,
		baseUri.Hostname(),
		tokenRequest.GrantType,
		tokenRequest.Scopes,
		tokenRequest.ClientId,
		tokenRequest.ClientSecret,
		tokenRequest.Code,
		tokenRequest.CodeVerifier,
		tokenRequest.RedirectUri,
		c.cacheKeyProperties(tokenRequest.Properties))
}

func (c IdentityClient) cacheKeyProperties(properties map[string]string) string {
	values := []string{}
	for key, value := range properties {
		values = append(values, key+"="+value)
	}
	return strings.Join(values, ",")
}

func NewIdentityClient(cache cache.Cache, baseUri string, debug bool, settings network.HttpClientSettings, logger log.Logger) *IdentityClient {
	return &IdentityClient{cache, baseUri, debug, settings, logger}
}

type tokenResponseJson struct {
	AccessToken string  `json:"access_token"`
	ExpiresIn   float32 `json:"expires_in"`
}
