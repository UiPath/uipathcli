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
	output      io.Writer
	errorOutput io.Writer
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
			fmt.Fprintf(l.output, "%s: %s\n", key, value)
		}
	}
	fmt.Fprint(l.output, "\n")
}

func (l *DebugLogger) LogRequest(request RequestInfo) {
	fmt.Fprintf(l.output, "%s %s %s\n", request.Method, request.Url, request.Protocol)
	l.writeHeaders(request.Header)
	n, _ := io.Copy(l.output, request.Body)
	if n > 0 {
		fmt.Fprint(l.output, "\n\n")
	}
	fmt.Fprint(l.output, "\n")
}

func (l DebugLogger) LogResponse(response ResponseInfo) {
	fmt.Fprintf(l.output, "%s %s\n", response.Protocol, response.Status)
	l.writeHeaders(response.Header)
	io.Copy(l.output, response.Body)
	fmt.Fprint(l.output, "\n\n\n")
}

func (l DebugLogger) LogDebug(message string) {
	fmt.Fprint(l.output, message)
}

func (l DebugLogger) LogError(message string) {
	fmt.Fprint(l.errorOutput, message)
}

func NewDebugLogger(output io.Writer, errorOutput io.Writer) *DebugLogger {
	return &DebugLogger{output, errorOutput}
}
