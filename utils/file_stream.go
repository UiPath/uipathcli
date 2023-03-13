package utils

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

func (s FileStream) Data() (io.ReadCloser, int64, error) {
	file, err := os.Open(s.path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, -1, fmt.Errorf("File '%s' not found", s.path)
	}
	if err != nil {
		return nil, -1, fmt.Errorf("Error reading file '%s': %v", s.path, err)
	}
	fileStat, err := file.Stat()
	if err != nil {
		return nil, -1, fmt.Errorf("Error reading file size '%s': %v", s.path, err)
	}
	return file, fileStat.Size(), nil
}

func NewFileStream(path string) *FileStream {
	return &FileStream{
		name: filepath.Base(path),
		path: path,
	}
}
