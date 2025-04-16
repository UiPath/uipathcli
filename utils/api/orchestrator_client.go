package api

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

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

var ErrPackageAlreadyExists = errors.New("Package already exists")

type OrchestratorClient struct {
	baseUri  string
	token    *auth.AuthToken
	debug    bool
	settings plugin.ExecutionSettings
	logger   log.Logger
}

func (c OrchestratorClient) GetSharedFolderId() (int, error) {
	request := c.createGetFolderRequest()
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer func() { _ = response.Body.Close() }()
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
		return -1, errors.New("Could not find 'Shared' orchestrator folder.")
	}
	return *folderId, nil
}

func (c OrchestratorClient) createGetFolderRequest() *network.HttpRequest {
	uri := c.baseUri + "/odata/Folders"
	header := http.Header{
		"Content-Type": {"application/json"},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) findFolderId(result getFoldersResponseJson) *int {
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

func (c OrchestratorClient) Upload(file stream.Stream, uploadBar *visualization.ProgressBar) error {
	context, cancel := context.WithCancelCause(context.Background())
	request := c.createUploadRequest(file, uploadBar, cancel)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.SendWithContext(request, context)
	if err != nil {
		return err
	}
	defer func() { _ = response.Body.Close() }()
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

func (c OrchestratorClient) createUploadRequest(file stream.Stream, uploadBar *visualization.ProgressBar, cancel context.CancelCauseFunc) *network.HttpRequest {
	bodyReader, bodyWriter := io.Pipe()
	streamSize, _ := file.Size()
	contentType := c.writeMultipartBody(bodyWriter, file, "application/octet-stream", cancel)
	uploadReader := c.progressReader("uploading...", "completing  ", bodyReader, streamSize, uploadBar)

	uri := c.baseUri + "/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage"
	header := http.Header{
		"Content-Type": {contentType},
	}
	return network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, uploadReader, -1)
}

func (c OrchestratorClient) writeMultipartBody(bodyWriter *io.PipeWriter, stream stream.Stream, contentType string, cancel context.CancelCauseFunc) string {
	formWriter := multipart.NewWriter(bodyWriter)
	go func() {
		defer func() { _ = bodyWriter.Close() }()
		defer func() { _ = formWriter.Close() }()
		err := c.writeMultipartForm(formWriter, stream, contentType)
		if err != nil {
			cancel(err)
			return
		}
	}()
	return formWriter.FormDataContentType()
}

func (c OrchestratorClient) writeMultipartForm(writer *multipart.Writer, stream stream.Stream, contentType string) error {
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
	defer func() { _ = data.Close() }()
	_, err = io.Copy(w, data)
	if err != nil {
		return fmt.Errorf("Error writing form field 'file': %w", err)
	}
	return nil
}

func (c OrchestratorClient) GetReleases(folderId int, processKey string) ([]Release, error) {
	request := c.createGetReleasesRequest(folderId, processKey)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return []Release{}, err
	}
	defer func() { _ = response.Body.Close() }()
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

func (c OrchestratorClient) convertToReleases(json getReleasesResponseJson) []Release {
	releases := []Release{}
	for _, v := range json.Value {
		releases = append(releases, *NewRelease(v.Id, v.Name))
	}
	return releases
}

func (c OrchestratorClient) createGetReleasesRequest(folderId int, processKey string) *network.HttpRequest {
	uri := c.baseUri + "/odata/Releases?$filter=ProcessKey%20eq%20'" + processKey + "'"
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) CreateOrUpdateRelease(folderId int, processKey string, processVersion string) (int, error) {
	releases, err := c.GetReleases(folderId, processKey)
	if err != nil {
		return -1, err
	}
	if len(releases) > 0 {
		return c.UpdateRelease(folderId, releases[0].Id, processKey, processVersion)
	}
	return c.CreateRelease(folderId, processKey, processVersion)
}

func (c OrchestratorClient) CreateRelease(folderId int, processKey string, processVersion string) (int, error) {
	request, err := c.createNewReleaseRequest(folderId, processKey, processVersion)
	if err != nil {
		return -1, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer func() { _ = response.Body.Close() }()
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

func (c OrchestratorClient) createNewReleaseRequest(folderId int, processKey string, processVersion string) (*network.HttpRequest, error) {
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
	return network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, bytes.NewBuffer(json), -1), nil
}

func (c OrchestratorClient) UpdateRelease(folderId int, releaseId int, processKey string, processVersion string) (int, error) {
	request, err := c.createUpdateReleaseRequest(folderId, releaseId, processKey, processVersion)
	if err != nil {
		return -1, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return releaseId, nil
}

func (c OrchestratorClient) createUpdateReleaseRequest(folderId int, releaseId int, processKey string, processVersion string) (*network.HttpRequest, error) {
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
	return network.NewHttpPatchRequest(uri, c.toAuthorization(c.token), header, bytes.NewBuffer(json), -1), nil
}

func (c OrchestratorClient) CreateTestSet(folderId int, releaseId int, processVersion string) (int, error) {
	request, err := c.createTestSetRequest(folderId, releaseId, processVersion)
	if err != nil {
		return -1, err
	}
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusCreated {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return strconv.Atoi(string(body))
}

func (c OrchestratorClient) createTestSetRequest(folderId int, releaseId int, processVersion string) (*network.HttpRequest, error) {
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
	return network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, bytes.NewBuffer(json), -1), nil
}

func (c OrchestratorClient) ExecuteTestSet(folderId int, testSetId int) (int, error) {
	request := c.createExecuteTestSetRequest(folderId, testSetId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return -1, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return -1, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return -1, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}
	return strconv.Atoi(string(body))
}

func (c OrchestratorClient) createExecuteTestSetRequest(folderId int, testSetId int) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/api/TestAutomation/StartTestSetExecution?testSetId=%d&triggerType=ExternalTool", testSetId)
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	return network.NewHttpPostRequest(uri, c.toAuthorization(c.token), header, bytes.NewReader([]byte{}), -1)
}

func (c OrchestratorClient) WaitForTestExecutionToFinish(folderId int, executionId int, timeout time.Duration, statusFunc func(TestExecution)) (*TestExecution, error) {
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

func (c OrchestratorClient) GetTestSet(folderId int, testSetId int) (*TestSet, error) {
	request := c.createGetTestSetRequest(folderId, testSetId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}

	var result getTestSetResponseJson
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("Orchestrator returned invalid response body '%v'", string(body))
	}
	return c.convertToTestSet(result), nil
}

func (c OrchestratorClient) createGetTestSetRequest(folderId int, testSetId int) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/odata/TestSets(%d)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", testSetId)
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) GetTestExecution(folderId int, executionId int) (*TestExecution, error) {
	request := c.createGetTestExecutionRequest(folderId, executionId)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return nil, err
	}
	defer func() { _ = response.Body.Close() }()
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

func (c OrchestratorClient) createGetTestExecutionRequest(folderId int, executionId int) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/odata/TestSetExecutions(%d)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", executionId)
	header := http.Header{
		"Content-Type":                {"application/json"},
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) GetRobotLogs(folderId int, jobKey string) ([]RobotLog, error) {
	request := c.createGetRobotLogsRequest(folderId, jobKey)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return []RobotLog{}, err
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return []RobotLog{}, fmt.Errorf("Error reading response: %w", err)
	}
	if response.StatusCode != http.StatusOK {
		return []RobotLog{}, fmt.Errorf("Orchestrator returned status code '%v' and body '%v'", response.StatusCode, string(body))
	}

	var result getRobotLogsResponseJson
	err = json.Unmarshal(body, &result)
	if err != nil {
		return []RobotLog{}, fmt.Errorf("Orchestrator returned invalid response body '%v'", string(body))
	}
	return c.convertToRobotLogs(result), nil
}

func (c OrchestratorClient) createGetRobotLogsRequest(folderId int, jobKey string) *network.HttpRequest {
	uri := c.baseUri + "/odata/RobotLogs?$filter=JobKey%20eq%20" + jobKey
	header := http.Header{
		"X-Uipath-Organizationunitid": {strconv.Itoa(folderId)},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) GetReadUrl(folderId int, bucketId int, path string) (string, error) {
	request := c.createReadUrlRequest(folderId, bucketId, path)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return "", fmt.Errorf("Error sending request: %w", err)
	}
	defer func() { _ = response.Body.Close() }()
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

func (c OrchestratorClient) createReadUrlRequest(folderId int, bucketId int, path string) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/odata/Buckets(%d)/UiPath.Server.Configuration.OData.GetReadUri?path=%s", bucketId, path)
	header := http.Header{
		"X-UiPath-OrganizationUnitId": {strconv.Itoa(folderId)},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) GetWriteUrl(folderId int, bucketId int, path string) (string, error) {
	request := c.createWriteUrlRequest(folderId, bucketId, path)
	client := network.NewHttpClient(c.logger, c.httpClientSettings())
	response, err := client.Send(request)
	if err != nil {
		return "", err
	}
	defer func() { _ = response.Body.Close() }()
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

func (c OrchestratorClient) createWriteUrlRequest(folderId int, bucketId int, path string) *network.HttpRequest {
	uri := c.baseUri + fmt.Sprintf("/odata/Buckets(%d)/UiPath.Server.Configuration.OData.GetWriteUri?path=%s", bucketId, path)
	header := http.Header{
		"X-UiPath-OrganizationUnitId": {strconv.Itoa(folderId)},
	}
	return network.NewHttpGetRequest(uri, c.toAuthorization(c.token), header)
}

func (c OrchestratorClient) convertToRobotLogs(json getRobotLogsResponseJson) []RobotLog {
	logs := []RobotLog{}
	for _, l := range json.Value {
		logs = append(logs, c.convertToRobotLog(l))
	}
	return logs
}

func (c OrchestratorClient) convertToRobotLog(json getRobotLogJson) RobotLog {
	return *NewRobotLog(
		json.Id,
		json.Level,
		json.WindowsIdentity,
		json.ProcessName,
		json.TimeStamp,
		json.Message,
		json.RobotName,
		json.HostMachineName,
		json.MachineId,
		json.MachineKey,
		json.RuntimeType,
	)
}

func (c OrchestratorClient) convertToTestSet(json getTestSetResponseJson) *TestSet {
	testCases := []TestCase{}
	for _, c := range json.TestCases {
		testCase := NewTestCase(
			c.Id,
			c.Definition.Name,
			c.Definition.PackageIdentifier,
		)
		testCases = append(testCases, *testCase)
	}
	testPackages := []TestPackage{}
	for _, p := range json.Packages {
		testPackage := NewTestPackage(
			p.PackageIdentifier,
			p.VersionMask,
		)
		testPackages = append(testPackages, *testPackage)
	}
	return NewTestSet(json.Id, testCases, testPackages)
}

func (c OrchestratorClient) convertToTestExecution(json getTestExecutionResponseJson) *TestExecution {
	return NewTestExecution(
		json.Id,
		json.Status,
		json.TestSetId,
		json.Name,
		json.StartTime,
		json.EndTime,
		c.convertToTestCaseExecutions(json.TestCaseExecutions),
	)
}

func (c OrchestratorClient) convertToTestCaseExecutions(json []testCaseExecutionJson) []TestCaseExecution {
	executions := []TestCaseExecution{}
	for _, v := range json {
		execution := NewTestCaseExecution(
			v.Id,
			v.Status,
			v.TestCaseId,
			v.VersionNumber,
			v.EntryPointPath,
			v.Info,
			v.StartTime,
			v.EndTime,
			v.JobId,
			v.JobKey,
			v.InputArguments,
			v.OutputArguments,
			v.DataVariationIdentifier,
			c.convertToTestCaseAssertions(v.TestCaseAssertions),
			nil,
		)
		executions = append(executions, *execution)
	}
	return executions
}

func (c OrchestratorClient) convertToTestCaseAssertions(json []testCaseExecutionAssertionJson) []TestCaseAssertion {
	assertions := []TestCaseAssertion{}
	for _, v := range json {
		assertion := NewTestCaseAssertion(v.Message, v.Succeeded)
		assertions = append(assertions, *assertion)
	}
	return assertions
}

func (c OrchestratorClient) progressReader(text string, completedText string, reader io.Reader, length int64, progressBar *visualization.ProgressBar) io.Reader {
	if progressBar == nil || length < 10*1024*1024 {
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

func (c OrchestratorClient) httpClientSettings() network.HttpClientSettings {
	return *network.NewHttpClientSettings(
		c.debug,
		c.settings.OperationId,
		c.settings.Header,
		c.settings.Timeout,
		c.settings.MaxAttempts,
		c.settings.Insecure)
}

func (c OrchestratorClient) toAuthorization(token *auth.AuthToken) *network.Authorization {
	if token == nil {
		return nil
	}
	return network.NewAuthorization(token.Type, token.Value)
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

type getTestSetResponseJson struct {
	Id        int                      `json:"Id"`
	TestCases []getTestSetTestCaseJson `json:"TestCases"`
	Packages  []getTestSetPackagesJson `json:"Packages"`
}

type getTestSetTestCaseJson struct {
	Id         int                              `json:"Id"`
	Definition getTestSetTestCaseDefinitionJson `json:"Definition"`
}

type getTestSetPackagesJson struct {
	PackageIdentifier string `json:"PackageIdentifier"`
	VersionMask       string `json:"VersionMask"`
}

type getTestSetTestCaseDefinitionJson struct {
	Name              string `json:"Name"`
	PackageIdentifier string `json:"PackageIdentifier"`
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
	Id                      int                              `json:"Id"`
	Status                  string                           `json:"Status"`
	TestCaseId              int                              `json:"TestCaseId"`
	VersionNumber           string                           `json:"VersionNumber"`
	EntryPointPath          string                           `json:"EntryPointPath"`
	Info                    string                           `json:"Info"`
	StartTime               time.Time                        `json:"StartTime"`
	EndTime                 time.Time                        `json:"EndTime"`
	JobId                   int                              `json:"JobId"`
	JobKey                  string                           `json:"JobKey"`
	InputArguments          string                           `json:"InputArguments"`
	OutputArguments         string                           `json:"OutputArguments"`
	DataVariationIdentifier string                           `json:"DataVariationIdentifier"`
	TestCaseAssertions      []testCaseExecutionAssertionJson `json:"TestCaseAssertions"`
}

type testCaseExecutionAssertionJson struct {
	Message   string `json:"Message"`
	Succeeded bool   `json:"Succeeded"`
}

type getRobotLogsResponseJson struct {
	Value []getRobotLogJson `json:"value"`
}

type getRobotLogJson struct {
	Id              int       `json:"Id"`
	Level           string    `json:"Level"`
	WindowsIdentity string    `json:"WindowsIdentity"`
	ProcessName     string    `json:"ProcessName"`
	TimeStamp       time.Time `json:"TimeStamp"`
	Message         string    `json:"Message"`
	RobotName       string    `json:"RobotName"`
	HostMachineName string    `json:"HostMachineName"`
	MachineId       int       `json:"MachineId"`
	MachineKey      string    `json:"MachineKey"`
	RuntimeType     string    `json:"RuntimeType"`
}

type urlResponse struct {
	Uri string `json:"Uri"`
}

func NewOrchestratorClient(baseUri string, token *auth.AuthToken, debug bool, settings plugin.ExecutionSettings, logger log.Logger) *OrchestratorClient {
	return &OrchestratorClient{baseUri, token, debug, settings, logger}
}
