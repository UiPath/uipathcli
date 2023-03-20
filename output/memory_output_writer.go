package output

import (
	"bytes"
	"io"
)

// The MemoryOutputWriter keeps the response in memory so it can read
// multiple times.
type MemoryOutputWriter struct {
	statusCode int
	status     string
	protocol   string
	header     map[string][]string
	body       []byte
}

func (w *MemoryOutputWriter) WriteResponse(response ResponseInfo) error {
	w.statusCode = response.StatusCode
	w.status = response.Status
	w.protocol = response.Protocol
	w.header = response.Header
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	w.body = body
	return nil
}

func (w MemoryOutputWriter) Response() ResponseInfo {
	return *NewResponseInfo(w.statusCode, w.status, w.protocol, w.header, bytes.NewReader(w.body))
}

func NewMemoryOutputWriter() *MemoryOutputWriter {
	return &MemoryOutputWriter{}
}
