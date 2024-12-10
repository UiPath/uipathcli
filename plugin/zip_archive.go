package plugin

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const MaxArchiveSize = 1 * 1024 * 1024 * 1024

type zipArchive struct{}

func (z zipArchive) Extract(zipArchive string, destinationFolder string, permissions os.FileMode) error {
	archive, err := zip.OpenReader(zipArchive)
	if err != nil {
		return err
	}
	defer archive.Close()

	for _, file := range archive.File {
		err := z.extractFile(file, destinationFolder, permissions)
		if err != nil {
			return err
		}
	}
	return nil
}

func (z zipArchive) extractFile(zipFile *zip.File, destinationFolder string, permissions os.FileMode) error {
	path, err := z.sanitizeArchivePath(destinationFolder, zipFile.Name)
	if err != nil {
		return err
	}

	if zipFile.FileInfo().IsDir() {
		return os.MkdirAll(path, permissions)
	}
	err = os.MkdirAll(filepath.Dir(path), permissions)
	if err != nil {
		return err
	}

	zipFileReader, err := zipFile.Open()
	if err != nil {
		return err
	}
	defer zipFileReader.Close()

	destinationFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, zipFile.Mode())
	if err != nil {
		return err
	}
	defer destinationFile.Close()

	_, err = io.CopyN(destinationFile, zipFileReader, MaxArchiveSize)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (z zipArchive) sanitizeArchivePath(directory string, name string) (string, error) {
	result := filepath.Join(directory, name)
	if strings.HasPrefix(result, filepath.Clean(directory)) {
		return result, nil
	}
	return "", fmt.Errorf("File path '%s' is not allowed", directory)
}

func newZipArchive() *zipArchive {
	return &zipArchive{}
}
