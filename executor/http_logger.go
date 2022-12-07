package executor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type HttpLogger struct {
	Output *bytes.Buffer
}

func (l HttpLogger) logHeaders(header http.Header) {
	for key, values := range header {
		for _, value := range values {
			fmt.Fprintf(l.Output, "%s: %s\n", key, value)
		}
	}
	fmt.Fprint(l.Output, "\n")
}

func (l HttpLogger) logBody(body io.Reader) (int, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return 0, fmt.Errorf("Error reading body: %v", err)
	}

	err = json.Indent(l.Output, data, "", "  ")
	if err != nil {
		fmt.Fprint(l.Output, string(data))
	}
	return len(data), nil
}

func (l HttpLogger) LogRequest(request *http.Request, body io.Reader, debug bool) error {
	if debug {
		fmt.Fprintf(l.Output, "%s %s %s\n", request.Method, request.URL, request.Proto)
		l.logHeaders(request.Header)
		len, err := l.logBody(body)
		if len > 0 {
			fmt.Fprint(l.Output, "\n\n")
		}
		fmt.Fprint(l.Output, "\n")
		return err
	}
	return nil
}

func (l HttpLogger) LogResponse(response *http.Response, debug bool) error {
	if debug {
		fmt.Fprintf(l.Output, "%s %s\n", response.Proto, response.Status)
		l.logHeaders(response.Header)
	}

	len, err := l.logBody(response.Body)
	if len == 0 && response.StatusCode >= 400 {
		fmt.Fprintf(l.Output, "%s %s\n", response.Proto, response.Status)
	}
	return err
}
