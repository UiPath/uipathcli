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

type nupkgReader struct {
	Path string
}

func (r nupkgReader) ReadNuspec() (*nuspec, error) {
	zip, err := zip.OpenReader(r.Path)
	if err != nil {
		return nil, fmt.Errorf("Could not read package '%s': %v", r.Path, err)
	}
	defer zip.Close()
	for _, file := range zip.File {
		if strings.HasSuffix(file.Name, ".nuspec") {
			return r.readNuspec(r.Path, file)
		}
	}
	return nil, fmt.Errorf("Could not find nuspec in package '%s'", r.Path)
}

func (r nupkgReader) readNuspec(source string, file *zip.File) (*nuspec, error) {
	reader, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("Could not read nuspec in package '%s': %v", source, err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("Could not read nuspec in package '%s': %v", source, err)
	}
	var nuspec nuspecXml
	err = xml.Unmarshal(data, &nuspec)
	if err != nil {
		return nil, fmt.Errorf("Could not read nuspec in package '%s': %v", source, err)
	}
	return newNuspec(nuspec.Metadata.Id, nuspec.Metadata.Title, nuspec.Metadata.Version), nil
}

func findLatestNupkg(directory string) string {
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

func newNupkgReader(path string) *nupkgReader {
	return &nupkgReader{path}
}
