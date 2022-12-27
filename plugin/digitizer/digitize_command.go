package plugin_digitizer

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils"
)

type DigitizeCommand struct{}

func (c DigitizeCommand) Command() plugin.Command {
	return *plugin.NewCommand("digitizer", "digitize", "Start digitization for the input file", []plugin.CommandParameter{
		*plugin.NewCommandParameter("file", plugin.ParameterTypeBinary, "The file to digitize", true),
	}, false)
}

func (c DigitizeCommand) Execute(context plugin.ExecutionContext, output io.Writer) error {
	logger := utils.HttpLogger{
		Output: output,
		Debug:  context.Debug,
	}
	operationId, err := c.digitize(context, &logger)
	if err != nil {
		return err
	}

	for i := 1; i <= 60; i++ {
		finished, err := c.waitForDigitization(operationId, context, &logger)
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

func (c DigitizeCommand) digitize(context plugin.ExecutionContext, logger *utils.HttpLogger) (string, error) {
	request, err := c.createDigitizeRequest(context)
	if err != nil {
		return "", err
	}
	if context.Debug {
		err = logger.LogRequest(request)
		if err != nil {
			return "", err
		}
	}
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %v", err)
	}
	if context.Debug {
		err = logger.LogResponse(response)
		if err != nil {
			return "", err
		}
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", fmt.Errorf("Error reading response: %v", err)
	}
	if response.StatusCode != http.StatusAccepted {
		return "", fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(data))
	}
	var result digitizeResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return "", fmt.Errorf("Error parsing json response: %v", err)
	}
	return result.OperationId, nil
}

func (c DigitizeCommand) waitForDigitization(operationId string, context plugin.ExecutionContext, logger *utils.HttpLogger) (bool, error) {
	request, err := c.createDigitizeStatusRequest(operationId, context)
	if err != nil {
		return true, err
	}
	if context.Debug {
		err = logger.LogRequest(request)
		if err != nil {
			return true, err
		}
	}
	response, err := c.sendRequest(request, context.Insecure)
	if err != nil {
		return true, fmt.Errorf("Error sending request: %v", err)
	}
	if context.Debug {
		err = logger.LogResponse(response)
		if err != nil {
			return true, err
		}
	}
	defer response.Body.Close()
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return true, fmt.Errorf("Error reading response: %v", err)
	}
	if response.StatusCode != http.StatusOK {
		return true, fmt.Errorf("Digitizer returned status code '%v' and body '%v'", response.StatusCode, string(data))
	}
	var result digitizeStatusResponse
	err = json.Unmarshal(data, &result)
	if err != nil {
		return true, fmt.Errorf("Error parsing json response: %v", err)
	}
	if result.Status == "NotStarted" || result.Status == "Running" {
		return false, nil
	}
	if !context.Debug {
		logger.LogBody(bytes.NewBuffer(data))
	}
	return true, nil
}

func (c DigitizeCommand) createDigitizeRequest(context plugin.ExecutionContext) (*http.Request, error) {
	org, err := c.getParameter("organization", context.Parameters)
	if err != nil {
		return nil, err
	}
	tenant, err := c.getParameter("tenant", context.Parameters)
	if err != nil {
		return nil, err
	}
	file, err := c.getFileParameter(context.Parameters)
	if err != nil {
		return nil, err
	}

	uri := fmt.Sprintf("%s://%s/%s/%s/du_/api/digitizer/digitize/start?api-version=1", context.BaseUri.Scheme, context.BaseUri.Host, org, tenant)
	body, contentType, err := c.createBody(*file)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest("POST", uri, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Add("Content-Type", contentType)
	for key, value := range context.Auth.Header {
		request.Header.Add(key, value)
	}
	return request, nil
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
