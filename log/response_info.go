package log

import "io"

// ResponseInfo contains the information used by the logger to print the executor result.
type ResponseInfo struct {
	StatusCode int
	Status     string
	Protocol   string
	Header     map[string][]string
	Body       io.Reader
}

func NewResponseInfo(statusCode int, status string, protocol string, header map[string][]string, body io.Reader) *ResponseInfo {
	return &ResponseInfo{statusCode, status, protocol, header, body}
}
