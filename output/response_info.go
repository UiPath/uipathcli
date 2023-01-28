package output

type ResponseInfo struct {
	StatusCode int
	Status     string
	Protocol   string
	Header     map[string][]string
	Body       []byte
}

func NewResponseInfo(statusCode int, status string, protocol string, header map[string][]string, body []byte) *ResponseInfo {
	return &ResponseInfo{statusCode, status, protocol, header, body}
}
