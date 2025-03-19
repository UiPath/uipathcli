package studio

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
	"strconv"
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

var ErrPackageAlreadyExists = errors.New("Package already exists")

type orchestratorClient struct {
	baseUri  string
	auth     plugin.AuthResult
	debug    bool
	settings plugin.ExecutionSettings
	logger   log.Logger
}

func (c orchestratorClient) GetSharedFolderId() (int, error) {
	request := c.createGetFolderRequest()
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}

	var result getFoldersResponseJson
	err = json.Unmarshal(body, &result)
	if err != nil {
		return -1, fmt.Errorf("Orchestrator returned invalid response body '%v'", string(body))
	}
	folderId := c.findFolderId(result)
	if folderId == nil {
		return -1, fmt.Errorf("Could not find 'Shared' orchestrator folder.")
	}
	return *folderId, nil
}

func (c orchestratorClient) createGetFolderRequest() *network.HttpRequest {
	uri := c.baseUri + "/odata/Folders"
	header := http.Header{
		"Content-Type": {"application/json"},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpGetRequest(uri, header)
}

func (c orchestratorClient) findFolderId(result getFoldersResponseJson) *int {
	for _, value := range result.Value {
		if value.Name == "Shared" {
			return &value.Id
		}
	}
	if len(result.Value) > 0 {
		return &result.Value[0].Id
	}
	return nil
}

func (c orchestratorClient) Upload(file stream.Stream, uploadBar *visualization.ProgressBar) error {
	context, cancel := context.WithCancelCause(context.Background())
	request := c.createUploadRequest(file, uploadBar, cancel)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.SendWithContext(request, context)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode == http.StatusConflict {
		return ErrPackageAlreadyExists
	}
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return nil
}

func (c orchestratorClient) createUploadRequest(file stream.Stream, uploadBar *visualization.ProgressBar, cancel context.CancelCauseFunc) *network.HttpRequest {
	bodyReader, bodyWriter := io.Pipe()
	streamSize, _ := file.Size()
	contentType := c.writeMultipartBody(bodyWriter, file, "application/octet-stream", cancel)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, streamSize, uploadBar)

	uri := c.baseUri + "/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage"
	header := http.Header{
		"Content-Type": {contentType},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpPostRequest(uri, header, uploadReader, -1)
}

func (c orchestratorClient) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, cancel context.CancelCauseFunc) string {
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

func (c orchestratorClient) writeMultipartForm(writer *multipart.Writer, stream stream.Stream, contentType string) error {
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

func (c orchestratorClient) GetReleases(folderId int, processKey string) ([]Release, error) {
	request := c.createGetReleasesRequest(folderId, processKey)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return []Release{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []Release{}, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return []Release{}, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}

	var result getReleasesResponseJson
	err = json.Unmarshal(body, &result)
	if err != nil {
		return []Release{}, fmt.Errorf("Orchestrator returned invalid response body '%v'", string(body))
	}
	return c.convertToReleases(result), nil
}

func (c orchestratorClient) convertToReleases(json getReleasesResponseJson) []Release {
	releases := []Release{}
	for _, v := range json.Value {
		releases = append(releases, *NewRelease(v.Id, v.Name))
	}
	return releases
}

func (c orchestratorClient) createGetReleasesRequest(folderId int, processKey string) *network.HttpRequest {
	uri := c.baseUri + "/odata/Releases?$filter=ProcessKey%20eq%20'" + processKey + "'"
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpGetRequest(uri, header)
}

func (c orchestratorClient) CreateOrUpdateRelease(folderId int, processKey string, processVersion string) (int, error) {
	releases, err := c.GetReleases(folderId, processKey)
	if err != nil {
		return -1, err
	}
	if len(releases) > 0 {
		return c.UpdateRelease(folderId, releases[0].Id, processKey, processVersion)
	}
	return c.CreateRelease(folderId, processKey, processVersion)
}

func (c orchestratorClient) CreateRelease(folderId int, processKey string, processVersion string) (int, error) {
	request, err := c.createNewReleaseRequest(folderId, processKey, processVersion)
	if err != nil {
		return -1, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusCreated {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}

	var result createReleaseResponseJson
	err = json.Unmarshal(body, &result)
	if err != nil {
		return -1, fmt.Errorf("Orchestrator returned invalid response body '%v'", string(body))
	}
	return result.Id, nil
}

func (c orchestratorClient) createNewReleaseRequest(folderId int, processKey string, processVersion string) (*network.HttpRequest, error) {
	json, err := json.Marshal(createReleaseRequestJson{
		Name:           processKey,
		ProcessKey:     processKey,
		ProcessVersion: processVersion,
	})
	if err != nil {
		return nil, err
	}
	uri := c.baseUri + "/odata/Releases"
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpPostRequest(uri, header, bytes.NewBuffer(json), -1), nil
}

func (c orchestratorClient) UpdateRelease(folderId int, releaseId int, processKey string, processVersion string) (int, error) {
	request, err := c.createUpdateReleaseRequest(folderId, releaseId, processKey, processVersion)
	if err != nil {
		return -1, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return releaseId, nil
}

func (c orchestratorClient) createUpdateReleaseRequest(folderId int, releaseId int, processKey string, processVersion string) (*network.HttpRequest, error) {
	json, err := json.Marshal(createReleaseRequestJson{
		Name:           processKey,
		ProcessKey:     processKey,
		ProcessVersion: processVersion,
	})
	if err != nil {
		return nil, err
	}
	uri := c.baseUri + fmt.Sprintf("/odata/Releases(%d)", releaseId)
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpPatchRequest(uri, header, bytes.NewBuffer(json), -1), nil
}

func (c orchestratorClient) CreateTestSet(folderId int, releaseId int, processVersion string) (int, error) {
	request, err := c.createTestSetRequest(folderId, releaseId, processVersion)
	if err != nil {
		return -1, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusCreated {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return strconv.Atoi(string(body))
}

func (c orchestratorClient) createTestSetRequest(folderId int, releaseId int, processVersion string) (*network.HttpRequest, error) {
	json, err := json.Marshal(createTestSetRequestJson{
		ReleaseId:     releaseId,
		VersionNumber: processVersion,
	})
	if err != nil {
		return nil, err
	}
	uri := c.baseUri + "/api/TestAutomation/CreateTestSetForReleaseVersion"
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpPostRequest(uri, header, bytes.NewBuffer(json), -1), nil
}

func (c orchestratorClient) ExecuteTestSet(folderId int, testSetId int) (int, error) {
	request := c.createExecuteTestSetRequest(folderId, testSetId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return strconv.Atoi(string(body))
}

func (c orchestratorClient) createExecuteTestSetRequest(folderId int, testSetId int) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/api/TestAutomation/StartTestSetExecution?testSetId=%d&triggerType=ExternalTool", testSetId)
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpPostRequest(uri, header, bytes.NewReader([]byte{}), -1)
}

func (c orchestratorClient) WaitForTestExecutionToFinish(folderId int, executionId int, timeout time.Duration, statusFunc func(TestExecution)) (*TestExecution, error) {
	startTime := time.Now()
	for {
		execution, err := c.GetTestExecution(folderId, executionId)
		if err != nil {
			return nil, err
		}
		statusFunc(*execution)
		if execution.IsCompleted() {
			return execution, nil
		}
		if time.Since(startTime) >= timeout {
			return nil, fmt.Errorf("Timeout waiting for test execution '%d' to finish.", executionId)
		}
		time.Sleep(1 * time.Second)
	}
}

func (c orchestratorClient) GetTestExecution(folderId int, executionId int) (*TestExecution, error) {
	request := c.createGetTestExecutionRequest(folderId, executionId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}

	var result getTestExecutionResponseJson
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Orchestrator returned invalid response body '%v'", string(body))
	}
	return c.convertToTestExecution(result), nil
}

func (c orchestratorClient) createGetTestExecutionRequest(folderId int, executionId int) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/odata/TestSetExecutions(%d)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", executionId)
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	for key, value := range c.auth.Header {
		header.Set(key, value)
	}
	return network.NewHttpGetRequest(uri, header)
}

func (c orchestratorClient) convertToTestExecution(json getTestExecutionResponseJson) *TestExecution {
	return NewTestExecution(
		json.Id,
		json.Status,
		json.TestSetId,
		json.Name,
		json.StartTime,
		json.EndTime,
		c.convertToTestCaseExecutions(json.TestCaseExecutions))
}

func (c orchestratorClient) convertToTestCaseExecutions(json []testCaseExecutionJson) []TestCaseExecution {
	executions := []TestCaseExecution{}
	for _, v := range json {
		execution := NewTestCaseExecution(
			v.Id,
			v.Status,
			v.TestCaseId,
			v.EntryPointPath,
			v.Info,
			v.StartTime,
			v.EndTime)
		executions = append(executions, *execution)
	}
	return executions
}

func (c orchestratorClient) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
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

func (c orchestratorClient) httpClientSettings() network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		c.debug,
		c.settings.OperationId,
		c.settings.Timeout,
		c.settings.MaxAttempts,
		c.settings.Insecure)
}

type createReleaseRequestJson struct {
	Name           string `json:"name"`
	ProcessKey     string `json:"processKey"`
	ProcessVersion string `json:"processVersion"`
}

type createTestSetRequestJson struct {
	ReleaseId     int    `json:"releaseId"`
	VersionNumber string `json:"versionNumber"`
}

type getFoldersResponseJson struct {
	Value []getFoldersResponseValueJson `json:"value"`
}

type getFoldersResponseValueJson struct {
	Id   int    `json:"Id"`
	Name string `json:"FullyQualifiedName"`
}

type getReleasesResponseJson struct {
	Value []getReleasesResponseValueJson `json:"value"`
}

type getReleasesResponseValueJson struct {
	Id   int    `json:"Id"`
	Name string `json:"Name"`
}

type createReleaseResponseJson struct {
	Id int `json:"id"`
}

type getTestExecutionResponseJson struct {
	Id                 int                     `json:"Id"`
	Status             string                  `json:"Status"`
	TestSetId          int                     `json:"TestSetId"`
	Name               string                  `json:"Name"`
	StartTime          time.Time               `json:"StartTime"`
	EndTime            time.Time               `json:"EndTime"`
	TestCaseExecutions []testCaseExecutionJson `json:"TestCaseExecutions"`
}

type testCaseExecutionJson struct {
	Id             int       `json:"Id"`
	Status         string    `json:"Status"`
	TestCaseId     int       `json:"TestCaseId"`
	EntryPointPath string    `json:"EntryPointPath"`
	Info           string    `json:"Info"`
	StartTime      time.Time `json:"StartTime"`
	EndTime        time.Time `json:"EndTime"`
}

func newOrchestratorClient(baseUri string, auth plugin.AuthResult, debug bool, settings plugin.ExecutionSettings, logger log.Logger) *orchestratorClient {
	return &orchestratorClient{baseUri, auth, debug, settings, logger}
}
