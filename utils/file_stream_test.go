package utils

import (
	"io"
	"os"
	"testing"
)

func TestFileStreamName(t *testing.T) {
	param := NewFileStream("test-path/my-file.txt")

	name := param.Name()

	if name != "my-file.txt" {
		t.Errorf("Did not return provided file name, but got: %v", name)
	}
}

func TestFileStreamSize(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "uipath-test")
	defer tempFile.Close()
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	err := os.WriteFile(tempFile.Name(), []byte("hello-world"), 0600)
	if err != nil {
		t.Fatalf("Error writing file '%s': %v", tempFile.Name(), err)
	}
	param := NewFileStream(tempFile.Name())

	size, err := param.Size()

	if size != int64(len("hello-world")) {
		t.Errorf("Did not return correct file size, but got: %v", size)
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestFileStreamData(t *testing.T) {
	tempFile, _ := os.CreateTemp("", "uipath-test")
	defer tempFile.Close()
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	err := os.WriteFile(tempFile.Name(), []byte("hello-world"), 0600)
	if err != nil {
		t.Fatalf("Error writing file '%s': %v", tempFile.Name(), err)
	}
	param := NewFileStream(tempFile.Name())

	reader, err := param.Data()
	data, _ := io.ReadAll(reader)

	if string(data) != "hello-world" {
		t.Errorf("Did not return provided data, but got: %v", string(data))
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestFileStreamFileNotFound(t *testing.T) {
	param := NewFileStream("unknown-path/my-file.txt")

	_, err := param.Data()

	if err.Error() != "File 'unknown-path/my-file.txt' not found" {
		t.Errorf("Should return file not found error, but got: %v", err)
	}
}
