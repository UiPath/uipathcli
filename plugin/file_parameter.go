package plugin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

type FileParameter struct {
	path     string
	filename string
	data     []byte
}

func (f FileParameter) Filename() string {
	return f.filename
}

func (f FileParameter) Data() (io.ReadCloser, int64, error) {
	if f.data != nil {
		return io.NopCloser(bytes.NewReader(f.data)), int64(len(f.data)), nil
	}
	file, err := os.Open(f.path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, -1, fmt.Errorf("File '%s' not found", f.path)
	}
	if err != nil {
		return nil, -1, fmt.Errorf("Error reading file '%s': %v", f.path, err)
	}
	fileStat, err := file.Stat()
	if err != nil {
		return nil, -1, fmt.Errorf("Error reading file size '%s': %v", f.path, err)
	}
	return file, fileStat.Size(), nil
}

func NewFileParameter(path string, filename string, data []byte) *FileParameter {
	return &FileParameter{
		path:     path,
		filename: filename,
		data:     data,
	}
}
