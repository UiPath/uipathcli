package log

type Logger interface {
	Log(message string)
	LogError(message string)
	LogRequest(request RequestInfo)
	LogResponse(response ResponseInfo)
}
