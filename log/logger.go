package log

type Logger interface {
	LogDebug(message string)
	LogError(message string)
	LogRequest(request RequestInfo)
	LogResponse(response ResponseInfo)
}
