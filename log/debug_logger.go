package log

import (
	"fmt"
	"io"
	"net/http"
	"sort"
)

// The DebugLogger provides more insights into which operations the CLI is performing.
//
// It can be enabled using the --debug flag.
type DebugLogger struct {
	Output      io.Writer
	ErrorOutput io.Writer
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
	n, _ := io.Copy(l.Output, request.Body)
	if n > 0 {
		fmt.Fprint(l.Output, "\n\n")
	}
	fmt.Fprint(l.Output, "\n")
}

func (l DebugLogger) LogResponse(response ResponseInfo) {
	fmt.Fprintf(l.Output, "%s %s\n", response.Protocol, response.Status)
	l.writeHeaders(response.Header)
	io.Copy(l.Output, response.Body)
	fmt.Fprint(l.Output, "\n\n\n")
}

func (l DebugLogger) LogDebug(message string) {
	fmt.Fprint(l.Output, message)
}

func (l DebugLogger) LogError(message string) {
	fmt.Fprint(l.ErrorOutput, message)
}
