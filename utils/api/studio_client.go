package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/converter"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// StudioClient is an HTTP client for the Studio Web backend API.
type StudioClient struct {
	baseUri      url.URL
	organization string
	token        *auth.AuthToken
	debug        bool
	settings     plugin.ExecutionSettings
	logger       log.Logger
}

// PushSolution uploads a .uis file to Studio Web.
func (c StudioClient) PushSolution(file stream.Stream, solutionId string, uploadBar *visualization.ProgressBar) (*PushSolutionResponse, error) {
	ctx, cancel := context.WithCancelCause(context.Background())
	request := c.createPushSolutionRequest(file, solutionId, uploadBar, cancel)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.SendWithContext(request, ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("Studio Web returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result PushSolutionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return &PushSolutionResponse{}, nil
	}
	return &result, nil
}

func (c StudioClient) createPushSolutionRequest(file stream.Stream, solutionId string, uploadBar *visualization.ProgressBar, cancel context.CancelCauseFunc) *network.HttpRequest {
	bodyReader, bodyWriter := io.Pipe()
	streamSize, _ := file.Size()
	contentType := c.writeMultipartBody(bodyWriter, file, "application/octet-stream", cancel)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, streamSize, uploadBar)

	uriBuilder := c.newUriBuilder("/api/v1/ExternalSolution/Push")
	if solutionId != "" {
		uriBuilder.AddQueryString("solutionId", solutionId)
	}
	uri := uriBuilder.Build()
	header := http.Header{
		"Content-Type": {contentType},
	}
	return network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, uploadReader, -1)
}

// PullSolution downloads a solution from Studio Web as a .uis file.
func (c StudioClient) PullSolution(solutionId string) (io.ReadCloser, error) {
	request := c.createPullSolutionRequest(solutionId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		defer func() { _ = response.Body.Close() }()
		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("Error reading response: %w", err)
		}
		return nil, fmt.Errorf("Studio Web returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return response.Body, nil
}

func (c StudioClient) createPullSolutionRequest(solutionId string) *network.HttpRequest {
	uri := c.newUriBuilder("/api/v1/ExternalSolution/Pull").
		AddQueryString("solutionId", solutionId).
		Build()
	header := http.Header{
		"Accept": {"application/octet-stream"},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

// ListSolutions retrieves the list of solutions from Studio Web.
func (c StudioClient) ListSolutions() ([]SolutionInfo, error) {
	request := c.createListSolutionsRequest()
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Studio Web returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result []SolutionInfo
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Studio Web returned invalid response body '%v'", string(body))
	}
	return result, nil
}

func (c StudioClient) createListSolutionsRequest() *network.HttpRequest {
	uri := c.newUriBuilder("/api/v1/ExternalSolution/List").Build()
	header := http.Header{
		"Content-Type": {"application/json"},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

// PublishSolution publishes a solution for deployment.
func (c StudioClient) PublishSolution(solutionId string) (*PublishSolutionResponse, error) {
	requestBody, err := json.Marshal(publishSolutionRequestJson{
		SolutionId: solutionId,
	})
	if err != nil {
		return nil, err
	}

	uri := c.newUriBuilder("/api/v1/Publish-Requests").Build()
	header := http.Header{
		"Content-Type": {"application/json"},
	}
	request := network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, bytes.NewBuffer(requestBody), -1)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK && response.StatusCode != http.StatusCreated && response.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("Studio Web returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result PublishSolutionResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return &PublishSolutionResponse{}, nil
	}
	return &result, nil
}

func (c StudioClient) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, cancel context.CancelCauseFunc) string {
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

func (c StudioClient) writeMultipartForm(writer *multipart.Writer, stream stream.Stream, contentType string) error {
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

func (c StudioClient) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	if progressBar == nil || length < 10*1024*1024 {
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

func (c StudioClient) httpClientSettings() network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		c.debug,
		c.settings.OperationId,
		c.settings.Header,
		c.settings.Timeout,
		c.settings.MaxAttempts,
		c.settings.Insecure)
}

func (c StudioClient) newUriBuilder(path string) *converter.UriBuilder {
	baseUri := c.baseUri
	if baseUri.Path == "" {
		baseUri.Path = "/{organization}/studio_/backend"
	}
	return converter.NewUriBuilder(baseUri, path).
		FormatPath("organization", c.organization)
}

func (c StudioClient) toAuthorization(token *auth.AuthToken) *network.Authorization {
	if token == nil {
		return nil
	}
	return network.NewAuthorization(token.Type, token.Value)
}

type publishSolutionRequestJson struct {
	SolutionId string `json:"solutionId"`
}

// PushSolutionResponse is the response from pushing a solution.
type PushSolutionResponse struct {
	SolutionId string `json:"solutionId"`
	Status     string `json:"status"`
}

// PublishSolutionResponse is the response from publishing a solution.
type PublishSolutionResponse struct {
	RequestId string `json:"requestId"`
	Status    string `json:"status"`
}

// SolutionInfo describes a solution returned from List.
type SolutionInfo struct {
	SolutionId string `json:"solutionId"`
	Name       string `json:"name"`
	Status     string `json:"status"`
}

// NewStudioClient creates a new Studio Web API client.
func NewStudioClient(
	baseUri url.URL,
	organization string,
	token *auth.AuthToken,
	debug bool,
	settings plugin.ExecutionSettings,
	logger log.Logger,
) *StudioClient {
	return &StudioClient{
		baseUri,
		organization,
		token,
		debug,
		settings,
		logger,
	}
}
