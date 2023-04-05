package utils

import (
	"bytes"
	"io"
)

// The MemoryStream implements the stream interface using a simple
// byte array.
//
// The byte array length is used as the stream length.
// The name needs to be provided when initializing a new MemoryStream
// instance.
type MemoryStream struct {
	name string
	data []byte
}

func (s MemoryStream) Name() string {
	return s.name
}

func (s MemoryStream) Size() (int64, error) {
	return int64(len(s.data)), nil
}

func (s MemoryStream) Data() (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(s.data)), nil
}

func NewMemoryStream(name string, data []byte) *MemoryStream {
	return &MemoryStream{
		name: name,
		data: data,
	}
}
