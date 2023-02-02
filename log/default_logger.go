package log

import (
	"fmt"
	"io"
)

type DefaultLogger struct {
	ErrorOutput io.Writer
}

func (l *DefaultLogger) LogRequest(request RequestInfo) {
}

func (l DefaultLogger) LogResponse(response ResponseInfo) {
}

func (l DefaultLogger) Log(message string) {
}

func (l DefaultLogger) LogError(message string) {
	fmt.Fprint(l.ErrorOutput, message)
}
