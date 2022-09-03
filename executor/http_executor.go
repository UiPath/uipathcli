package executor

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type HttpExecutor struct {
	TokenProvider TokenProvider
}

func RequestId() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

func (e HttpExecutor) addHeaders(request *http.Request, headerParameters []ExecutionParameter) {
	request.Header.Add("x-request-id", RequestId())
	for _, parameter := range headerParameters {
		request.Header.Add(parameter.Name, parameter.Value.(string))
	}
}

func (e HttpExecutor) createJson(parameters []ExecutionParameter) ([]byte, string, error) {
	var body = map[string]interface{}{}
	for _, parameter := range parameters {
		body[parameter.Name] = parameter.Value
	}
	result, err := json.Marshal(body)
	if err != nil {
		return []byte{}, "", fmt.Errorf("Error creating body: %v", err)
	}
	return result, "application/json", nil
}

func (e HttpExecutor) createBody(bodyParameters []ExecutionParameter, formParameters []ExecutionParameter) ([]byte, string, error) {
	if len(bodyParameters) > 0 {
		return e.createJson(bodyParameters)
	}
	return []byte{}, "", nil
}

func (e HttpExecutor) formatUri(baseUri url.URL, route string, pathParameters []ExecutionParameter, queryParameters []ExecutionParameter) (*url.URL, error) {
	uri := fmt.Sprintf("%s://%s%s%s", baseUri.Scheme, baseUri.Host, baseUri.Path, route)

	for _, parameter := range pathParameters {
		uri = strings.Replace(uri, "{"+parameter.Name+"}", parameter.Value.(string), -1)
	}

	querySeparator := "?"
	for _, parameter := range queryParameters {
		uri = uri + querySeparator + parameter.Name + "=" + url.QueryEscape(parameter.Value.(string))
		querySeparator = "&"
	}

	result, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("Invalid URI '%s': %v", uri, err)
	}
	return result, nil
}

func (e HttpExecutor) send(client *http.Client, req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	for i := 0; i < 3; i++ {
		resp, err = client.Do(req)
		if err == nil && resp.StatusCode < 500 {
			return resp, nil
		}
	}
	return resp, err
}

func (e HttpExecutor) Call(context ExecutionContext) (string, error) {
	uri, err := e.formatUri(context.BaseUri, context.Route, context.PathParameters, context.QueryParameters)
	if err != nil {
		return "", err
	}
	body, contentType, err := e.createBody(context.BodyParameters, context.FormParameters)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequest(context.Method, uri.String(), bytes.NewReader(body))
	if contentType != "" {
		request.Header.Add("Content-Type", contentType)
	}
	e.addHeaders(request, context.HeaderParameters)
	if err != nil {
		return "", fmt.Errorf("Error preparing request: %v", err)
	}

	if context.ClientId != "" && context.ClientSecret != "" {
		tokenRequest := NewTokenRequest(
			fmt.Sprintf("%s://%s/identity_", uri.Scheme, uri.Host),
			context.ClientId,
			context.ClientSecret,
			context.Insecure)
		token, err := e.TokenProvider.GetToken(*tokenRequest)
		if err != nil {
			return "", fmt.Errorf("Error retrieving bearer token: %v", err)
		}
		request.Header.Add("Authorization", "Bearer "+token)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: context.Insecure},
	}
	client := &http.Client{Transport: transport}
	logger := HttpLogger{
		Output: &bytes.Buffer{},
	}
	err = logger.LogRequest(request, bytes.NewReader(body), context.Debug)
	if err != nil {
		return "", err
	}
	response, err := e.send(client, request)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer response.Body.Close()
	err = logger.LogResponse(response, context.Debug)
	if err != nil {
		return "", err
	}
	return logger.Output.String(), nil
}
