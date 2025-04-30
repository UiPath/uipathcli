package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/converter"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

type DuClient struct {
	baseUri  string
	token    *auth.AuthToken
	debug    bool
	settings plugin.ExecutionSettings
	logger   log.Logger
}

func (c DuClient) StartDigitization(projectId string, file stream.Stream, contentType string, uploadBar *visualization.ProgressBar) (string, error) {
	context, cancel := context.WithCancelCause(context.Background())
	request := c.createStartDigitizationRequest(projectId, file, contentType, uploadBar, cancel)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.SendWithContext(request, context)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result digitizeResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %w", err)
	}
	return result.DocumentId, nil
}

func (c DuClient) createStartDigitizationRequest(projectId string, file stream.Stream, contentType string, uploadBar *visualization.ProgressBar, cancel context.CancelCauseFunc) *network.HttpRequest {
	streamSize, _ := file.Size()
	bodyReader, bodyWriter := io.Pipe()
	formDataContentType := c.writeMultipartBody(bodyWriter, file, contentType, cancel)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, streamSize, uploadBar)

	uri := converter.NewUriBuilder(c.baseUri, "/api/framework/projects/{ProjectId}/digitization/start").
		FormatPath("ProjectId", projectId).
		AddQueryString("api-version", "1").
		Build()
	header := http.Header{
		"Content-Type": {formDataContentType},
	}
	return network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, uploadReader, -1)
}

func (c DuClient) GetDigitizationResult(projectId string, documentId string) (string, error) {
	request := c.createDigitizeStatusRequest(projectId, documentId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result digitizeResultResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %w", err)
	}
	if result.Status == "NotStarted" || result.Status == "Running" {
		return "", nil
	}
	return string(body), err
}

func (c DuClient) createDigitizeStatusRequest(projectId string, documentId string) *network.HttpRequest {
	uri := converter.NewUriBuilder(c.baseUri, "/api/framework/projects/{ProjectId}/digitization/result/{DocumentId}").
		FormatPath("ProjectId", projectId).
		FormatPath("DocumentId", documentId).
		AddQueryString("api-version", "1").
		Build()
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), http.Header{})
}

func (c DuClient) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, cancel context.CancelCauseFunc) string {
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		defer func() { _ = formWriter.Close() }()
		err := c.writeMultipartForm(formWriter, stream, contentType)
		if err != nil {
			cancel(err)
			return
		}
	}()
	return formWriter.FormDataContentType()
}

func (c DuClient) writeMultipartForm(writer *multipart.Writer, stream stream.Stream, contentType string) error {
	filePart := textproto.MIMEHeader{}
	filePart.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, stream.Name()))
	filePart.Set("Content-Type", contentType)
	w, err := writer.CreatePart(filePart)
	if err != nil {
		return fmt.Errorf("Error creating form field 'file': %w", err)
	}
	data, err := stream.Data()
	if err != nil {
		return err
	}
	defer func() { _ = data.Close() }()
	_, err = io.Copy(w, data)
	if err != nil {
		return fmt.Errorf("Error writing form field 'file': %w", err)
	}
	return nil
}

func (c DuClient) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c DuClient) httpClientSettings() network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		c.debug,
		c.settings.OperationId,
		c.settings.Header,
		c.settings.Timeout,
		c.settings.MaxAttempts,
		c.settings.Insecure)
}

func (c DuClient) toAuthorization(token *auth.AuthToken) *network.Authorization {
	if token == nil {
		return nil
	}
	return network.NewAuthorization(token.Type, token.Value)
}

func NewDuClient(baseUri string, token *auth.AuthToken, debug bool, settings plugin.ExecutionSettings, logger log.Logger) *DuClient {
	return &DuClient{baseUri, token, debug, settings, logger}
}

type digitizeResponse struct {
	DocumentId string `json:"documentId"`
}

type digitizeResultResponse struct {
	Status string `json:"status"`
}
