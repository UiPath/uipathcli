package studio

import (
	"archive/zip"
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
)

type NupkgWriter struct {
	Path        string
	nuspec      *Nuspec
	fileName    string
	fileContent []byte
}

func (w *NupkgWriter) WithNuspec(nuspec Nuspec) *NupkgWriter {
	w.nuspec = &nuspec
	return w
}

func (w *NupkgWriter) WithFile(name string, content []byte) *NupkgWriter {
	w.fileName = name
	w.fileContent = content
	return w
}

func (w *NupkgWriter) writeNuspec(zipWriter *zip.Writer, nuspec Nuspec) error {
	name := nuspec.Id + ".nuspec"
	nuspecXml := nuspecXml{Metadata: nuspecPackageMetadataXml(nuspec)}
	content, err := xml.MarshalIndent(nuspecXml, "", "  ")
	if err != nil {
		return err
	}
	nuspecWriter, err := zipWriter.Create(name)
	if err != nil {
		return err
	}
	_, err = nuspecWriter.Write(content)
	if err != nil {
		return err
	}
	return nil
}

func (w *NupkgWriter) writeFile(zipWriter *zip.Writer, name string, content []byte) error {
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

func (w *NupkgWriter) Write() error {
	_ = os.MkdirAll(filepath.Dir(w.Path), 0700)
	archive, err := os.OpenFile(w.Path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("Could not create nupkg file '%s': %w", w.Path, err)
	}
	defer func() { _ = archive.Close() }()
	zipWriter := zip.NewWriter(archive)

	if w.nuspec != nil {
		err = w.writeNuspec(zipWriter, *w.nuspec)
		if err != nil {
			return fmt.Errorf("Could not write nuspec file content '%s': %w", w.nuspec.Id, err)
		}
	}

	if w.fileName != "" {
		err = w.writeFile(zipWriter, w.fileName, w.fileContent)
		if err != nil {
			return fmt.Errorf("Could not write file '%s': %w", w.fileName, err)
		}
	}

	err = zipWriter.Close()
	if err != nil {
		return fmt.Errorf("Could not close nupkg file '%s': %w", w.Path, err)
	}
	return nil
}

func NewNupkgWriter(path string) *NupkgWriter {
	return &NupkgWriter{Path: path}
}
