package studio

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type NupkgReader struct {
	Path string
}

func (r NupkgReader) ReadNuspec() (*Nuspec, error) {
	zip, err := zip.OpenReader(r.Path)
	if err != nil {
		return nil, fmt.Errorf("Could not read package '%s': %w", r.Path, err)
	}
	defer func() { _ = zip.Close() }()
	for _, file := range zip.File {
		if strings.HasSuffix(file.Name, ".nuspec") {
			return r.readNuspec(r.Path, file)
		}
	}
	return nil, fmt.Errorf("Could not find nuspec in package '%s'", r.Path)
}

func (r NupkgReader) readNuspec(source string, file *zip.File) (*Nuspec, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("Could not read nuspec in package '%s': %w", source, err)
	}
	defer func() { _ = reader.Close() }()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Could not read nuspec in package '%s': %w", source, err)
	}
	var nuspec nuspecXml
	err = xml.Unmarshal(data, &nuspec)
	if err != nil {
		return nil, fmt.Errorf("Could not read nuspec in package '%s': %w", source, err)
	}
	return NewNuspec(nuspec.Metadata.Id, nuspec.Metadata.Title, nuspec.Metadata.Version), nil
}

func FindLatestNupkg(directory string) string {
	newestFile := ""
	newestTime := time.Time{}

	files, _ := os.ReadDir(directory)
	for _, file := range files {
		extension := filepath.Ext(file.Name())
		if strings.EqualFold(extension, ".nupkg") {
			fileInfo, _ := file.Info()
			time := fileInfo.ModTime()
			if time.After(newestTime) {
				newestTime = time
				newestFile = filepath.Join(directory, file.Name())
			}
		}
	}
	return newestFile
}

func NewNupkgReader(path string) *NupkgReader {
	return &NupkgReader{path}
}
