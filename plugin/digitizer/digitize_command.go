package digitzer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/plugin"
)

type DigitizeCommand struct{}

func (c DigitizeCommand) Command() plugin.Command {
	return *plugin.NewCommand("digitizer", "digitize", "Start digitization for the input file", []plugin.CommandParameter{
		*plugin.NewCommandParameter("file", plugin.ParameterTypeBinary, "The file to digitize", true),
	}, false)
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
	request, body, err := c.createDigitizeRequest(context)
	if err != nil {
		return "", err
	}
	logger.LogRequest(*log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, body))
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}
	logger.LogResponse(*log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, responseBody))
	if response.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(responseBody))
	}
	var result digitizeResponse
	err = json.Unmarshal(responseBody, &result)
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
	logger.LogRequest(*log.NewRequestInfo(request.Method, request.URL.String(), request.Proto, request.Header, []byte{}))
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return true, fmt.Errorf("Error sending request: %v", err)
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return true, fmt.Errorf("Error reading response: %v", err)
	}
	logger.LogResponse(*log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, responseBody))
	if response.StatusCode != http.StatusOK {
		return true, fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(responseBody))
	}
	var result digitizeStatusResponse
	err = json.Unmarshal(responseBody, &result)
	if err != nil {
		return true, fmt.Errorf("Error parsing json response: %v", err)
	}
	if result.Status == "NotStarted" || result.Status == "Running" {
		return false, nil
	}
	err = writer.WriteResponse(*output.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, responseBody))
	return true, err
}

func (c DigitizeCommand) createDigitizeRequest(context plugin.ExecutionContext) (*http.Request, []byte, error) {
	org, err := c.getParameter("organization", context.Parameters)
	if err != nil {
		return nil, []byte{}, err
	}
	tenant, err := c.getParameter("tenant", context.Parameters)
	if err != nil {
		return nil, []byte{}, err
	}
	file, err := c.getFileParameter(context.Parameters)
	if err != nil {
		return nil, []byte{}, err
	}

	uri := fmt.Sprintf("%s://%s/%s/%s/du_/api/digitizer/digitize/start?api-version=1", context.BaseUri.Scheme, context.BaseUri.Host, org, tenant)
	body, contentType, err := c.createBody(*file)
	if err != nil {
		return nil, []byte{}, err
	}
	request, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		return nil, []byte{}, err
	}
	request.Header.Add("Content-Type", contentType)
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, body, nil
}

func (c DigitizeCommand) createDigitizeStatusRequest(operationId string, context plugin.ExecutionContext) (*http.Request, error) {
	org, err := c.getParameter("organization", context.Parameters)
	if err != nil {
		return nil, err
	}
	tenant, err := c.getParameter("tenant", context.Parameters)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s/%s/%s/du_/api/digitizer/digitize/result/%s?api-version=1", context.BaseUri.Scheme, context.BaseUri.Host, org, tenant, operationId)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("GET", uri, &bytes.Buffer{})
	if err != nil {
		return nil, err
	}
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
}

func (c DigitizeCommand) createBody(file plugin.FileParameter) ([]byte, string, error) {
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	w, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return []byte{}, "", fmt.Errorf("Error creating form field 'file': %v", err)
	}
	_, err = w.Write(file.Data)
	if err != nil {
		return []byte{}, "", fmt.Errorf("Error writing form field 'file': %v", err)
	}
	writer.Close()
	return b.Bytes(), writer.FormDataContentType(), nil
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
