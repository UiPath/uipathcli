package digitzer

import (
	"bytes"
	"crypto/tls"
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

func (c DigitizeCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if context.Organization == "" {
		return errors.New("Organization is not set")
	}
	if context.Tenant == "" {
		return errors.New("Tenant is not set")
	}
	documentId, err := c.startDigitization(context, logger)
	if err != nil {
		return err
	}

	for i := 1; i <= 60; i++ {
		finished, err := c.waitForDigitization(documentId, context, writer, logger)
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

func (c DigitizeCommand) startDigitization(context plugin.ExecutionContext, logger log.Logger) (string, error) {
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()
	requestError := make(chan error)
	request, err := c.createDigitizeRequest(context, uploadBar, requestError)
	if err != nil {
		return "", err
	}
	if context.Debug {
		c.logRequest(logger, request)
	}
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

func (c DigitizeCommand) waitForDigitization(documentId string, context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) (bool, error) {
	request, err := c.createDigitizeStatusRequest(documentId, context)
	if err != nil {
		return true, err
	}
	if context.Debug {
		c.logRequest(logger, request)
	}
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return true, fmt.Errorf("Error sending request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return true, fmt.Errorf("Error reading response: %w", err)
	}
	c.logResponse(logger, response, body)
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

func (c DigitizeCommand) createDigitizeRequest(context plugin.ExecutionContext, uploadBar *visualization.ProgressBar, requestError chan error) (*http.Request, error) {
	projectId := c.getProjectId(context.Parameters)

	var err error
	file := context.Input
	if file == nil {
		file = c.getFileParameter(context.Parameters)
	}
	contentType := c.getParameter("content-type", context.Parameters)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	bodyReader, bodyWriter := io.Pipe()
	contentType, contentLength := c.writeMultipartBody(bodyWriter, file, contentType, requestError)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, contentLength, uploadBar)

	uri := c.formatUri(context.BaseUri, context.Organization, context.Tenant, projectId) + "/digitization/start?api-version=1"
	request, err := http.NewRequest("POST", uri, uploadReader)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", contentType)
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c DigitizeCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c DigitizeCommand) createDigitizeStatusRequest(documentId string, context plugin.ExecutionContext) (*http.Request, error) {
	projectId := c.getProjectId(context.Parameters)
	uri := c.formatUri(context.BaseUri, context.Organization, context.Tenant, projectId) + fmt.Sprintf("/digitization/result/%s?api-version=1", documentId)
	request, err := http.NewRequest("GET", uri, &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c DigitizeCommand) calculateMultipartSize(stream stream.Stream) int64 {
	size, _ := stream.Size()
	return size
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

func (c DigitizeCommand) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, errorChan chan error) (string, int64) {
	contentLength := c.calculateMultipartSize(stream)
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer bodyWriter.Close()
		defer formWriter.Close()
		err := c.writeMultipartForm(formWriter, stream, contentType)
		if err != nil {
			errorChan <- err
			return
		}
	}()
	return formWriter.FormDataContentType(), contentLength
}

func (c DigitizeCommand) send(request *http.Request, insecure bool, errorChan chan error) (*http.Response, error) {
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

func (c DigitizeCommand) sendRequest(request *http.Request, insecure bool) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint // This is user configurable and disabled by default
	}
	client := &http.Client{Transport: transport}
	return client.Do(request)
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

func (c DigitizeCommand) logRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	_, _ = buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (c DigitizeCommand) logResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}

func NewDigitizeCommand() *DigitizeCommand {
	return &DigitizeCommand{}
}
