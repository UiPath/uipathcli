package log

type DefaultLogger struct {
}

func (l *DefaultLogger) LogRequest(request RequestInfo) {
}

func (l DefaultLogger) LogResponse(response ResponseInfo) {
}

func (l DefaultLogger) Log(message string) {
}
