package executor

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
)

type HttpExecutor struct {
	Authenticators []auth.Authenticator
}

func RequestId() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

func (e HttpExecutor) addHeaders(request *http.Request, headerParameters []ExecutionParameter) {
	formatter := TypeFormatter{}
	request.Header.Add("x-request-id", RequestId())
	for _, parameter := range headerParameters {
		headerValue := formatter.FormatHeader(parameter)
		request.Header.Add(parameter.Name, headerValue)
	}
}

func (e HttpExecutor) createForm(parameters []ExecutionParameter) ([]byte, string, error) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	for _, parameter := range parameters {
		switch v := parameter.Value.(type) {
		case string:
			w, err := writer.CreateFormField(parameter.Name)
			if err != nil {
				return []byte{}, "", fmt.Errorf("Error creating form field '%s': %v", parameter.Name, err)
			}
			_, err = w.Write([]byte(v))
			if err != nil {
				return []byte{}, "", fmt.Errorf("Error writing form field '%s': %v", parameter.Name, err)
			}
		case FileReference:
			w, err := writer.CreateFormFile(parameter.Name, v.Filename)
			if err != nil {
				return []byte{}, "", fmt.Errorf("Error writing form file '%s': %v", parameter.Name, err)
			}
			_, err = w.Write(v.Data)
			if err != nil {
				return []byte{}, "", fmt.Errorf("Error writing form file '%s': %v", parameter.Name, err)
			}
		}
	}
	writer.Close()
	return b.Bytes(), writer.FormDataContentType(), nil
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
	if len(formParameters) > 0 {
		return e.createForm(formParameters)
	}
	if len(bodyParameters) > 0 {
		return e.createJson(bodyParameters)
	}
	return []byte{}, "", nil
}

func (e HttpExecutor) formatUri(baseUri url.URL, route string, pathParameters []ExecutionParameter, queryParameters []ExecutionParameter) (*url.URL, error) {
	formatter := TypeFormatter{}
	normalizedPath := strings.Trim(baseUri.Path, "/")
	normalizedRoute := strings.Trim(route, "/")
	path := path.Join(normalizedPath, normalizedRoute)

	uri := fmt.Sprintf("%s://%s/%s", baseUri.Scheme, baseUri.Host, path)
	for _, parameter := range pathParameters {
		pathValue := formatter.FormatPath(parameter)
		uri = strings.ReplaceAll(uri, "{"+parameter.Name+"}", pathValue)
	}

	querySeparator := "?"
	for _, parameter := range queryParameters {
		queryStringValue := formatter.FormatQueryString(parameter)
		uri = uri + querySeparator + queryStringValue
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

func (e HttpExecutor) executeAuthenticators(authConfig config.AuthConfig, debug bool, insecure bool, request *http.Request) (*auth.AuthenticatorResult, error) {
	authRequest := *auth.NewAuthenticatorRequest(request.URL.String(), map[string]string{})
	ctx := *auth.NewAuthenticatorContext(authConfig.Type, authConfig.Config, debug, insecure, authRequest)
	for _, authProvider := range e.Authenticators {
		result := authProvider.Auth(ctx)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		ctx.Config = result.Config
		for k, v := range result.RequestHeader {
			ctx.Request.Header[k] = v
		}
	}
	return auth.AuthenticatorSuccess(ctx.Request.Header, ctx.Config), nil
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
	auth, err := e.executeAuthenticators(context.AuthConfig, context.Debug, context.Insecure, request)
	if err != nil {
		return "", err
	}
	for k, v := range auth.RequestHeader {
		request.Header.Add(k, v)
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
