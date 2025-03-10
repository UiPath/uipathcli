package network

import (
	"bytes"
	"io"
	"net/http"
)

type HttpRequest struct {
	Proto         string // e.g. "HTTP/1.0"
	Method        string
	URL           string
	Header        http.Header
	Body          io.Reader
	ContentLength int64
}

func NewHttpGetRequest(url string, header http.Header) *HttpRequest {
	return NewHttpRequest(http.MethodGet, url, header, &bytes.Buffer{}, -1)
}

func NewHttpPostRequest(url string, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return NewHttpRequest(http.MethodPost, url, header, body, contentLength)
}

func NewHttpPutRequest(url string, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return NewHttpRequest(http.MethodPut, url, header, body, contentLength)
}

func NewHttpRequest(method string, url string, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return &HttpRequest{"HTTP/1.1", method, url, header, body, contentLength}
}
