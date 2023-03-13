package utils

import (
	"io"
	"strings"
	"testing"
)

func TestReaderStreamName(t *testing.T) {
	param := NewReaderStream("my-file.txt", io.NopCloser(strings.NewReader("hello-world")))

	name := param.Name()

	if name != "my-file.txt" {
		t.Errorf("Did not return provided name, but got: %v", name)
	}
}

func TestReaderStreamData(t *testing.T) {
	param := NewReaderStream("my-file.txt", io.NopCloser(strings.NewReader("hello-world")))

	reader, size, err := param.Data()
	data, _ := io.ReadAll(reader)

	if string(data) != "hello-world" {
		t.Errorf("Did not return provided data, but got: %v", string(data))
	}
	if size != -1 {
		t.Errorf("Did not return correct file size, but got: %v", size)
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}
