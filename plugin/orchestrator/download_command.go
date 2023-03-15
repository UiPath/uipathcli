package orchestrator

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils"
)

// The DownloadCommand is a custom command for the orchestrator service which makes downloading
// files more convenient. It provides a wrapper over retrieving the read url and actually
// performing the download.
type DownloadCommand struct{}

func (c DownloadCommand) Command() plugin.Command {
	return *plugin.NewCommand("orchestrator").
		WithCategory("buckets", "Orchestrator Buckets").
		WithOperation("download", "Downloads the file with the given path from the bucket").
		WithParameter("folder-id", plugin.ParameterTypeInteger, "Folder/OrganizationUnit Id", true).
		WithParameter("key", plugin.ParameterTypeInteger, "The Bucket Id", true).
		WithParameter("path", plugin.ParameterTypeString, "The BlobFile full path", true)
}

func (c DownloadCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	writeUrl, err := c.getReadUrl(context, logger)
	if err != nil {
		return err
	}
	return c.download(context, writer, logger, writeUrl)
}

func (c DownloadCommand) download(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger, url string) error {
	requestError := make(chan error)
	request, err := http.NewRequest("GET", url, &bytes.Buffer{})
	if err != nil {
		return err
	}
	if context.Debug {
		c.logRequest(logger, request)
	}
	response, err := c.send(request, context.Insecure, requestError)
	if err != nil {
		return fmt.Errorf("Error sending request: %w", err)
	}
	defer response.Body.Close()
	downloadBar := utils.NewProgressBar(logger)
	downloadReader := c.progressReader("downloading...", "completing    ", response.Body, response.ContentLength, downloadBar)
	defer downloadBar.Remove()
	body, err := io.ReadAll(downloadReader)
	if err != nil {
		return fmt.Errorf("Error reading response body: %w", err)
	}
	c.logResponse(logger, response, body)
	err = writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body)))
	if err != nil {
		return err
	}
	return nil
}

func (c DownloadCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *utils.ProgressBar) io.Reader {
	if length < 10*1024*1024 {
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

func (c DownloadCommand) getReadUrl(context plugin.ExecutionContext, logger log.Logger) (string, error) {
	request, err := c.createReadUrlRequest(context)
	if err != nil {
		return "", err
	}
	if context.Debug {
		c.logRequest(logger, request)
	}
	requestError := make(chan error)
	response, err := c.send(request, context.Insecure, requestError)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %w", err)
	}
	c.logResponse(logger, response, body)
	if response.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result urlResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %w", err)
	}
	return result.Uri, nil
}

func (c DownloadCommand) createReadUrlRequest(context plugin.ExecutionContext) (*http.Request, error) {
	if context.Organization == "" {
		return nil, errors.New("Organization is not set")
	}
	if context.Tenant == "" {
		return nil, errors.New("Tenant is not set")
	}
	folderId, err := c.getIntParameter("folder-id", context.Parameters)
	if err != nil {
		return nil, err
	}
	bucketId, err := c.getIntParameter("key", context.Parameters)
	if err != nil {
		return nil, err
	}
	path, err := c.getStringParameter("path", context.Parameters)
	if err != nil {
		return nil, err
	}

	uri := c.formatUri(context.BaseUri, context.Organization, context.Tenant) + fmt.Sprintf("/odata/Buckets(%d)/UiPath.Server.Configuration.OData.GetReadUri?path=%s", bucketId, path)
	request, err := http.NewRequest("GET", uri, &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	request.Header.Add("X-UiPath-OrganizationUnitId", fmt.Sprintf("%d", folderId))
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c DownloadCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/orchestrator_"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c DownloadCommand) send(request *http.Request, insecure bool, errorChan chan error) (*http.Response, error) {
	responseChan := make(chan *http.Response)
	go func(request *http.Request) {
		response, err := c.sendRequest(request, insecure)
		if err != nil {
			errorChan <- err
			return
		}
		responseChan <- response
	}(request)

	select {
	case err := <-errorChan:
		return nil, err
	case response := <-responseChan:
		return response, nil
	}
}

func (c DownloadCommand) sendRequest(request *http.Request, insecure bool) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint // This is user configurable and disabled by default
	}
	client := &http.Client{Transport: transport}
	return client.Do(request)
}

func (c DownloadCommand) getStringParameter(name string, parameters []plugin.ExecutionParameter) (string, error) {
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(string); ok {
				return data, nil
			}
		}
	}
	return "", fmt.Errorf("Could not find '%s' parameter", name)
}

func (c DownloadCommand) getIntParameter(name string, parameters []plugin.ExecutionParameter) (int, error) {
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(int); ok {
				return data, nil
			}
		}
	}
	return 0, fmt.Errorf("Could not find '%s' parameter", name)
}

func (c DownloadCommand) logRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	_, _ = buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (c DownloadCommand) logResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}
