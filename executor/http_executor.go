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
	"runtime"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/utils"
	"github.com/UiPath/uipathcli/utils/converter"
	"github.com/UiPath/uipathcli/utils/resiliency"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

const NotConfiguredErrorTemplate = `Run config command to set organization and tenant:

    uipath config

For more information you can view the help:

    uipath config --help
`

var UserAgent = fmt.Sprintf("uipathcli/%s (%s; %s)", utils.Version, runtime.GOOS, runtime.GOARCH)

// The HttpExecutor implements the Executor interface and constructs HTTP request
// from the given command line parameters and configurations.
type HttpExecutor struct {
	authenticators []auth.Authenticator
}

func (e HttpExecutor) Call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	return resiliency.RetryN(context.MaxAttempts, func() error {
		return e.call(context, writer, logger)
	})
}

func (e HttpExecutor) requestId() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

func (e HttpExecutor) addHeaders(request *http.Request, headerParameters []ExecutionParameter) {
	converter := converter.NewStringConverter()
	request.Header.Add("x-request-id", e.requestId())
	request.Header.Add("User-Agent", UserAgent)
	for _, parameter := range headerParameters {
		headerValue := converter.ToString(parameter.Value)
		request.Header.Add(parameter.Name, headerValue)
	}
}

func (e HttpExecutor) calculateMultipartSize(parameters []ExecutionParameter) int64 {
	result := int64(0)
	for _, parameter := range parameters {
		switch v := parameter.Value.(type) {
		case string:
			result = result + int64(len(v))
		case stream.Stream:
			size, err := v.Size()
			if err == nil {
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
				return fmt.Errorf("Error creating form field '%s': %w", parameter.Name, err)
			}
			_, err = w.Write([]byte(v))
			if err != nil {
				return fmt.Errorf("Error writing form field '%s': %w", parameter.Name, err)
			}
		case stream.Stream:
			w, err := writer.CreateFormFile(parameter.Name, v.Name())
			if err != nil {
				return fmt.Errorf("Error writing form file '%s': %w", parameter.Name, err)
			}
			data, err := v.Data()
			if err != nil {
				return err
			}
			defer data.Close()
			_, err = io.Copy(w, data)
			if err != nil {
				return fmt.Errorf("Error writing form file '%s': %w", parameter.Name, err)
			}
		}
	}
	return nil
}

func (e HttpExecutor) serializeJson(body io.Writer, parameters []ExecutionParameter) error {
	data := map[string]interface{}{}
	for _, parameter := range parameters {
		data[parameter.Name] = parameter.Value
	}
	result, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("Error creating body: %w", err)
	}
	_, err = body.Write(result)
	if err != nil {
		return fmt.Errorf("Error writing body: %w", err)
	}
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
		return nil, fmt.Errorf("Invalid URI '%s': %w", uri, err)
	}
	return result, nil
}

func (e HttpExecutor) formatUri(baseUri url.URL, route string, pathParameters []ExecutionParameter, queryParameters []ExecutionParameter) (*url.URL, error) {
	uriBuilder := converter.NewUriBuilder(baseUri, route)
	for _, parameter := range pathParameters {
		uriBuilder.FormatPath(parameter.Name, parameter.Value)
	}
	for _, parameter := range queryParameters {
		uriBuilder.AddQueryString(parameter.Name, parameter.Value)
	}
	return e.validateUri(uriBuilder.Build())
}

func (e HttpExecutor) executeAuthenticators(authConfig config.AuthConfig, identityUri url.URL, debug bool, insecure bool, request *http.Request) (*auth.AuthenticatorResult, error) {
	authRequest := *auth.NewAuthenticatorRequest(request.URL.String(), map[string]string{})
	ctx := *auth.NewAuthenticatorContext(authConfig.Type, authConfig.Config, identityUri, debug, insecure, authRequest)
	for _, authProvider := range e.authenticators {
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

func (e HttpExecutor) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	if length < 10*1024*1024 {
		return reader
	}
	progressReader := visualization.NewProgressReader(reader, func(progress visualization.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.UpdateProgress(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
	return progressReader
}

func (e HttpExecutor) writeMultipartBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, errorChan chan error) (string, int64) {
	multipartSize := e.calculateMultipartSize(parameters)
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
	return formWriter.FormDataContentType(), multipartSize
}

func (e HttpExecutor) writeInputBody(bodyWriter *io.PipeWriter, input stream.Stream, errorChan chan error) {
	go func() {
		defer bodyWriter.Close()
		data, err := input.Data()
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

func (e HttpExecutor) writeUrlEncodedBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, errorChan chan error) {
	go func() {
		defer bodyWriter.Close()
		queryStringBuilder := converter.NewQueryStringBuilder()
		for _, parameter := range parameters {
			queryStringBuilder.Add(parameter.Name, parameter.Value)
		}
		queryString := queryStringBuilder.Build()
		_, err := bodyWriter.Write([]byte(queryString))
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

func (e HttpExecutor) writeBody(context ExecutionContext, errorChan chan error) (io.Reader, string, int64, int64) {
	if context.Input != nil {
		reader, writer := io.Pipe()
		e.writeInputBody(writer, context.Input, errorChan)
		contentLength, _ := context.Input.Size()
		return reader, context.ContentType, contentLength, contentLength
	}
	formParameters := context.Parameters.Form()
	if len(formParameters) > 0 {
		reader, writer := io.Pipe()
		contentType, multipartSize := e.writeMultipartBody(writer, formParameters, errorChan)
		return reader, contentType, -1, multipartSize
	}
	bodyParameters := context.Parameters.Body()
	if len(bodyParameters) > 0 && context.ContentType == "application/x-www-form-urlencoded" {
		reader, writer := io.Pipe()
		e.writeUrlEncodedBody(writer, bodyParameters, errorChan)
		return reader, context.ContentType, -1, -1
	}
	if len(bodyParameters) > 0 {
		reader, writer := io.Pipe()
		e.writeJsonBody(writer, bodyParameters, errorChan)
		return reader, context.ContentType, -1, -1
	}
	return bytes.NewReader([]byte{}), context.ContentType, -1, -1
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

func (e HttpExecutor) logRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	_, _ = buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (e HttpExecutor) logResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}

func (e HttpExecutor) pathParameters(context ExecutionContext) []ExecutionParameter {
	pathParameters := context.Parameters.Path()
	if context.Organization != "" {
		pathParameters = append(pathParameters, *NewExecutionParameter("organization", context.Organization, "path"))
	}
	if context.Tenant != "" {
		pathParameters = append(pathParameters, *NewExecutionParameter("tenant", context.Tenant, "path"))
	}
	return pathParameters
}

func (e HttpExecutor) call(context ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	uri, err := e.formatUri(context.BaseUri, context.Route, e.pathParameters(context), context.Parameters.Query())
	if err != nil {
		return err
	}
	requestError := make(chan error)
	bodyReader, contentType, contentLength, size := e.writeBody(context, requestError)
	uploadBar := visualization.NewProgressBar(logger)
	uploadReader := e.progressReader("uploading...", "completing  ", bodyReader, size, uploadBar)
	defer uploadBar.Remove()
	request, err := http.NewRequest(context.Method, uri.String(), uploadReader)
	if err != nil {
		return fmt.Errorf("Error preparing request: %w", err)
	}
	if contentType != "" {
		request.Header.Add("Content-Type", contentType)
	}
	if contentLength != -1 {
		request.ContentLength = contentLength
	}
	e.addHeaders(request, context.Parameters.Header())
	auth, err := e.executeAuthenticators(context.AuthConfig, context.IdentityUri, context.Debug, context.Insecure, request)
	if err != nil {
		return err
	}
	for k, v := range auth.RequestHeader {
		request.Header.Add(k, v)
	}

	transport := &http.Transport{
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: context.Insecure}, //nolint // This is user configurable and disabled by default
		ResponseHeaderTimeout: context.Timeout,
	}
	client := &http.Client{Transport: transport}
	if context.Debug {
		e.logRequest(logger, request)
	}
	response, err := e.send(client, request, requestError)
	if err != nil {
		return resiliency.Retryable(fmt.Errorf("Error sending request: %w", err))
	}
	defer response.Body.Close()
	downloadBar := visualization.NewProgressBar(logger)
	downloadReader := e.progressReader("downloading...", "completing    ", response.Body, response.ContentLength, downloadBar)
	defer downloadBar.Remove()
	body, err := io.ReadAll(downloadReader)
	if err != nil {
		return resiliency.Retryable(fmt.Errorf("Error reading response body: %w", err))
	}
	e.logResponse(logger, response, body)
	if response.StatusCode >= 500 {
		return resiliency.Retryable(fmt.Errorf("Service returned status code '%v' and body '%v'", response.StatusCode, string(body)))
	}
	err = writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body)))
	if err != nil {
		return err
	}
	return nil
}

func NewHttpExecutor(authenticators []auth.Authenticator) *HttpExecutor {
	return &HttpExecutor{authenticators}
}
