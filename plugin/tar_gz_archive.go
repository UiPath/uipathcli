package plugin

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type tarGzArchive struct{}

func (t tarGzArchive) Extract(filePath string, destinationFolder string, permissions os.FileMode) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()

	gzipReader, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer func() { _ = gzipReader.Close() }()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}

		err = t.extractFile(header, tarReader, destinationFolder, permissions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t tarGzArchive) extractFile(header *tar.Header, reader *tar.Reader, destinationFolder string, permissions os.FileMode) error {
	path, err := t.sanitizeArchivePath(destinationFolder, header.Name)
	if err != nil {
		return err
	}

	if header.FileInfo().IsDir() {
		return os.MkdirAll(path, permissions)
	}

	err = os.MkdirAll(filepath.Dir(path), permissions)
	if err != nil {
		return err
	}

	destinationFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, permissions)
	if err != nil {
		return err
	}
	defer func() { _ = destinationFile.Close() }()

	_, err = io.CopyN(destinationFile, reader, MaxArchiveSize)
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	return nil
}

func (t tarGzArchive) sanitizeArchivePath(directory string, name string) (string, error) {
	result := filepath.Join(directory, name)
	if strings.HasPrefix(result, filepath.Clean(directory)) {
		return result, nil
	}
	return "", fmt.Errorf("File path '%s' is not allowed", directory)
}

func newTarGzArchive() *tarGzArchive {
	return &tarGzArchive{}
}
