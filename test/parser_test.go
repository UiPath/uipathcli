package commandline

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
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: my-ping-operation
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice"}, context)

	expected := "my-ping-operation"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain custom operation, expected: %v, got: %v", expected, result.StdOut)
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

func TestSnakeCaseBodyParameter(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myParameter:
                  type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--my-parameter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain snake cased parameter, expected: %v, got: %v", expected, result.StdOut)
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
