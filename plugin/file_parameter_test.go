package plugin

import (
	"io"
	"os"
	"testing"
)

func TestFileParameterFilename(t *testing.T) {
	param := NewFileParameter("test-path", "my-file.txt", nil)

	filename := param.Filename()

	if filename != "my-file.txt" {
		t.Errorf("Did not return provided file name, but got: %v", filename)
	}
}

func TestFileParameterFromData(t *testing.T) {
	param := NewFileParameter("test-path", "my-file.txt", []byte("hello-world"))

	file, size, err := param.Data()
	data, _ := io.ReadAll(file)

	if string(data) != "hello-world" {
		t.Errorf("Did not return provided data, but got: %v", string(data))
	}
	if size != int64(len("hello-world")) {
		t.Errorf("Did not return correct file size, but got: %v", size)
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestFileParameterWithoutData(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "uipath-test")
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	os.WriteFile(tempFile.Name(), []byte("hello-world"), 0600)
	param := NewFileParameter(tempFile.Name(), "my-file.txt", nil)

	file, size, err := param.Data()
	data, _ := io.ReadAll(file)

	if string(data) != "hello-world" {
		t.Errorf("Did not return provided data, but got: %v", string(data))
	}
	if size != int64(len("hello-world")) {
		t.Errorf("Did not return correct file size, but got: %v", size)
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestFileParameterFileNotFound(t *testing.T) {
	param := NewFileParameter("unknown-path/my-file.txt", "my-file.txt", nil)

	_, _, err := param.Data()

	if err.Error() != "File 'unknown-path/my-file.txt' not found" {
		t.Errorf("Should return file not found error, but got: %v", err)
	}
}
