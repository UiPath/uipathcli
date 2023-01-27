package test

import (
	"strings"
	"testing"
)

func TestDefinitionParsedSuccessfully(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice"}, context)

	if result.Error != nil {
		t.Errorf("Unexpected error, got: %v", result.Error)
	}
}

func TestServiceDescription(t *testing.T) {
	definition := `
info:
  title: This is my custom service
`
	context := NewContextBuilder().
		WithDefinition("testservice", definition).
		Build()

	result := runCli([]string{"testservice", "--help"}, context)

	expected := "This is my custom service"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain service description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestMultipleOperationsSortedByName(t *testing.T) {
	definition := `
paths:
  /hello:
    get:
      summary: hello
      operationId: hello-operation
  /aaaaa:
    get:
      summary: aaaaa
      operationId: aaaaa-operation
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "--help"}, context)

	if strings.Index(result.StdOut, "aaaaa-operation") >= strings.Index(result.StdOut, "hello-operation") {
		t.Errorf("Expected aaaaa operation to be first, got: %v", result.StdOut)
	}
}

func TestOperationName(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice"}, context)

	expected := "get-ping"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain operation name, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestCustomOperationName(t *testing.T) {
	t.Run("SimpleOperationId", func(t *testing.T) { CustomOperationName(t, "my-ping-operation", "my-ping-operation") })
	t.Run("OperationIdWithSlash", func(t *testing.T) { CustomOperationName(t, "my/operation", "my-operation") })
	t.Run("OperationIdWithUnderscore", func(t *testing.T) { CustomOperationName(t, "my_operation", "my-operation") })
	t.Run("UppercaseOperationId", func(t *testing.T) { CustomOperationName(t, "MY_Operation", "my-operation") })
	t.Run("CamelCaseOperationId", func(t *testing.T) { CustomOperationName(t, "myOperationName", "my-operation-name") })
	t.Run("AlreadySnakeCasedOperationId", func(t *testing.T) { CustomOperationName(t, "my-Operation-Name", "my-operation-name") })
}

func CustomOperationName(t *testing.T, operationId string, expectedName string) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ` + operationId

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice"}, context)

	if !strings.Contains(result.StdOut, expectedName) {
		t.Errorf("stdout does not contain custom operation, expected: %v, got: %v", expectedName, result.StdOut)
	}
}

func TestOperationDescription(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice"}, context)

	expected := "Simple ping"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain ping operation summary, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestParameterDescription(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter 
        schema:
          type: string
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "ping", "--filter", "my-filter"}, context)

	expected := "/ping?filter=my-filter"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Request url did not contain query string, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestParameterWithoutSchema(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "ping", "--filter", "my-filter"}, context)

	expected := "/ping?filter=my-filter"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Request url did not contain query string, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestParameterDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { ParameterDataType(t, "string") })
	t.Run("Integer", func(t *testing.T) { ParameterDataType(t, "integer") })
	t.Run("Number", func(t *testing.T) { ParameterDataType(t, "number") })
	t.Run("Boolean", func(t *testing.T) { ParameterDataType(t, "boolean") })
}

func ParameterDataType(t *testing.T, datatype string) {
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
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myparameter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain request body parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestHelpShowsParameterIsRequired(t *testing.T) {
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
                  description: This is my parameter
              required:
                - myparameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "This is my parameter (required)"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain request body parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestHelpShowsParameterWithDefaultValue(t *testing.T) {
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
                  description: This is my parameter
                  default: '1'
              required:
                - myparameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "This is my parameter (default: 1)"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain request body parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestParameterWithoutType(t *testing.T) {
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
                  description: This is my parameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myparameter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain request body parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestBodyParameterDescription(t *testing.T) {
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
                  description: This is my parameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "This is my parameter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain parameter description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestBodyParameterSchemaRef(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ValidationRequest'
components:
  schemas:
    ValidationRequest:
      type: object
      properties:
        myname:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myname"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain body parameter from schema reference, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestBodyParameterSchemaRefWithoutType(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ValidationRequest'
components:
  schemas:
    ValidationRequest:
      type: object
      properties:
        myname:
          description: this is my parameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myname"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain body parameter from schema reference, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestCustomParameterName(t *testing.T) {
	t.Run("SimpleParameter", func(t *testing.T) { CustomParameterName(t, "myparam", "--myparam") })
	t.Run("ParameterWithDollarSign", func(t *testing.T) { CustomParameterName(t, "$myparam", "--myparam") })
	t.Run("UppercaseParameter", func(t *testing.T) { CustomParameterName(t, "MY-PARAMETER", "--my-parameter") })
	t.Run("CamelCaseParameter", func(t *testing.T) { CustomParameterName(t, "myParameterName", "--my-parameter-name") })
	t.Run("AlreadySnakeCasedParameter", func(t *testing.T) { CustomParameterName(t, "my-Parameter-Name", "--my-parameter-name") })
}

func CustomParameterName(t *testing.T, name string, expectedParameter string) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                ` + name + `:
                  type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	if !strings.Contains(result.StdOut, expectedParameter) {
		t.Errorf("stdout does not contain properly cased parameter, expected: %v, got: %v", expectedParameter, result.StdOut)
	}
}

func TestRemoveDollarSignInParameter(t *testing.T) {
	definition := `
paths:
  /data:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                $top:
                  type: integer
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-data", "--help"}, context)

	expected := "--top"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not remove dollar sign from parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestBodyParameterArray(t *testing.T) {
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
                  description: This is my parameter
                  type: array
                  items:
                    type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "This is my parameter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain array parameter description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestBodyParameterNestedSchemaRef(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/ValidationRequest'
components:
  schemas:
    ValidationRequest:
      type: object
      properties:
        level1:
          $ref: '#/components/schemas/Data'
    Data:
      type: object
      properties:
        level2:
          $ref: '#/components/schemas/NestedData'
    NestedData:
      type: object
      properties:
        level3:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--level1"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain myparameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestFormParameterDescription(t *testing.T) {
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

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "The file to upload"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain form parameter description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestRawRequestBodyShowsInputParameter(t *testing.T) {
	definition := `
paths:
  /upload:
    post:
      operationId: upload
      requestBody:
        content:
          application/octet-stream:
            schema:
              type: string
              format: binary
              description: The file to upload
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "upload", "--help"}, context)

	expected := "The file to upload"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain raw request body parameter description, expected: %v, got: %v", expected, result.StdOut)
	}

	expected = "--input"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain input parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}
