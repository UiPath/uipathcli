package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/network"
)

type identityClient struct {
	logger log.Logger
}

const tokenRoute = "/connect/token"
const redactedMessage = "**redacted**"

func (c identityClient) GetToken(tokenRequest tokenRequest) (*tokenResponse, error) {
	request := tokenRequestForm{
		GrantType:    tokenRequest.GrantType,
		Scopes:       tokenRequest.Scopes,
		ClientId:     tokenRequest.ClientId,
		ClientSecret: tokenRequest.ClientSecret,
		Code:         tokenRequest.Code,
		CodeVerifier: tokenRequest.CodeVerifier,
		RedirectUri:  tokenRequest.RedirectUri,
		RefreshToken: tokenRequest.RefreshToken,
		Properties:   tokenRequest.Properties,
	}
	response, err := c.retrieveToken(tokenRequest.BaseUri, request, tokenRequest.Settings)
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(time.Duration(response.ExpiresIn) * time.Second)
	return newTokenResponse(response.AccessToken, expiresAt, response.RefreshToken), nil
}

func (c identityClient) retrieveToken(baseUri url.URL, form tokenRequestForm, settings network.HttpClientSettings) (*tokenResponseJson, error) {
	uri := baseUri.JoinPath(tokenRoute)
	header := http.Header{
		"Content-Type": {"application/x-www-form-urlencoded"},
	}
	request := network.NewHttpPostRequest(uri.String(), nil, header, strings.NewReader(form.Encode()), -1)
	c.logRequest(settings.Debug, request, strings.NewReader(form.Redacted()))

	settingsNoDebug := c.networkSettingsNoDebug(settings)
	client := network.NewHttpClient(nil, settingsNoDebug)
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		c.logResponse(settings.Debug, response, bytes.NewReader(responseBody))
		return nil, fmt.Errorf("Token service returned status code '%d' and body '%s'", response.StatusCode, string(responseBody))
	}

	var responseJson tokenResponseJson
	err = json.Unmarshal(responseBody, &responseJson)
	if err != nil {
		c.logResponse(settings.Debug, response, bytes.NewReader(responseBody))
		return nil, fmt.Errorf("Error parsing json response: %w", err)
	}
	c.logResponseJson(settings.Debug, response, responseJson)
	return &responseJson, nil
}

func (c identityClient) logRequest(debug bool, request *network.HttpRequest, body io.Reader) {
	if !debug {
		return
	}
	requestInfo := log.NewRequestInfo(request.Method, request.URL, request.Proto, request.Header, body)
	c.logger.LogRequest(*requestInfo)
}

func (c identityClient) logResponse(debug bool, response *network.HttpResponse, body io.Reader) {
	if !debug {
		return
	}
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, body)
	c.logger.LogResponse(*responseInfo)
}

func (c identityClient) logResponseJson(debug bool, response *network.HttpResponse, body tokenResponseJson) {
	if !debug {
		return
	}
	data, _ := json.Marshal(body.Redacted())
	c.logResponse(debug, response, bytes.NewReader(data))
}

func (c identityClient) networkSettingsNoDebug(settings network.HttpClientSettings) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		false,
		settings.OperationId,
		settings.Header,
		settings.Timeout,
		settings.MaxAttempts,
		settings.Insecure,
	)
}

func newIdentityClient(logger log.Logger) *identityClient {
	return &identityClient{logger}
}
