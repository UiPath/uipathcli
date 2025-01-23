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
