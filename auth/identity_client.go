package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/utils"
)

type identityClient struct {
	Cache cache.Cache
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
	token, expiresIn := c.Cache.Get(cacheKey)
	if token != "" {
		return newTokenResponse(token, expiresIn), nil
	}

	response, err := c.retrieveToken(tokenRequest.BaseUri, form, tokenRequest.Insecure)
	if err != nil {
		return nil, err
	}
	c.Cache.Set(cacheKey, response.AccessToken, response.ExpiresIn)
	return response, nil
}

func (c identityClient) retrieveToken(baseUri url.URL, form url.Values, insecure bool) (*tokenResponse, error) {
	var response *tokenResponse
	var err error
	utils.Retry(func() error {
		response, err = c.send(baseUri, form, insecure)
		return err
	})
	return response, err
}

func (c identityClient) send(baseUri url.URL, form url.Values, insecure bool) (*tokenResponse, error) {
	uri := baseUri.JoinPath(TokenRoute)
	request, err := http.NewRequest("POST", uri.String(), strings.NewReader(form.Encode()))
	if err != nil {
		return nil, fmt.Errorf("Error preparing request: %v", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client := http.Client{Transport: transport}
	response, err := client.Do(request)
	if err != nil {
		return nil, utils.Retryable(fmt.Errorf("Error sending request: %v", err))
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, utils.Retryable(fmt.Errorf("Error reading response: %v", err))
	}
	if response.StatusCode >= 500 {
		return nil, utils.Retryable(fmt.Errorf("Token service returned status code '%v' and body '%v'", response.StatusCode, string(bytes)))
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Token service returned status code '%v' and body '%v'", response.StatusCode, string(bytes))
	}

	var result tokenResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return nil, fmt.Errorf("Error parsing json response: %v", err)
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
