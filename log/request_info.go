package log

import "io"

// RequestInfo contains the information used by the logger to print debug information which includes
// the request information.
type RequestInfo struct {
	Method   string
	Url      string
	Protocol string
	Header   map[string][]string
	Body     io.Reader
}

func NewRequestInfo(method string, url string, protocol string, header map[string][]string, body io.Reader) *RequestInfo {
	return &RequestInfo{method, url, protocol, header, body}
}
