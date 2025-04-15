package studio

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const nuspecTemplate = `
<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2013/05/nuspec.xsd">
  <metadata minClientVersion="3.3">
    <id>%s</id>
    <title>%s</title>
    <version>%s</version>
  </metadata>
</package>`

type NupkgWriter struct {
	Path          string
	nuspecName    string
	nuspecContent string
	fileName      string
	fileContent   []byte
}

func (w *NupkgWriter) WithNuspec(nuspec Nuspec) *NupkgWriter {
	w.nuspecName = nuspec.Id + ".nuspec"
	w.nuspecContent = fmt.Sprintf(nuspecTemplate, nuspec.Id, nuspec.Title, nuspec.Version)
	return w
}

func (w *NupkgWriter) WithFile(name string, content []byte) *NupkgWriter {
	w.fileName = name
	w.fileContent = content
	return w
}

func (w NupkgWriter) writeNuspec(zipWriter *zip.Writer, name string, content string) error {
	nuspecWriter, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = io.WriteString(nuspecWriter, content)
	if err != nil {
		return err
	}
	return nil
}

func (w NupkgWriter) writeFile(zipWriter *zip.Writer, name string, content []byte) error {
	writer, err := zipWriter.CreateHeader(&zip.FileHeader{
		Name:   name,
		Method: zip.Store,
	})
	if err != nil {
		return err
	}
	_, err = writer.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func (w NupkgWriter) Write() error {
	_ = os.MkdirAll(filepath.Dir(w.Path), 0700)
	archive, err := os.OpenFile(w.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Could not create nupkg file '%s': %v", w.Path, err)
	}
	defer archive.Close()
	zipWriter := zip.NewWriter(archive)

	if w.nuspecName != "" {
		err = w.writeNuspec(zipWriter, w.nuspecName, w.nuspecContent)
		if err != nil {
			return fmt.Errorf("Could not write nuspec file content '%s': %v", w.nuspecName, err)
		}
	}

	if w.fileName != "" {
		err = w.writeFile(zipWriter, w.fileName, w.fileContent)
		if err != nil {
			return fmt.Errorf("Could not write file '%s': %v", w.fileName, err)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return fmt.Errorf("Could not close nupkg file '%s': %v", w.Path, err)
	}
	return nil
}

func NewNupkgWriter(path string) *NupkgWriter {
	return &NupkgWriter{Path: path}
}
