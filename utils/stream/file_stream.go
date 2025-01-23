package stream

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// The FileStream implements the stream interface for files on disk.
//
// It provides the stream length by reading the file stats.
// The file name is used for the stream name.
type FileStream struct {
	name string
	path string
}

func (s FileStream) Name() string {
	return s.name
}

func (s FileStream) Size() (int64, error) {
	fileStat, err := os.Stat(s.path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return -1, fmt.Errorf("File '%s' not found", s.path)
	}
	if err != nil {
		return -1, fmt.Errorf("Error reading file size '%s': %w", s.path, err)
	}
	return fileStat.Size(), nil
}

func (s FileStream) Data() (io.ReadCloser, error) {
	file, err := os.Open(s.path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("File '%s' not found", s.path)
	}
	if err != nil {
		return nil, fmt.Errorf("Error reading file '%s': %w", s.path, err)
	}
	return file, nil
}

func NewFileStream(path string) *FileStream {
	return &FileStream{
		name: filepath.Base(path),
		path: path,
	}
}
