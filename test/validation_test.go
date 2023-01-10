package test

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingRequiredParameterShowsError(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: string
              required:
              - myparameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate"}, context)

	expected := "Argument --myparameter is missing"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain missing parameter error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestEmptyRequiredParameterShowsError(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: string
              required:
              - myparameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", ""}, context)

	expected := "Argument --myparameter is missing"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain missing parameter error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestInvalidDataTypesShowError(t *testing.T) {
	t.Run("Integer", func(t *testing.T) { InvalidDataTypeShowsError(t, "integer") })
	t.Run("Number", func(t *testing.T) { InvalidDataTypeShowsError(t, "number") })
	t.Run("Boolean", func(t *testing.T) { InvalidDataTypeShowsError(t, "boolean") })
}

func InvalidDataTypeShowsError(t *testing.T, datatype string) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: ` + datatype + `
              nullable: false
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", "INVALID"}, context)

	expected := "Cannot convert 'myparameter' value 'INVALID' to " + datatype
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain missing parameter error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestInvalidArrayDataTypesShowError(t *testing.T) {
	t.Run("Integer", func(t *testing.T) { InvalidArrayDataTypeShowsError(t, "integer", "1,INVALID,3") })
	t.Run("Number", func(t *testing.T) { InvalidArrayDataTypeShowsError(t, "number", "1.3, 1.0, INVALID") })
	t.Run("Boolean", func(t *testing.T) { InvalidArrayDataTypeShowsError(t, "boolean", "true, false, invalid") })
}

func InvalidArrayDataTypeShowsError(t *testing.T, datatype string, values string) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: array
                  items:
                    type: ` + datatype + `
              nullable: false
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", values}, context)

	expected := "Cannot convert 'myparameter' values '" + values + "' to " + datatype + " array"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain missing parameter error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestAssignNestedObjectKeyShowsError(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: object
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", "hello.a=x;hello=5"}, context)

	expected := "Cannot convert 'myparameter' value because object key 'hello' is already defined"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain object key already defined error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestSameObjectKeysShowsError(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: object
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", "hello.a=x;hello.a=5"}, context)

	expected := "Cannot convert 'myparameter' value because object key 'a' is already defined"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain object key already defined error, expected: %v, got: %v", expected, result.StdErr)
	}
}

func TestInvalidFileReference(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          multipart/form-data:
            schema:
              properties:
                file:
                  type: string
                  format: binary
                  description: The file to upload
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()
	path := filepath.Join(t.TempDir(), "not-found.txt")

	result := runCli([]string{"myservice", "post-validate", "--file", "file://" + path}, context)

	expected := "File '" + path + "' not found"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr does not contain file not found error, expected: %v, got: %v", expected, result.StdErr)
	}
}
