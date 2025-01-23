package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/network"
)

type DataServiceClient struct {
	baseUri  string
	token    *auth.AuthToken
	debug    bool
	settings plugin.ExecutionSettings
	logger   log.Logger
}

func (c DataServiceClient) GetEntitySpecification() ([]byte, error) {
	request := c.createGetEntitySpecificationRequest()
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return []byte{}, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Data Service returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return body, nil
}

func (c DataServiceClient) createGetEntitySpecificationRequest() *network.HttpRequest {
	uri := c.baseUri + "/api/DataService"
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), http.Header{})
}

func (c DataServiceClient) CallEntity(method string, uri string, header map[string]string, body map[string]interface{}) (*network.HttpResponse, error) {
	request, err := c.createCallEntityRequest(method, uri, header, body)
	if err != nil {
		return nil, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	return client.Send(request)
}

func (c DataServiceClient) createCallEntityRequest(method string, uri string, header map[string]string, body map[string]interface{}) (*network.HttpRequest, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("Error creating body: %w", err)
	}
	httpHeader := http.Header{
		"Content-Type": {"application/json"},
	}
	for k, v := range header {
		httpHeader.Add(k, v)
	}
	request := network.NewHttpRequest(method, uri, c.toAuthorization(c.token), httpHeader, bytes.NewReader(data), -1)
	return request, nil
}

func (c DataServiceClient) httpClientSettings() network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		c.debug,
		c.settings.OperationId,
		c.settings.Header,
		c.settings.Timeout,
		c.settings.MaxAttempts,
		c.settings.Insecure)
}

func (c DataServiceClient) toAuthorization(token *auth.AuthToken) *network.Authorization {
	if token == nil {
		return nil
	}
	return network.NewAuthorization(token.Type, token.Value)
}

func NewDataServiceClient(baseUri string, token *auth.AuthToken, debug bool, settings plugin.ExecutionSettings, logger log.Logger) *DataServiceClient {
	return &DataServiceClient{baseUri, token, debug, settings, logger}
}
