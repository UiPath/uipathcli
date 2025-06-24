// Package network is an abstraction over the net/http client
// which adds resiliency through retries and makes sure every
// request contains the common headers
package network

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"runtime"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils"
	"github.com/UiPath/uipathcli/utils/resiliency"
)

type HttpClient struct {
	logger   log.Logger
	settings HttpClientSettings
}

const bufferLimit = 10 * 1024 * 1024
const loggingLimit = 1 * 1024 * 1024

var UserAgent = fmt.Sprintf("uipathcli/%s (%s; %s)", utils.Version, runtime.GOOS, runtime.GOARCH)

func (c HttpClient) Send(request *HttpRequest) (*HttpResponse, error) {
	return c.sendWithRetries(request, context.Background())
}

func (c HttpClient) SendWithContext(request *HttpRequest, ctx context.Context) (*HttpResponse, error) {
	return c.sendWithRetries(request, ctx)
}

func (c HttpClient) sendWithRetries(request *HttpRequest, ctx context.Context) (*HttpResponse, error) {
	request.Header.Set("User-Agent", UserAgent)
	request.Header.Set("x-request-id", c.settings.OperationId)
	if request.Authorization != nil {
		request.Header.Set("Authorization", fmt.Sprintf("%s %s", request.Authorization.Type, request.Authorization.Value))
	}

	if c.settings.Debug {
		request.Body = newResettableReader(request.Body, bufferLimit, func(body []byte) { c.logRequest(request, body) })
	} else if c.settings.MaxAttempts > 1 {
		request.Body = newResettableReader(request.Body, bufferLimit, func(body []byte) {})
	}

	var response *HttpResponse
	var err error
	err = resiliency.RetryN(c.settings.MaxAttempts, func(attempt int) error {
		if attempt > 1 && !c.resetReader(request.Body) {
			return err
		}

		response, err = c.send(request, ctx)
		if err != nil {
			return resiliency.Retryable(err)
		}

		if c.settings.Debug {
			response.Body = newResettableReader(response.Body, bufferLimit, func(body []byte) { c.logResponse(response, body) })
		}

		if response.StatusCode == 0 || response.StatusCode >= 500 {
			defer func() { _ = response.Body.Close() }()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				return resiliency.Retryable(fmt.Errorf("Error reading response: %w", err))
			}
			return resiliency.Retryable(fmt.Errorf("Service returned status code '%v' and body '%v'", response.StatusCode, string(body)))
		}
		return nil
	})
	return response, err
}

func (c HttpClient) resetReader(reader io.Reader) bool {
	resettableReader, ok := reader.(*resettableReader)
	if ok {
		return resettableReader.Reset()
	}
	return false
}

func (c HttpClient) send(request *HttpRequest, ctx context.Context) (*HttpResponse, error) {
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: c.settings.Insecure}, //nolint:gosec // This is user configurable and disabled by default
		ResponseHeaderTimeout: c.settings.Timeout,
	}
	client := &http.Client{Transport: transport}

	responseChan := make(chan *HttpResponse)
	ctx, cancel := context.WithCancelCause(ctx)
	go func(client *http.Client, request *HttpRequest) {
		req, err := http.NewRequestWithContext(ctx, request.Method, request.URL, request.Body)
		if err != nil {
			cancel(fmt.Errorf("Error preparing request: %w", err))
			return
		}
		req.Header = request.Header
		for k, v := range c.settings.Header {
			req.Header.Set(k, v)
		}
		req.ContentLength = request.ContentLength

		resp, err := client.Do(req) //nolint:bodyclose // The response body needs to be closed by the caller to support streaming
		if err != nil {
			cancel(fmt.Errorf("Error sending request: %w", err))
			return
		}

		response := NewHttpResponse(
			resp.Status,
			resp.StatusCode,
			resp.Proto,
			resp.Header,
			resp.Body,
			resp.ContentLength)
		responseChan <- response
	}(client, request)

	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("Error sending request: %w", context.Cause(ctx))
	case response := <-responseChan:
		return response, nil
	}
}

func (c HttpClient) logRequest(request *HttpRequest, body []byte) {
	reader := bytes.NewReader(c.truncate(body, loggingLimit))
	requestInfo := log.NewRequestInfo(request.Method, request.URL, request.Proto, request.Header, reader)
	c.logger.LogRequest(*requestInfo)
}

func (c HttpClient) logResponse(response *HttpResponse, body []byte) {
	reader := bytes.NewReader(c.truncate(body, loggingLimit))
	responseInfo := log.NewResponseInfo(response.StatusCode, response.Status, response.Proto, response.Header, reader)
	c.logger.LogResponse(*responseInfo)
}

func (c HttpClient) truncate(data []byte, size int) []byte {
	if len(data) > size {
		return data[:size]
	}
	return data
}

func NewHttpClient(logger log.Logger, settings HttpClientSettings) *HttpClient {
	return &HttpClient{logger, settings}
}
