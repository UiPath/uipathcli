package executor

import (
	"time"
)

// The ExecutionSettings provides global settings for executing commands.
type ExecutionSettings struct {
	OperationId string
	Header      map[string]string
	Timeout     time.Duration
	MaxAttempts int
	Insecure    bool
}

func NewExecutionSettings(
	operationId string,
	header map[string]string,
	timeout time.Duration,
	maxAttempts int,
	insecure bool) *ExecutionSettings {
	return &ExecutionSettings{
		operationId,
		header,
		timeout,
		maxAttempts,
		insecure,
	}
}
