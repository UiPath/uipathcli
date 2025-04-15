package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The UploadCommand is a custom command for the orchestrator service which makes uploading
// files more convenient. It provides a wrapper over retrieving the write url and actually
// performing the upload.
type UploadCommand struct{}

func (c UploadCommand) Command() plugin.Command {
	return *plugin.NewCommand("orchestrator").
		WithCategory("buckets", "Orchestrator Buckets", "Buckets provide a per-folder storage solution for RPA developers to leverage in creating automation projects.").
		WithOperation("upload", "Upload file", "Uploads the provided file to the bucket").
		WithParameter("folder-id", plugin.ParameterTypeInteger, "Folder/OrganizationUnit Id", true).
		WithParameter("key", plugin.ParameterTypeInteger, "The Bucket Id", true).
		WithParameter("path", plugin.ParameterTypeString, "The BlobFile full path", true).
		WithParameter("file", plugin.ParameterTypeBinary, "The file to upload", true)
}

func (c UploadCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	writeUrl, err := c.getWriteUrl(ctx, logger)
	if err != nil {
		return err
	}
	return c.upload(ctx, logger, writeUrl)
}

func (c UploadCommand) upload(ctx plugin.ExecutionContext, logger log.Logger, url string) error {
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()
	context, cancel := context.WithCancelCause(context.Background())
	request := c.createUploadRequest(ctx, url, uploadBar, cancel)
	client := network.NewHttpClient(logger, c.httpClientSettings(ctx))
	response, err := client.SendWithContext(request, context)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusCreated {
		return fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return nil
}

func (c UploadCommand) createUploadRequest(ctx plugin.ExecutionContext, url string, uploadBar *visualization.ProgressBar, cancel context.CancelCauseFunc) *network.HttpRequest {
	file := ctx.Input
	if file == nil {
		file = c.getFileParameter(ctx.Parameters)
	}
	bodyReader, bodyWriter := io.Pipe()
	contentType, contentLength := c.writeBody(bodyWriter, file, cancel)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, contentLength, uploadBar)

	header := http.Header{
		"Content-Type":   {contentType},
		"x-ms-blob-type": {"BlockBlob"},
	}
	return network.NewHttpPutRequest(url, nil, header, uploadReader, contentLength)
}

func (c UploadCommand) writeBody(bodyWriter *io.PipeWriter, input stream.Stream, cancel context.CancelCauseFunc) (string, int64) {
	go func() {
		defer bodyWriter.Close()
		data, err := input.Data()
		if err != nil {
			cancel(err)
			return
		}
		defer data.Close()
		_, err = io.Copy(bodyWriter, data)
		if err != nil {
			cancel(err)
			return
		}
	}()
	size, _ := input.Size()
	return "application/octet-stream", size
}

func (c UploadCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c UploadCommand) getWriteUrl(ctx plugin.ExecutionContext, logger log.Logger) (string, error) {
	if ctx.Organization == "" {
		return "", errors.New("Organization is not set")
	}
	if ctx.Tenant == "" {
		return "", errors.New("Tenant is not set")
	}
	folderId := c.getIntParameter("folder-id", ctx.Parameters)
	bucketId := c.getIntParameter("key", ctx.Parameters)
	path := c.getStringParameter("path", ctx.Parameters)

	baseUri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant)
	client := api.NewOrchestratorClient(baseUri, ctx.Auth.Token, ctx.Debug, ctx.Settings, logger)
	return client.GetWriteUrl(folderId, bucketId, path)
}

func (c UploadCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/orchestrator_"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c UploadCommand) getStringParameter(name string, parameters []plugin.ExecutionParameter) string {
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

func (c UploadCommand) getIntParameter(name string, parameters []plugin.ExecutionParameter) int {
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

func (c UploadCommand) getFileParameter(parameters []plugin.ExecutionParameter) stream.Stream {
	var result stream.Stream
	for _, p := range parameters {
		if p.Name == "file" {
			if stream, ok := p.Value.(stream.Stream); ok {
				result = stream
				break
			}
		}
	}
	return result
}

func (c UploadCommand) httpClientSettings(ctx plugin.ExecutionContext) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		ctx.Debug,
		ctx.Settings.OperationId,
		ctx.Settings.Timeout,
		ctx.Settings.MaxAttempts,
		ctx.Settings.Insecure)
}

func NewUploadCommand() *UploadCommand {
	return &UploadCommand{}
}
