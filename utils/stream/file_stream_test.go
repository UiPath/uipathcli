package stream

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	path := createFile(t, "my-file.txt", "hello-world")
	param := NewFileStream(path)

	size, err := param.Size()

	if size != int64(len("hello-world")) {
		t.Errorf("Did not return correct file size, but got: %v", size)
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestFileStreamData(t *testing.T) {
	path := createFile(t, "my-file.txt", "hello-world")
	stream := NewFileStream(path)

	reader, err := stream.Data()
	data, _ := io.ReadAll(reader)
	defer func() {
		err := reader.Close()
		if err != nil {
			t.Errorf("Should not return error when closing reader, but got: %v", err)
		}
	}()

	if string(data) != "hello-world" {
		t.Errorf("Did not return provided data, but got: %v", string(data))
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestFileStreamFileNotFound(t *testing.T) {
	stream := NewFileStream("unknown-path/my-file.txt")

	_, err := stream.Data()

	if err.Error() != "File 'unknown-path/my-file.txt' not found" {
		t.Errorf("Should return file not found error, but got: %v", err)
	}
}

func createFile(t *testing.T, name string, content string) string {
	path := filepath.Join(t.TempDir(), name)
	err := os.WriteFile(path, []byte(content), 0600)
	if err != nil {
		t.Fatal(fmt.Errorf("Error writing file '%s': %w", path, err))
	}
	return path
}
