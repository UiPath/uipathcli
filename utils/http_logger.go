package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
)

type HttpLogger struct {
	Output       io.Writer
	Debug        bool
	requestCount int
}

func (l HttpLogger) logHeaders(header http.Header) {
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

func (l HttpLogger) LogBody(body io.Reader) ([]byte, error) {
	data, err := io.ReadAll(body)
	if err != nil {
		return []byte{}, fmt.Errorf("Error reading body: %v", err)
	}

	dst := bytes.Buffer{}
	err = json.Indent(&dst, data, "", "  ")
	l.Output.Write(dst.Bytes())
	if err != nil {
		fmt.Fprint(l.Output, string(data))
	}
	return data, nil
}

func (l *HttpLogger) LogRequest(request *http.Request) error {
	l.requestCount = l.requestCount + 1
	if l.requestCount > 1 {
		fmt.Fprint(l.Output, "\n\n")
	}

	if l.Debug {
		fmt.Fprintf(l.Output, "%s %s %s\n", request.Method, request.URL, request.Proto)
		l.logHeaders(request.Header)
		body, err := l.LogBody(request.Body)
		request.Body = io.NopCloser(bytes.NewBuffer(body))
		if len(body) > 0 {
			fmt.Fprint(l.Output, "\n\n")
		}
		fmt.Fprint(l.Output, "\n")
		return err
	}
	return nil
}

func (l HttpLogger) LogResponse(response *http.Response) error {
	if l.Debug {
		fmt.Fprintf(l.Output, "%s %s\n", response.Proto, response.Status)
		l.logHeaders(response.Header)
	}

	body, err := l.LogBody(response.Body)
	response.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) == 0 && response.StatusCode >= 400 {
		fmt.Fprintf(l.Output, "%s %s\n", response.Proto, response.Status)
	} else {
		fmt.Fprint(l.Output, "\n")
	}
	return err
}
