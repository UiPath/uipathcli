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
	Authorization *Authorization
	Header        http.Header
	Body          io.Reader
	ContentLength int64
}

func NewHttpGetRequest(url string, authorization *Authorization, header http.Header) *HttpRequest {
	return NewHttpRequest(http.MethodGet, url, authorization, header, &bytes.Buffer{}, -1)
}

func NewHttpPostRequest(url string, authorization *Authorization, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return NewHttpRequest(http.MethodPost, url, authorization, header, body, contentLength)
}

func NewHttpPutRequest(url string, authorization *Authorization, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return NewHttpRequest(http.MethodPut, url, authorization, header, body, contentLength)
}

func NewHttpPatchRequest(url string, authorization *Authorization, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return NewHttpRequest(http.MethodPatch, url, authorization, header, body, contentLength)
}

func NewHttpRequest(method string, url string, authorization *Authorization, header http.Header, body io.Reader, contentLength int64) *HttpRequest {
	return &HttpRequest{"HTTP/1.1", method, url, authorization, header, body, contentLength}
}
