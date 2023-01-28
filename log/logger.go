package log

type Logger interface {
	Log(message string)
	LogRequest(request RequestInfo)
	LogResponse(response ResponseInfo)
}
