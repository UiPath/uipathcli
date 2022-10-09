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
)

type identityClient struct {
	Cache cache.Cache
}

const TokenRoute = "/connect/token"

func (c identityClient) GetToken(tokenRequest tokenRequest) (string, error) {
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
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Token service returned status code '%v' and body '%v'", response.StatusCode, string(bytes))
	}

	var result identityResponse
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %v", err)
	}

	c.Cache.Set(cacheKey, result.AccessToken, result.ExpiresIn)
	return result.AccessToken, nil
}

func (c identityClient) cacheKey(tokenRequest tokenRequest) string {
	return fmt.Sprintf("token|%v|%v|%v", tokenRequest.BaseUri, tokenRequest.ClientId, tokenRequest.ClientSecret)
}
