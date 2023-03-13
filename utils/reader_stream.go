package utils

import (
	"io"
)

// The ReaderStream implements the stream interface by wrapping the
// generic io.ReadCloser interface.
//
// It does not provide any stream length information due to performance
// reasons.
// The reader does not have a name, so it needs to be provided when
// initializing a new ReaderStream instance.
type ReaderStream struct {
	name   string
	reader io.ReadCloser
}

func (s ReaderStream) Name() string {
	return s.name
}

func (s ReaderStream) Data() (io.ReadCloser, int64, error) {
	return s.reader, -1, nil
}

func NewReaderStream(name string, reader io.ReadCloser) *ReaderStream {
	return &ReaderStream{
		name:   name,
		reader: reader,
	}
}
