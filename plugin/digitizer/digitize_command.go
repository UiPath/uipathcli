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
	"github.com/UiPath/uipathcli/utils"
)

// The DigitizeCommand is a convenient wrapper over the async digitizer API
// to make it seem like it is a single sync call.
type DigitizeCommand struct{}

func (c DigitizeCommand) Command() plugin.Command {
	return *plugin.NewCommand("du").
		WithCategory("digitization", "Document Digitization").
		WithOperation("digitize", "Start digitization for the input file").
		WithParameter("file", plugin.ParameterTypeBinary, "The file to digitize", true).
		WithParameter("content-type", plugin.ParameterTypeString, "The content type", false)
}

func (c DigitizeCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	operationId, err := c.digitize(context, writer, logger)
	if err != nil {
		return err
	}

	for i := 1; i <= 60; i++ {
		finished, err := c.waitForDigitization(operationId, context, writer, logger)
		if err != nil {
			return err
		}
		if finished {
			return nil
		}
		time.Sleep(1 * time.Second)
	}
	return fmt.Errorf("Digitization with operationId '%s' did not finish in time", operationId)
}

func (c DigitizeCommand) digitize(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) (string, error) {
	uploadBar := utils.NewProgressBar(logger)
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
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}
	c.logResponse(logger, response, body)
	if response.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result digitizeResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %v", err)
	}
	return result.OperationId, nil
}

func (c DigitizeCommand) waitForDigitization(operationId string, context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) (bool, error) {
	request, err := c.createDigitizeStatusRequest(operationId, context)
	if err != nil {
		return true, err
	}
	if context.Debug {
		c.logRequest(logger, request)
	}
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return true, fmt.Errorf("Error sending request: %v", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return true, fmt.Errorf("Error reading response: %v", err)
	}
	c.logResponse(logger, response, body)
	if response.StatusCode != http.StatusOK {
		return true, fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	var result digitizeResultResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		return true, fmt.Errorf("Error parsing json response: %v", err)
	}
	if result.Status == "NotStarted" || result.Status == "Running" {
		return false, nil
	}
	err = writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body)))
	return true, err
}

func (c DigitizeCommand) createDigitizeRequest(context plugin.ExecutionContext, uploadBar *utils.ProgressBar, requestError chan error) (*http.Request, error) {
	if context.Organization == "" {
		return nil, errors.New("Organization is not set")
	}
	if context.Tenant == "" {
		return nil, errors.New("Tenant is not set")
	}
	var err error
	file := context.Input
	if file == nil {
		file, err = c.getFileParameter(context.Parameters)
		if err != nil {
			return nil, err
		}
	}
	contentType, _ := c.getParameter("content-type", context.Parameters)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	bodyReader, bodyWriter := io.Pipe()
	contentType, contentLength := c.writeMultipartBody(bodyWriter, file, contentType, requestError)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, contentLength, uploadBar)

	uri := c.formatUri(context.BaseUri, context.Organization, context.Tenant) + "/digitize/start?api-version=1"
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

func (c DigitizeCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *utils.ProgressBar) io.Reader {
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

func (c DigitizeCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/du_/api/digitizer"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c DigitizeCommand) createDigitizeStatusRequest(operationId string, context plugin.ExecutionContext) (*http.Request, error) {
	if context.Organization == "" {
		return nil, errors.New("Organization is not set")
	}
	if context.Tenant == "" {
		return nil, errors.New("Tenant is not set")
	}

	uri := c.formatUri(context.BaseUri, context.Organization, context.Tenant) + fmt.Sprintf("/digitize/result/%s?api-version=1", operationId)
	request, err := http.NewRequest("GET", uri, &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c DigitizeCommand) calculateMultipartSize(file *plugin.FileParameter) int64 {
	data, size, err := file.Data()
	if err == nil {
		defer data.Close()
	}
	return size
}

func (c DigitizeCommand) writeMultipartForm(writer *multipart.Writer, file *plugin.FileParameter, contentType string) error {
	filePart := textproto.MIMEHeader{}
	filePart.Set("Content-Disposition", fmt.Sprintf(`form-data; name="file"; filename="%s"`, file.Filename()))
	filePart.Set("Content-Type", contentType)
	w, err := writer.CreatePart(filePart)
	if err != nil {
		return fmt.Errorf("Error creating form field 'file': %v", err)
	}
	data, _, err := file.Data()
	if err != nil {
		return err
	}
	defer data.Close()
	_, err = io.Copy(w, data)
	if err != nil {
		return fmt.Errorf("Error writing form field 'file': %v", err)
	}
	return nil
}

func (c DigitizeCommand) writeMultipartBody(bodyWriter *io.PipeWriter, file *plugin.FileParameter, contentType string, errorChan chan error) (string, int64) {
	contentLength := c.calculateMultipartSize(file)
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer bodyWriter.Close()
		defer formWriter.Close()
		err := c.writeMultipartForm(formWriter, file, contentType)
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
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure},
	}
	client := &http.Client{Transport: transport}
	return client.Do(request)
}

func (c DigitizeCommand) getParameter(name string, parameters []plugin.ExecutionParameter) (string, error) {
	for _, p := range parameters {
		if p.Name == name {
			if data, ok := p.Value.(string); ok {
				return data, nil
			}
		}
	}
	return "", fmt.Errorf("Could not find '%s' parameter", name)
}

func (c DigitizeCommand) getFileParameter(parameters []plugin.ExecutionParameter) (*plugin.FileParameter, error) {
	for _, p := range parameters {
		if p.Name == "file" {
			if fileParameter, ok := p.Value.(plugin.FileParameter); ok {
				return &fileParameter, nil
			}
		}
	}
	return nil, fmt.Errorf("Could not find 'file' parameter")
}

func (c DigitizeCommand) logRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (c DigitizeCommand) logResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}
