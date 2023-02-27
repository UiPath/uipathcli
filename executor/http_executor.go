package executor

import (
	"bytes"
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/utils"
)

const NotConfiguredErrorTemplate = `Run config command to set organization and tenant:

    uipathcli config

For more information you can view the help:

    uipathcli config --help
`

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

func (e HttpExecutor) calculateMultipartSize(parameters []ExecutionParameter) int64 {
	result := int64(0)
	for _, parameter := range parameters {
		switch v := parameter.Value.(type) {
		case string:
			result = result + int64(len(v))
		case FileReference:
			data, size, err := v.Data()
			if err == nil {
				defer data.Close()
				result = result + size
			}
		}
	}
	return result
}

func (e HttpExecutor) writeMultipartForm(writer *multipart.Writer, parameters []ExecutionParameter) error {
	for _, parameter := range parameters {
		switch v := parameter.Value.(type) {
		case string:
			w, err := writer.CreateFormField(parameter.Name)
			if err != nil {
				return fmt.Errorf("Error creating form field '%s': %v", parameter.Name, err)
			}
			_, err = w.Write([]byte(v))
			if err != nil {
				return fmt.Errorf("Error writing form field '%s': %v", parameter.Name, err)
			}
		case FileReference:
			w, err := writer.CreateFormFile(parameter.Name, v.Filename())
			if err != nil {
				return fmt.Errorf("Error writing form file '%s': %v", parameter.Name, err)
			}
			data, _, err := v.Data()
			if err != nil {
				return err
			}
			defer data.Close()
			_, err = io.Copy(w, data)
			if err != nil {
				return fmt.Errorf("Error writing form file '%s': %v", parameter.Name, err)
			}
		}
	}
	return nil
}

func (e HttpExecutor) serializeJson(body io.Writer, parameters []ExecutionParameter) error {
	var data = map[string]interface{}{}
	for _, parameter := range parameters {
		data[parameter.Name] = parameter.Value
	}
	result, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error creating body: %v", err)
	}
	body.Write(result)
	return nil
}

func (e HttpExecutor) validateUri(uri string) (*url.URL, error) {
	if strings.Contains(uri, "{organization}") {
		return nil, fmt.Errorf("Missing organization parameter!\n\n%s", NotConfiguredErrorTemplate)
	}
	if strings.Contains(uri, "{tenant}") {
		return nil, fmt.Errorf("Missing tenant parameter!\n\n%s", NotConfiguredErrorTemplate)
	}

	result, err := url.Parse(uri)
	if err != nil {
		return nil, fmt.Errorf("Invalid URI '%s': %v", uri, err)
	}
	return result, nil
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
	return e.validateUri(uri)
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

func (e HttpExecutor) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *utils.ProgressBar) io.Reader {
	if length == -1 || length < 10*1024*1024 {
		return reader
	}
	progressReader := utils.NewProgressReader(reader, func(progress utils.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.Update(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
	return progressReader
}

func (e HttpExecutor) writeMultipartBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, errorChan chan error) (string, int64) {
	contentLength := e.calculateMultipartSize(parameters)
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer bodyWriter.Close()
		defer formWriter.Close()
		err := e.writeMultipartForm(formWriter, parameters)
		if err != nil {
			errorChan <- err
			return
		}
	}()
	return formWriter.FormDataContentType(), contentLength
}

func (e HttpExecutor) writeInputBody(bodyWriter *io.PipeWriter, input FileReference, errorChan chan error) {
	go func() {
		defer bodyWriter.Close()
		data, _, err := input.Data()
		if err != nil {
			errorChan <- err
			return
		}
		defer data.Close()
		_, err = io.Copy(bodyWriter, data)
		if err != nil {
			errorChan <- err
			return
		}
	}()
}

func (e HttpExecutor) writeJsonBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, errorChan chan error) {
	go func() {
		defer bodyWriter.Close()
		err := e.serializeJson(bodyWriter, parameters)
		if err != nil {
			errorChan <- err
			return
		}
	}()
}

func (e HttpExecutor) writeBody(context ExecutionContext, errorChan chan error) (io.Reader, string, int64) {
	bodyReader, bodyWriter := io.Pipe()
	if context.Input != nil {
		e.writeInputBody(bodyWriter, *context.Input, errorChan)
		data, size, err := context.Input.Data()
		if err == nil {
			defer data.Close()
		}
		return bodyReader, context.ContentType, size
	}
	if len(context.Parameters.Form) > 0 {
		contentType, contentLength := e.writeMultipartBody(bodyWriter, context.Parameters.Form, errorChan)
		return bodyReader, contentType, contentLength
	}
	if len(context.Parameters.Body) > 0 {
		e.writeJsonBody(bodyWriter, context.Parameters.Body, errorChan)
		return bodyReader, context.ContentType, -1
	}
	go func() {
		defer bodyWriter.Close()
	}()
	return bodyReader, context.ContentType, -1
}

func (e HttpExecutor) send(client *http.Client, request *http.Request, errorChan chan error) (*http.Response, error) {
	responseChan := make(chan *http.Response)
	go func(client *http.Client, request *http.Request) {
		response, err := client.Do(request)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}(client, request)

	select {
	case err := <-errorChan:
		return nil, err
	case response := <-responseChan:
		return response, nil
	}
}

func (e HttpExecutor) LogRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (e HttpExecutor) LogResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}

func (e HttpExecutor) pathParameters(context ExecutionContext) []ExecutionParameter {
	pathParameters := context.Parameters.Path
	if context.Organization != "" {
		pathParameters = append(pathParameters, *NewExecutionParameter("organization", context.Organization))
	}
	if context.Tenant != "" {
		pathParameters = append(pathParameters, *NewExecutionParameter("tenant", context.Tenant))
	}
	return pathParameters
}

func (e HttpExecutor) Call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	uri, err := e.formatUri(context.BaseUri, context.Route, e.pathParameters(context), context.Parameters.Query)
	if err != nil {
		return err
	}
	requestError := make(chan error)
	bodyReader, contentType, contentLength := e.writeBody(context, requestError)
	uploadBar := utils.NewProgressBar(logger)
	uploadReader := e.progressReader("uploading...", "completing  ", bodyReader, contentLength, uploadBar)
	defer uploadBar.Remove()
	request, err := http.NewRequest(context.Method, uri.String(), uploadReader)
	if err != nil {
		return fmt.Errorf("Error preparing request: %v", err)
	}
	if contentType != "" {
		request.Header.Add("Content-Type", contentType)
	}
	e.addHeaders(request, context.Parameters.Header)
	auth, err := e.executeAuthenticators(context.AuthConfig, context.Debug, context.Insecure, request)
	if err != nil {
		return err
	}
	for k, v := range auth.RequestHeader {
		request.Header.Add(k, v)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: context.Insecure},
	}
	client := &http.Client{Transport: transport}
	if context.Debug {
		e.LogRequest(logger, request)
	}
	response, err := e.send(client, request, requestError)
	if err != nil {
		return fmt.Errorf("Error sending request: %v", err)
	}
	downloadBar := utils.NewProgressBar(logger)
	downloadReader := e.progressReader("downloading...", "completing    ", response.Body, response.ContentLength, downloadBar)
	defer downloadBar.Remove()
	defer response.Body.Close()
	body, err := io.ReadAll(downloadReader)
	if err != nil {
		return fmt.Errorf("Error reading response body: %v", err)
	}
	e.LogResponse(logger, response, body)
	err = writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body)))
	if err != nil {
		return err
	}
	return nil
}
