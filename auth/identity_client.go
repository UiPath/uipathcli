package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/utils/network"
)

type identityClient struct {
	cache cache.Cache
}

const TokenRoute = "/connect/token"

func (c identityClient) GetToken(tokenRequest tokenRequest) (*tokenResponse, error) {
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

	cacheKey := c.cacheKey(tokenRequest)
	token, expiresIn := c.cache.Get(cacheKey)
	if token != "" {
		return newTokenResponse(token, expiresIn), nil
	}

	response, err := c.retrieveToken(tokenRequest.BaseUri, form, tokenRequest.OperationId, tokenRequest.Insecure)
	if err != nil {
		return nil, err
	}
	c.cache.Set(cacheKey, response.AccessToken, response.ExpiresIn)
	return response, nil
}

func (c identityClient) retrieveToken(baseUri url.URL, form url.Values, operationId string, insecure bool) (*tokenResponse, error) {
	uri := baseUri.JoinPath(TokenRoute)
	header := http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	request := network.NewHttpPostRequest(uri.String(), nil, header, strings.NewReader(form.Encode()), -1)

	clientSettings := network.NewHttpClientSettings(false, operationId, map[string]string{}, time.Duration(60)*time.Second, 3, insecure)
	client := network.NewHttpClient(nil, *clientSettings)
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

	var result tokenResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, fmt.Errorf("Error parsing json response: %w", err)
	}
	return &result, nil
}

func (c identityClient) cacheKey(tokenRequest tokenRequest) string {
	return fmt.Sprintf("token|%s|%s|%s|%s|%s|%s|%s|%s|%s|%s",
		tokenRequest.BaseUri.Scheme,
		tokenRequest.BaseUri.Hostname(),
		tokenRequest.GrantType,
		tokenRequest.Scopes,
		tokenRequest.ClientId,
		tokenRequest.ClientSecret,
		tokenRequest.Code,
		tokenRequest.CodeVerifier,
		tokenRequest.RedirectUri,
		c.cacheKeyProperties(tokenRequest.Properties))
}

func (c identityClient) cacheKeyProperties(properties map[string]string) string {
	values := []string{}
	for key, value := range properties {
		values = append(values, key+"="+value)
	}
	return strings.Join(values, ",")
}

func newIdentityClient(cache cache.Cache) *identityClient {
	return &identityClient{cache}
}
