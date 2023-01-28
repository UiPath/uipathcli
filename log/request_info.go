package log

type RequestInfo struct {
	Method   string
	Url      string
	Protocol string
	Header   map[string][]string
	Body     []byte
}

func NewRequestInfo(method string, url string, protocol string, header map[string][]string, body []byte) *RequestInfo {
	return &RequestInfo{method, url, protocol, header, body}
}
