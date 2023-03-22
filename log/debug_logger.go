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
	writer io.Writer
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
			fmt.Fprintf(l.writer, "%s: %s\n", key, value)
		}
	}
	fmt.Fprint(l.writer, "\n")
}

func (l *DebugLogger) LogRequest(request RequestInfo) {
	fmt.Fprintf(l.writer, "%s %s %s\n", request.Method, request.Url, request.Protocol)
	l.writeHeaders(request.Header)
	n, _ := io.Copy(l.writer, request.Body)
	if n > 0 {
		fmt.Fprint(l.writer, "\n\n")
	}
	fmt.Fprint(l.writer, "\n")
}

func (l DebugLogger) LogResponse(response ResponseInfo) {
	fmt.Fprintf(l.writer, "%s %s\n", response.Protocol, response.Status)
	l.writeHeaders(response.Header)
	_, _ = io.Copy(l.writer, response.Body)
	fmt.Fprint(l.writer, "\n\n\n")
}

func (l DebugLogger) LogError(message string) {
	fmt.Fprint(l.writer, message)
}

func NewDebugLogger(writer io.Writer) *DebugLogger {
	return &DebugLogger{writer}
}
