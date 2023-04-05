package utils

import (
	"io"
	"testing"
)

func TestMemoryStreamName(t *testing.T) {
	param := NewMemoryStream("my-file.txt", []byte("hello-world"))

	name := param.Name()

	if name != "my-file.txt" {
		t.Errorf("Did not return provided name, but got: %v", name)
	}
}

func TestMemoryStreamSize(t *testing.T) {
	param := NewMemoryStream("my-file.txt", []byte("hello-world"))

	size, err := param.Size()

	if size != int64(len("hello-world")) {
		t.Errorf("Did not return correct file size, but got: %v", size)
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}

func TestMemoryStreamData(t *testing.T) {
	param := NewMemoryStream("my-file.txt", []byte("hello-world"))

	reader, err := param.Data()
	data, _ := io.ReadAll(reader)

	if string(data) != "hello-world" {
		t.Errorf("Did not return provided data, but got: %v", string(data))
	}
	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
}
