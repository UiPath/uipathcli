package executor

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type IdentityClient struct {
	Cache Cache
}

const TokenRoute = "/connect/token"

func (c IdentityClient) GetToken(tokenRequest TokenRequest) (string, error) {
	cacheKey := c.cacheKey(tokenRequest)
	token := c.Cache.Get(cacheKey)
	if token != "" {
		return token, nil
	}

	form := url.Values{}
	form.Add("client_id", tokenRequest.ClientId)
	form.Add("client_secret", tokenRequest.ClientSecret)
	form.Add("grant_type", "client_credentials")

	uri := tokenRequest.BaseUri + TokenRoute
	request, err := http.NewRequest("POST", uri, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("Error preparing request: %v", err)
	}
	request.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: tokenRequest.Insecure},
	}
	client := http.Client{Transport: transport}
	response, err := client.Do(request)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}

	var result identityResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %v", err)
	}

	c.Cache.Set(cacheKey, result.AccessToken, result.ExpiresIn)
	return result.AccessToken, nil
}

func (c IdentityClient) cacheKey(tokenRequest TokenRequest) string {
	return fmt.Sprintf("token|%v|%v|%v", tokenRequest.BaseUri, tokenRequest.ClientId, tokenRequest.ClientSecret)
}
