package studio

import (
	"sync"

	"github.com/UiPath/uipathcli/log"
)

var mltiLoggerMutex sync.Mutex

type MultiLogger struct {
	logger log.Logger
	prefix string
}

func (l MultiLogger) LogRequest(request log.RequestInfo) {
	mltiLoggerMutex.Lock()
	defer mltiLoggerMutex.Unlock()

	l.logger.Log(l.prefix)
	l.logger.LogRequest(request)
}

func (l MultiLogger) LogResponse(response log.ResponseInfo) {
	mltiLoggerMutex.Lock()
	defer mltiLoggerMutex.Unlock()

	l.logger.Log(l.prefix)
	l.logger.LogResponse(response)
}

func (l MultiLogger) Log(message string) {
	mltiLoggerMutex.Lock()
	defer mltiLoggerMutex.Unlock()

	l.logger.Log(l.prefix + message)
}

func (l MultiLogger) LogError(message string) {
	mltiLoggerMutex.Lock()
	defer mltiLoggerMutex.Unlock()

	l.logger.LogError(l.prefix + message)
}

func NewMultiLogger(logger log.Logger, prefix string) *MultiLogger {
	return &MultiLogger{logger, prefix}
}
