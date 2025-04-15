package network

import (
	"time"
)

type HttpClientSettings struct {
	Debug       bool
	OperationId string
	Header      map[string]string
	Timeout     time.Duration
	MaxAttempts int
	Insecure    bool
}

func NewHttpClientSettings(
	debug bool,
	operationId string,
	header map[string]string,
	timeout time.Duration,
	maxAttempts int,
	insecure bool) *HttpClientSettings {
	return &HttpClientSettings{
		debug,
		operationId,
		header,
		timeout,
		maxAttempts,
		insecure,
	}
}
