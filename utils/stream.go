package utils

import (
	"io"
)

// A Stream provides access to data and abstracts where the data resides.
//
// It enables streaming large files when sending the HTTP request instead of
// loading them in memory first.
type Stream interface {
	Name() string
	Data() (io.ReadCloser, int64, error)
}
