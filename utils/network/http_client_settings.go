package network

import "time"

type HttpClientSettings struct {
	Debug       bool
	OperationId string
	Timeout     time.Duration
	MaxAttempts int
	Insecure    bool
}

func NewHttpClientSettings(
	debug bool,
	operationId string,
	timeout time.Duration,
	maxAttempts int,
	insecure bool) *HttpClientSettings {
	return &HttpClientSettings{
		debug,
		operationId,
		timeout,
		maxAttempts,
		insecure,
	}
}
