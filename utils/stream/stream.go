// Package stream provides methods to accessed stream data while abstracting
// where the data comes from. This allows clients to read data from streams
// without the need to know where the data is streamed from.
package stream

import (
	"io"
)

// A Stream provides access to data and abstracts where the data resides.
//
// It enables streaming large files when sending the HTTP request instead of
// loading them in memory first.
type Stream interface {
	Name() string
	Size() (int64, error)
	Data() (io.ReadCloser, error)
}
