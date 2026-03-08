package api

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

func newTestClient(t *testing.T, serverURL string, token *auth.AuthToken) *StudioClient {
	t.Helper()
	baseUri, _ := url.Parse(serverURL)
	logger := log.NewDefaultLogger(io.Discard)
	settings := plugin.ExecutionSettings{
		MaxAttempts: 1,
	}
	return NewStudioClient(*baseUri, "my-org", token, false, settings, logger)
}

func TestToAuthorizationWithNilToken(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)

	result := client.toAuthorization(nil)

	if result != nil {
		t.Errorf("Expected nil authorization for nil token, but got: %v", result)
	}
}

func TestToAuthorizationWithToken(t *testing.T) {
	token := &auth.AuthToken{Type: "Bearer", Value: "test-token"}
	client := newTestClient(t, "http://localhost", token)

	result := client.toAuthorization(token)

	if result == nil {
		t.Errorf("Expected authorization for valid token, but got nil")
	}
}

func TestNewUriBuilderWithEmptyPath(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)

	builder := client.newUriBuilder("/api/v1/test")
	uri := builder.Build()

	if !strings.Contains(uri, "/my-org/studio_/backend/api/v1/test") {
		t.Errorf("Expected URI to contain default studio backend path, but got: %v", uri)
	}
}

func TestNewUriBuilderWithCustomPath(t *testing.T) {
	baseUri, _ := url.Parse("http://localhost/custom/path")
	logger := log.NewDefaultLogger(io.Discard)
	settings := plugin.ExecutionSettings{}
	client := NewStudioClient(*baseUri, "my-org", nil, false, settings, logger)

	builder := client.newUriBuilder("/api/v1/test")
	uri := builder.Build()

	if !strings.Contains(uri, "/custom/path/api/v1/test") {
		t.Errorf("Expected URI to use custom path, but got: %v", uri)
	}
}

func TestProgressReaderWithNilProgressBar(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)
	reader := strings.NewReader("test-data")

	result := client.progressReader("uploading", "done", reader, 100*1024*1024, nil)

	if result != reader {
		t.Errorf("Expected original reader when progressBar is nil")
	}
}

func TestProgressReaderWithSmallFile(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)
	reader := strings.NewReader("test-data")
	logger := log.NewDefaultLogger(io.Discard)

	result := client.progressReader("uploading", "done", reader, 100, newTestProgressBar(logger))

	if result != reader {
		t.Errorf("Expected original reader for small files")
	}
}

func TestProgressReaderWithLargeFile(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)
	reader := strings.NewReader("test-data")
	logger := log.NewDefaultLogger(io.Discard)

	result := client.progressReader("uploading", "done", reader, 20*1024*1024, newTestProgressBar(logger))

	if result == reader {
		t.Errorf("Expected wrapped reader for large files, but got original reader")
	}
	// Verify the wrapped reader still returns data
	data, err := io.ReadAll(result)
	if err != nil {
		t.Errorf("Expected no error reading from wrapped reader, but got: %v", err)
	}
	if string(data) != "test-data" {
		t.Errorf("Expected test-data, but got: %v", string(data))
	}
}

func TestListSolutionsWithNilToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`[{"solutionId":"sol-1","name":"Test","status":"active"}]`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	solutions, err := client.ListSolutions()

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if len(solutions) != 1 || solutions[0].SolutionId != "sol-1" {
		t.Errorf("Expected 1 solution with id sol-1, but got: %v", solutions)
	}
}

func TestPullSolutionWithNilToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("solution-data"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	body, err := client.PullSolution("test-id")

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	defer func() { _ = body.Close() }()
	data, _ := io.ReadAll(body)
	if string(data) != "solution-data" {
		t.Errorf("Expected solution-data, but got: %v", string(data))
	}
}

func TestPushSolutionWithNilToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"solutionId":"pushed-id","status":"ok"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	file := stream.NewMemoryStream("test.uis", []byte("test-content"))
	result, err := client.PushSolution(file, "", nil)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if result.SolutionId != "pushed-id" {
		t.Errorf("Expected solutionId pushed-id, but got: %v", result.SolutionId)
	}
}

func TestPublishSolutionWithNilToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"requestId":"req-1","status":"queued"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	result, err := client.PublishSolution("test-id")

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if result.RequestId != "req-1" {
		t.Errorf("Expected requestId req-1, but got: %v", result.RequestId)
	}
}

func TestPushSolutionWithSolutionId(t *testing.T) {
	var requestURL string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestURL = r.URL.String()
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"solutionId":"existing-id"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	file := stream.NewMemoryStream("test.uis", []byte("data"))
	_, err := client.PushSolution(file, "existing-id", nil)

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if !strings.Contains(requestURL, "solutionId=existing-id") {
		t.Errorf("Expected solutionId query param, but URL was: %v", requestURL)
	}
}

func TestPushSolutionNonJsonResponseReturnsEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json-response"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	file := stream.NewMemoryStream("test.uis", []byte("content"))
	result, err := client.PushSolution(file, "", nil)

	if err != nil {
		t.Errorf("Expected no error for non-JSON response, but got: %v", err)
	}
	if result.SolutionId != "" {
		t.Errorf("Expected empty solutionId for non-JSON response, but got: %v", result.SolutionId)
	}
}

func TestPublishSolutionNonJsonResponseReturnsEmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	result, err := client.PublishSolution("test-id")

	if err != nil {
		t.Errorf("Expected no error for non-JSON response, but got: %v", err)
	}
	if result.RequestId != "" {
		t.Errorf("Expected empty requestId for non-JSON response, but got: %v", result.RequestId)
	}
}

func TestListSolutionsInvalidJsonReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("not-json"))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	_, err := client.ListSolutions()

	if err == nil || !strings.Contains(err.Error(), "invalid response body") {
		t.Errorf("Expected invalid response body error, but got: %v", err)
	}
}

func TestListSolutionsBadRequestReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	_, err := client.ListSolutions()

	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error with status code 400, but got: %v", err)
	}
}

func TestPullSolutionServerErrorReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"not found"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	_, err := client.PullSolution("missing-id")

	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("Expected error with status code 404, but got: %v", err)
	}
}

func TestPushSolutionServerErrorReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	file := stream.NewMemoryStream("test.uis", []byte("content"))
	_, err := client.PushSolution(file, "", nil)

	if err == nil || !strings.Contains(err.Error(), "400") {
		t.Errorf("Expected error with status code 400, but got: %v", err)
	}
}

func TestPublishSolutionServerErrorReturnsError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	defer srv.Close()

	client := newTestClient(t, srv.URL, nil)
	_, err := client.PublishSolution("test-id")

	if err == nil || !strings.Contains(err.Error(), "403") {
		t.Errorf("Expected error with status code 403, but got: %v", err)
	}
}

func TestHttpClientSettings(t *testing.T) {
	baseUri, _ := url.Parse("http://localhost")
	logger := log.NewDefaultLogger(io.Discard)
	settings := plugin.ExecutionSettings{
		OperationId: "test-op",
		Insecure:    true,
	}
	client := NewStudioClient(*baseUri, "my-org", nil, true, settings, logger)

	result := client.httpClientSettings()

	if !result.Debug {
		t.Errorf("Expected debug to be true")
	}
}

func TestWriteMultipartForm(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)
	file := stream.NewMemoryStream("test.uis", []byte("file-content"))

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	err := client.writeMultipartForm(writer, file, "application/octet-stream")
	_ = writer.Close()

	if err != nil {
		t.Errorf("Expected no error, but got: %v", err)
	}
	if !strings.Contains(buf.String(), "file-content") {
		t.Errorf("Expected multipart body to contain file content")
	}
	if !strings.Contains(buf.String(), "test.uis") {
		t.Errorf("Expected multipart body to contain filename")
	}
}

func TestWriteMultipartFormWithFailingStream(t *testing.T) {
	client := newTestClient(t, "http://localhost", nil)
	file := &failingStream{name: "test.uis"}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	err := client.writeMultipartForm(writer, file, "application/octet-stream")

	if err == nil {
		t.Errorf("Expected error from failing stream, but got nil")
	}
}

type failingStream struct {
	name string
}

func (s *failingStream) Name() string           { return s.name }
func (s *failingStream) Size() (int64, error)    { return 0, errors.New("size error") }
func (s *failingStream) Data() (io.ReadCloser, error) {
	return nil, errors.New("data error")
}

func newTestProgressBar(logger log.Logger) *visualization.ProgressBar {
	return visualization.NewProgressBar(logger)
}
