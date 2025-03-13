package digitzer

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The DigitizeCommand is a convenient wrapper over the async digitizer API
// to make it seem like it is a single sync call.
type DigitizeCommand struct{}

func (c DigitizeCommand) Command() plugin.Command {
	return *plugin.NewCommand("du").
		WithCategory("digitization", "Document Digitization", "Digitizes a document, extracting its Document Object Model (DOM) and text.").
		WithOperation("digitize", "Digitize file", "Digitize the given file").
		WithParameter("project-id", plugin.ParameterTypeString, "The project id", false).
		WithParameter("file", plugin.ParameterTypeBinary, "The file to digitize", true).
		WithParameter("content-type", plugin.ParameterTypeString, "The content type", false)
}

func (c DigitizeCommand) Execute(ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Organization == "" {
		return errors.New("Organization is not set")
	}
	if ctx.Tenant == "" {
		return errors.New("Tenant is not set")
	}
	documentId, err := c.startDigitization(ctx, logger)
	if err != nil {
		return err
	}

	for i := 1; i <= 60; i++ {
		finished, err := c.waitForDigitization(documentId, ctx, writer, logger)
		if err != nil {
			return err
		}
		if finished {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Digitization with documentId '%s' did not finish in time", documentId)
}

func (c DigitizeCommand) startDigitization(ctx plugin.ExecutionContext, logger log.Logger) (string, error) {
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()
	context, cancel := context.WithCancelCause(context.Background())
	request := c.createDigitizeRequest(ctx, uploadBar, cancel)
	client := network.NewHttpClient(logger, c.httpClientSettings(ctx))
	response, err := client.SendWithContext(request, context)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
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

func (c DigitizeCommand) waitForDigitization(documentId string, ctx plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) (bool, error) {
	request := c.createDigitizeStatusRequest(documentId, ctx)
	client := network.NewHttpClient(logger, c.httpClientSettings(ctx))
	response, err := client.Send(request)
	if err != nil {
		return true, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return true, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return true, fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result digitizeResultResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return true, fmt.Errorf("Error parsing json response: %w", err)
	}
	if result.Status == "NotStarted" || result.Status == "Running" {
		return false, nil
	}
	err = writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body)))
	return true, err
}

func (c DigitizeCommand) createDigitizeRequest(ctx plugin.ExecutionContext, uploadBar *visualization.ProgressBar, cancel context.CancelCauseFunc) *network.HttpRequest {
	projectId := c.getProjectId(ctx.Parameters)

	file := ctx.Input
	if file == nil {
		file = c.getFileParameter(ctx.Parameters)
	}
	contentType := c.getParameter("content-type", ctx.Parameters)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	streamSize, _ := file.Size()
	bodyReader, bodyWriter := io.Pipe()
	formDataContentType := c.writeMultipartBody(bodyWriter, file, contentType, cancel)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, streamSize, uploadBar)

	uri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant, projectId) + "/digitization/start?api-version=1"
	header := http.Header{
		"Content-Type": {formDataContentType},
	}
	for key, value := range ctx.Auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpPostRequest(uri, header, uploadReader, -1)
}

func (c DigitizeCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c DigitizeCommand) formatUri(baseUri url.URL, org string, tenant string, projectId string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/du_/api/framework/projects/{projectId}"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.ReplaceAll(path, "{projectId}", projectId)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c DigitizeCommand) createDigitizeStatusRequest(documentId string, ctx plugin.ExecutionContext) *network.HttpRequest {
	projectId := c.getProjectId(ctx.Parameters)
	uri := c.formatUri(ctx.BaseUri, ctx.Organization, ctx.Tenant, projectId) + fmt.Sprintf("/digitization/result/%s?api-version=1", documentId)
	header := http.Header{}
	for key, value := range ctx.Auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpGetRequest(uri, header)
}

func (c DigitizeCommand) writeMultipartForm(writer *multipart.Writer, stream stream.Stream, contentType string) error {
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
	defer data.Close()
	_, err = io.Copy(w, data)
	if err != nil {
		return fmt.Errorf("Error writing form field 'file': %w", err)
	}
	return nil
}

func (c DigitizeCommand) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, cancel context.CancelCauseFunc) string {
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer bodyWriter.Close()
		defer formWriter.Close()
		err := c.writeMultipartForm(formWriter, stream, contentType)
		if err != nil {
			cancel(err)
			return
		}
	}()
	return formWriter.FormDataContentType()
}

func (c DigitizeCommand) getProjectId(parameters []plugin.ExecutionParameter) string {
	projectId := c.getParameter("project-id", parameters)
	if projectId == "" {
		projectId = "00000000-0000-0000-0000-000000000000"
	}
	return projectId
}

func (c DigitizeCommand) getParameter(name string, parameters []plugin.ExecutionParameter) string {
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

func (c DigitizeCommand) getFileParameter(parameters []plugin.ExecutionParameter) stream.Stream {
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

func (c DigitizeCommand) httpClientSettings(ctx plugin.ExecutionContext) network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		ctx.Debug,
		ctx.Settings.OperationId,
		ctx.Settings.Timeout,
		ctx.Settings.MaxAttempts,
		ctx.Settings.Insecure)
}

func NewDigitizeCommand() *DigitizeCommand {
	return &DigitizeCommand{}
}
