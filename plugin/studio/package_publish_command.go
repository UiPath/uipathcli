package studio

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
	"os"
	"path/filepath"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// The PackagePublishCommand publishes a package
type PackagePublishCommand struct {
}

func (c PackagePublishCommand) Command() plugin.Command {
	return *plugin.NewCommand("studio").
		WithCategory("package", "Package", "UiPath Studio package-related actions").
		WithOperation("publish", "Publish Package", "Publishes the package to orchestrator").
		WithParameter("source", plugin.ParameterTypeString, "Path to package (default: .)", false)
}

func (c PackagePublishCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if context.Organization == "" {
		return errors.New("Organization is not set")
	}
	if context.Tenant == "" {
		return errors.New("Tenant is not set")
	}
	source, err := c.getSource(context)
	if err != nil {
		return err
	}
	nupkgReader := newNupkgReader(source)
	nuspec, err := nupkgReader.ReadNuspec()
	if err != nil {
		return err
	}
	baseUri := c.formatUri(context.BaseUri, context.Organization, context.Tenant)
	params := newPackagePublishParams(source, nuspec.Title, nuspec.Version, baseUri, context.Auth, context.Insecure, context.Debug)
	result, err := c.publish(*params, logger)
	if err != nil {
		return err
	}

	json, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("Publish command failed: %v", err)
	}
	return writer.WriteResponse(*output.NewResponseInfo(200, "200 OK", "HTTP/1.1", map[string][]string{}, bytes.NewReader(json)))
}

func (c PackagePublishCommand) publish(params packagePublishParams, logger log.Logger) (*packagePublishResult, error) {
	file := stream.NewFileStream(params.Source)
	uploadBar := visualization.NewProgressBar(logger)
	defer uploadBar.Remove()
	requestError := make(chan error)
	request, err := c.createUploadRequest(file, params, uploadBar, requestError)
	if err != nil {
		return nil, err
	}
	if params.Debug {
		c.logRequest(logger, request)
	}
	response, err := c.send(request, params.Insecure, requestError)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %w", err)
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	c.logResponse(logger, response, body)
	if response.StatusCode == http.StatusConflict {
		errorMessage := fmt.Sprintf("Package '%s' already exists", filepath.Base(params.Source))
		return newFailedPackagePublishResult(errorMessage, params.Source, params.Name, params.Version), nil
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return newSucceededPackagePublishResult(params.Source, params.Name, params.Version), nil
}

func (c PackagePublishCommand) createUploadRequest(file stream.Stream, params packagePublishParams, uploadBar *visualization.ProgressBar, requestError chan error) (*http.Request, error) {
	bodyReader, bodyWriter := io.Pipe()
	contentType, contentLength := c.writeMultipartBody(bodyWriter, file, "application/octet-stream", requestError)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, contentLength, uploadBar)

	uri := params.BaseUri + "/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage"
	request, err := http.NewRequest("POST", uri, uploadReader)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", contentType)
	for key, value := range params.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c PackagePublishCommand) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c PackagePublishCommand) calculateMultipartSize(stream stream.Stream) int64 {
	size, _ := stream.Size()
	return size
}

func (c PackagePublishCommand) writeMultipartForm(writer *multipart.Writer, stream stream.Stream, contentType string) error {
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

func (c PackagePublishCommand) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, errorChan chan error) (string, int64) {
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

func (c PackagePublishCommand) send(request *http.Request, insecure bool, errorChan chan error) (*http.Response, error) {
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

func (c PackagePublishCommand) sendRequest(request *http.Request, insecure bool) (*http.Response, error) {
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: insecure}, //nolint // This is user configurable and disabled by default
	}
	client := &http.Client{Transport: transport}
	return client.Do(request)
}

func (c PackagePublishCommand) getSource(context plugin.ExecutionContext) (string, error) {
	source := c.getParameter("source", ".", context.Parameters)
	source, _ = filepath.Abs(source)
	fileInfo, err := os.Stat(source)
	if err != nil {
		return "", fmt.Errorf("Package not found.")
	}
	if fileInfo.IsDir() {
		source = findLatestNupkg(source)
	}
	if source == "" {
		return "", errors.New("Could not find package to publish")
	}
	return source, nil
}

func (c PackagePublishCommand) getParameter(name string, defaultValue string, parameters []plugin.ExecutionParameter) string {
	result := defaultValue
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

func (c PackagePublishCommand) formatUri(baseUri url.URL, org string, tenant string) string {
	path := baseUri.Path
	if baseUri.Path == "" {
		path = "/{organization}/{tenant}/orchestrator_"
	}
	path = strings.ReplaceAll(path, "{organization}", org)
	path = strings.ReplaceAll(path, "{tenant}", tenant)
	path = strings.TrimSuffix(path, "/")
	return fmt.Sprintf("%s://%s%s", baseUri.Scheme, baseUri.Host, path)
}

func (c PackagePublishCommand) logRequest(logger log.Logger, request *http.Request) {
	buffer := &bytes.Buffer{}
	_, _ = buffer.ReadFrom(request.Body)
	body := buffer.Bytes()
	request.Body = io.NopCloser(bytes.NewReader(body))
	requestInfo := log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, bytes.NewReader(body))
	logger.LogRequest(*requestInfo)
}

func (c PackagePublishCommand) logResponse(logger log.Logger, response *http.Response, body []byte) {
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, bytes.NewReader(body))
	logger.LogResponse(*responseInfo)
}

func NewPackagePublishCommand() *PackagePublishCommand {
	return &PackagePublishCommand{}
}
