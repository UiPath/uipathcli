package log

import (
	"fmt"
	"io"
	"net/http"
	"sort"
)

type DebugLogger struct {
	Output io.Writer
}

func (l DebugLogger) writeHeaders(header http.Header) {
	keys := []string{}
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		values := header[key]
		for _, value := range values {
			fmt.Fprintf(l.Output, "%s: %s\n", key, value)
		}
	}
	fmt.Fprint(l.Output, "\n")
}

func (l *DebugLogger) LogRequest(request RequestInfo) {
	fmt.Fprintf(l.Output, "%s %s %s\n", request.Method, request.Url, request.Protocol)
	l.writeHeaders(request.Header)
	fmt.Fprint(l.Output, string(request.Body))
	if len(request.Body) > 0 {
		fmt.Fprint(l.Output, "\n\n")
	}
	fmt.Fprint(l.Output, "\n")
}

func (l DebugLogger) LogResponse(response ResponseInfo) {
	fmt.Fprintf(l.Output, "%s %s\n", response.Protocol, response.Status)
	l.writeHeaders(response.Header)
	fmt.Fprint(l.Output, string(response.Body))
	fmt.Fprint(l.Output, "\n\n\n")
}

func (l DebugLogger) Log(message string) {
	fmt.Fprint(l.Output, message)
}
