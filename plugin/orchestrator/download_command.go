package orchestrator

import (
	"bytes"
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
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The DownloadCommand is a custom command for the orchestrator service which makes downloading
// files more convenient. It provides a wrapper over retrieving the read url and actually
// performing the download.
type DownloadCommand struct{}

func (c DownloadCommand) Command() plugin.Command {
	return *plugin.NewCommand("orchestrator").
		WithCategory("buckets", "Orchestrator Buckets", "Buckets provide a per-folder storage solution for RPA developers to leverage in creating automation projects.").
		WithOperation("download", "Download file", "Downloads the file with the given path from the bucket").
		WithParameter("folder-id", plugin.ParameterTypeInteger, "Folder/OrganizationUnit Id", true).
		WithParameter("key", plugin.ParameterTypeInteger, "The Bucket Id", true).
		WithParameter("path", plugin.ParameterTypeString, "The BlobFile full path", true)
}

func (c DownloadCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	writeUrl, err := c.getReadUrl(ctx, logger)
	if err != nil {
		return err
	}
	return c.download(ctx, writer, logger, writeUrl)
}

func (c DownloadCommand) download(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger, url string) error {
	request := network.NewHttpGetRequest(url, http.Header{})
	client := network.NewHttpClient(logger, c.httpClientSettings(ctx))
	response, err := client.Send(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	downloadBar := visualization.NewProgressBar(logger)
	downloadReader := c.progressReader("downloading...", "completing    ", response.Body, response.ContentLength, downloadBar)
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

func (c DownloadCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c DownloadCommand) getReadUrl(ctx plugin.ExecutionContext, logger log.Logger) (string, error) {
	if ctx.Organization == "" {
		return "", errors.New("Organization is not set")
	}
	if ctx.Tenant == "" {
		return "", errors.New("Tenant is not set")
	}
	request := c.createReadUrlRequest(ctx)
	client := network.NewHttpClient(logger, c.httpClientSettings(ctx))
	response, err := client.Send(request)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %w", err)
	}
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

func (c DownloadCommand) createReadUrlRequest(ctx plugin.ExecutionContext) *network.HttpRequest {
	folderId := c.getIntParameter("folder-id", ctx.Parameters)
	bucketId := c.getIntParameter("key", ctx.Parameters)
	path := c.getStringParameter("path", ctx.Parameters)

	uri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant) + fmt.Sprintf("/odata/Buckets(%d)/UiPath.Server.Configuration.OData.GetReadUri?path=%s", bucketId, path)
	header := http.Header{
		"X-UiPath-OrganizationUnitId": {fmt.Sprintf("%d", folderId)},
	}
	for key, value := range ctx.Auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpGetRequest(uri, header)
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

func (c DownloadCommand) getStringParameter(name string, parameters []plugin.ExecutionParameter) string {
	result := ""
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(string); ok {
				result = data
				break
			}
		}
	}
	return result
}

func (c DownloadCommand) getIntParameter(name string, parameters []plugin.ExecutionParameter) int {
	result := 0
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(int); ok {
				result = data
				break
			}
		}
	}
	return result
}

func (c DownloadCommand) httpClientSettings(ctx plugin.ExecutionContext) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		ctx.Debug,
		ctx.Settings.OperationId,
		ctx.Settings.Timeout,
		ctx.Settings.MaxAttempts,
		ctx.Settings.Insecure)
}

func NewDownloadCommand() *DownloadCommand {
	return &DownloadCommand{}
}
