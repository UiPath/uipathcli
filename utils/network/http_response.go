package network

import (
	"io"
	"net/http"
)

type HttpResponse struct {
	Status        string // e.g. "200 OK"
	StatusCode    int    // e.g. 200
	Proto         string // e.g. "HTTP/1.0"
	Header        http.Header
	Body          io.ReadCloser
	ContentLength int64
}

func NewHttpResponse(status string, statusCode int, proto string, header http.Header, body io.ReadCloser, contentLength int64) *HttpResponse {
	return &HttpResponse{status, statusCode, proto, header, body, contentLength}
}
