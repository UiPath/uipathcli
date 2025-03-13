package network

import (
	"bytes"
	"io"
)

type resettableReader struct {
	reader           io.Reader
	buffer           *bytes.Buffer
	bufferLimit      int64
	bytesRead        int64
	onFinishCallback func([]byte)
}

func (r *resettableReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	r.bytesRead = r.bytesRead + int64(n)

	if n > 0 && !r.bufferExceeded() {
		_, _ = r.buffer.Write(p[:n])
	}
	if err == io.EOF {
		r.onFinishCallback(r.buffer.Bytes())
	}
	return n, err
}

func (r *resettableReader) Reset() bool {
	if r.bufferExceeded() {
		return false
	}

	r.bytesRead = 0
	data := r.buffer.Bytes()
	r.reader = io.NopCloser(bytes.NewReader(data))
	r.buffer = new(bytes.Buffer)
	return true
}

func (r *resettableReader) bufferExceeded() bool {
	return r.bytesRead > r.bufferLimit
}

func (r *resettableReader) Close() error {
	closer, ok := r.reader.(io.Closer)
	if ok {
		return closer.Close()
	}
	return nil
}

func newResettableReader(reader io.Reader, bufferLimit int64, onFinishCallback func([]byte)) *resettableReader {
	return &resettableReader{
		reader:           reader,
		buffer:           new(bytes.Buffer),
		bufferLimit:      bufferLimit,
		bytesRead:        0,
		onFinishCallback: onFinishCallback,
	}
}
