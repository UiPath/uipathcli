package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/utils/converter"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

const NotConfiguredErrorTemplate = `Run config command to set organization and tenant:

    uipath config

For more information you can view the help:

    uipath config --help
`

// The HttpExecutor implements the Executor interface and constructs HTTP request
// from the given command line parameters and configurations.
type HttpExecutor struct {
	authenticators []auth.Authenticator
}

func (e HttpExecutor) addHeaders(header http.Header, headerParameters []ExecutionParameter) {
	converter := converter.NewStringConverter()
	for _, parameter := range headerParameters {
		headerValue := converter.ToString(parameter.Value)
		header.Set(parameter.Name, headerValue)
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
			defer func() { _ = data.Close() }()
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

func (e HttpExecutor) authenticatorContext(ctx ExecutionContext, url string) auth.AuthenticatorContext {
	authRequest := *auth.NewAuthenticatorRequest(url, map[string]string{})
	return *auth.NewAuthenticatorContext(
		ctx.AuthConfig.Type,
		ctx.AuthConfig.Config,
		ctx.IdentityUri,
		ctx.Settings.OperationId,
		ctx.Settings.Insecure,
		authRequest)
}

func (e HttpExecutor) executeAuthenticators(ctx ExecutionContext, url string) (*auth.AuthenticatorResult, error) {
	var token *auth.AuthToken = nil
	for _, authProvider := range e.authenticators {
		authContext := e.authenticatorContext(ctx, url)
		result := authProvider.Auth(authContext)
		if result.Error != "" {
			return nil, errors.New(result.Error)
		}
		if result.Token != nil {
			token = result.Token
		}
	}
	return auth.AuthenticatorSuccess(token), nil
}

func (e HttpExecutor) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	if length < 10*1024*1024 {
		return reader
	}
	return visualization.NewProgressReader(reader, func(progress visualization.Progress) {
		displayText := text
		if progress.Completed {
			displayText = completedText
		}
		progressBar.UpdateProgress(displayText, progress.BytesRead, length, progress.BytesPerSecond)
	})
}

func (e HttpExecutor) writeMultipartBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, cancel context.CancelCauseFunc) (string, int64) {
	multipartSize := e.calculateMultipartSize(parameters)
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		defer func() { _ = formWriter.Close() }()
		err := e.writeMultipartForm(formWriter, parameters)
		if err != nil {
			cancel(err)
			return
		}
	}()
	return formWriter.FormDataContentType(), multipartSize
}

func (e HttpExecutor) writeInputBody(bodyWriter *io.PipeWriter, input stream.Stream, cancel context.CancelCauseFunc) {
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		data, err := input.Data()
		if err != nil {
			cancel(err)
			return
		}
		defer func() { _ = data.Close() }()
		_, err = io.Copy(bodyWriter, data)
		if err != nil {
			cancel(err)
			return
		}
	}()
}

func (e HttpExecutor) writeUrlEncodedBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, cancel context.CancelCauseFunc) {
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		queryStringBuilder := converter.NewQueryStringBuilder()
		for _, parameter := range parameters {
			queryStringBuilder.Add(parameter.Name, parameter.Value)
		}
		queryString := queryStringBuilder.Build()
		_, err := bodyWriter.Write([]byte(queryString))
		if err != nil {
			cancel(err)
			return
		}
	}()
}

func (e HttpExecutor) writeJsonBody(bodyWriter *io.PipeWriter, parameters []ExecutionParameter, cancel context.CancelCauseFunc) {
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		err := e.serializeJson(bodyWriter, parameters)
		if err != nil {
			cancel(err)
			return
		}
	}()
}

func (e HttpExecutor) writeBody(ctx ExecutionContext, cancel context.CancelCauseFunc) (io.ReadCloser, string, int64, int64) {
	if ctx.Input != nil {
		reader, writer := io.Pipe()
		e.writeInputBody(writer, ctx.Input, cancel)
		contentLength, _ := ctx.Input.Size()
		return reader, ctx.ContentType, contentLength, contentLength
	}
	formParameters := ctx.Parameters.Form()
	if len(formParameters) > 0 {
		reader, writer := io.Pipe()
		contentType, multipartSize := e.writeMultipartBody(writer, formParameters, cancel)
		return reader, contentType, -1, multipartSize
	}
	bodyParameters := ctx.Parameters.Body()
	if len(bodyParameters) > 0 && ctx.ContentType == "application/x-www-form-urlencoded" {
		reader, writer := io.Pipe()
		e.writeUrlEncodedBody(writer, bodyParameters, cancel)
		return reader, ctx.ContentType, -1, -1
	}
	if len(bodyParameters) > 0 {
		reader, writer := io.Pipe()
		e.writeJsonBody(writer, bodyParameters, cancel)
		return reader, ctx.ContentType, -1, -1
	}
	return io.NopCloser(bytes.NewReader([]byte{})), ctx.ContentType, -1, -1
}

func (e HttpExecutor) pathParameters(ctx ExecutionContext) []ExecutionParameter {
	pathParameters := ctx.Parameters.Path()
	if ctx.Organization != "" {
		pathParameters = append(pathParameters, *NewExecutionParameter("organization", ctx.Organization, "path"))
	}
	if ctx.Tenant != "" {
		pathParameters = append(pathParameters, *NewExecutionParameter("tenant", ctx.Tenant, "path"))
	}
	return pathParameters
}

func (e HttpExecutor) httpClientSettings(ctx ExecutionContext) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		ctx.Debug,
		ctx.Settings.OperationId,
		ctx.Settings.Header,
		ctx.Settings.Timeout,
		ctx.Settings.MaxAttempts,
		ctx.Settings.Insecure)
}

func (e HttpExecutor) toAuthorization(token *auth.AuthToken) *network.Authorization {
	if token == nil {
		return nil
	}
	return network.NewAuthorization(token.Type, token.Value)
}

func (e HttpExecutor) Call(ctx ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	uri, err := e.formatUri(ctx.BaseUri, ctx.Route, e.pathParameters(ctx), ctx.Parameters.Query())
	if err != nil {
		return err
	}
	context, cancel := context.WithCancelCause(context.Background())
	bodyReader, contentType, contentLength, size := e.writeBody(ctx, cancel)
	uploadBar := visualization.NewProgressBar(logger)
	uploadReader := e.progressReader("uploading...", "completing  ", bodyReader, size, uploadBar)
	defer uploadBar.Remove()

	auth, err := e.executeAuthenticators(ctx, uri.String())
	if err != nil {
		return err
	}

	header := http.Header{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	e.addHeaders(header, ctx.Parameters.Header())
	request := network.NewHttpRequest(ctx.Method, uri.String(), e.toAuthorization(auth.Token), header, uploadReader, contentLength)

	client := network.NewHttpClient(logger, e.httpClientSettings(ctx))
	response, err := client.SendWithContext(request, context)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
	downloadBar := visualization.NewProgressBar(logger)
	downloadReader := e.progressReader("downloading...", "completing    ", response.Body, response.ContentLength, downloadBar)
	defer downloadBar.Remove()
	body, err := io.ReadAll(downloadReader)
	if err != nil {
		return fmt.Errorf("Error reading response body: %w", err)
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
