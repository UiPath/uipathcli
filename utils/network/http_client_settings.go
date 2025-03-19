package network

import "time"

type HttpClientSettings struct {
	OperationId string
	Timeout     time.Duration
	MaxAttempts int
	Insecure    bool
}

func NewHttpClientSettings(
	operationId string,
	timeout time.Duration,
	maxAttempts int,
	insecure bool) *HttpClientSettings {
	return &HttpClientSettings{
		operationId,
		timeout,
		maxAttempts,
		insecure,
	}
}
