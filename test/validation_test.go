package commandline

import (
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
