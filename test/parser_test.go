package test

import (
	"strings"
	"testing"
)

func TestInvalidDefinitionReturnsError(t *testing.T) {
	definition := `
paths: INVALID DEFINITION
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "--help"}, context)

	expected := "Error parsing definition file 'myservice'"
	if !strings.HasPrefix(result.StdErr, expected) {
		t.Errorf("Stderr did not contain definition parsing error, expected: %v, got: %v", expected, result.StdErr)
	}
}

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

	result := RunCli([]string{"myservice"}, context)

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

	result := RunCli([]string{"testservice", "--help"}, context)

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

	result := RunCli([]string{"myservice", "--help"}, context)

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

	result := RunCli([]string{"myservice"}, context)

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

	result := RunCli([]string{"myservice"}, context)

	if !strings.Contains(result.StdOut, expectedName) {
		t.Errorf("stdout does not contain custom operation, expected: %v, got: %v", expectedName, result.StdOut)
	}
}

func TestOperationsSummary(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice"}, context)

	expected := "Simple ping"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain operation summaries, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestOperationSummary(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--help"}, context)

	expected := "Simple ping"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain ping operation summary, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestOperationDescription(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      description: This is a long description
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--help"}, context)

	expected := "This is a long description"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain ping operation description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestCategory(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      tags:
        - MyCategory
      summary: Simple ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice"}, context)

	expected := "my-category"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain category, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestCategoryCommands(t *testing.T) {
	definition := `
paths:
  /:
    post:
      tags:
      - MyCategory
      operationId: create
    get:
      tags:
        - MyCategory
      operationId: list
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "my-category"}, context)

	if !strings.Contains(result.StdOut, "create") || !strings.Contains(result.StdOut, "list") {
		t.Errorf("stdout does not contain all commands, got: %v", result.StdOut)
	}
}

func TestCategoryDescription(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      tags:
        - MyCategory
      summary: Simple ping
tags:
- name: MyCategory
  description: This is a description for my category
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "my-category"}, context)

	expected := "This is a description for my category"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain category description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestCategoryCommandsSorted(t *testing.T) {
	definition := `
paths:
  /:
    post:
      tags:
      - MyCategory
      operationId: bbbbb
    get:
      tags:
        - MyCategory
      operationId: aaaaa
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "my-category"}, context)

	if strings.Index(result.StdOut, "aaaaa") >= strings.Index(result.StdOut, "bbbbb") {
		t.Errorf("category commands are not sorted, got: %v", result.StdOut)
	}
}

func TestCategoryMixedWithNoCategory(t *testing.T) {
	definition := `
paths:
  /:
    post:
      operationId: simple
    get:
      tags:
        - MyCategory
      operationId: inside-category
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice"}, context)

	if !strings.Contains(result.StdOut, "simple") {
		t.Errorf("Does not contain command outside of category, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "my-category") {
		t.Errorf("Does not contain category, got: %v", result.StdOut)
	}
	if strings.Contains(result.StdOut, "inside-category") {
		t.Errorf("Should not contain command inside of the category, got: %v", result.StdOut)
	}
}

func TestMultipleParametersSortedByName(t *testing.T) {
	definition := `
paths:
  /resource:
    get:
      summary: list
      operationId: list
      parameters:
      - name: bbbbb
      - name: aaaaa
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	if strings.Index(result.StdOut, "aaaaa") >= strings.Index(result.StdOut, "bbbbb") {
		t.Errorf("Expected aaaaa argument to be first, got: %v", result.StdOut)
	}
}

func TestMultipleParametersSortedRequiredFirst(t *testing.T) {
	definition := `
paths:
  /resource:
    get:
      summary: list
      operationId: list
      parameters:
      - name: bbbbb
      - name: aaaaa
      - name: ccccc
        required: true
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	if strings.Index(result.StdOut, "ccccc") >= strings.Index(result.StdOut, "aaaaa") {
		t.Errorf("Expected required argument to be before aaaaa argument, got: %v", result.StdOut)
	}
	if strings.Index(result.StdOut, "ccccc") >= strings.Index(result.StdOut, "bbbbb") {
		t.Errorf("Expected required argument to be before bbbbb argument, got: %v", result.StdOut)
	}
}

func TestMultipleRequiredParametersSortedByName(t *testing.T) {
	definition := `
paths:
  /resource:
    get:
      summary: list
      operationId: list
      parameters:
      - name: bbbbb
        required: true
      - name: aaaaa
        required: true
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	if strings.Index(result.StdOut, "aaaaa") >= strings.Index(result.StdOut, "bbbbb") {
		t.Errorf("Expected required argument to be first, got: %v", result.StdOut)
	}
}

func TestOperationWithCustomName(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: other
      x-uipathcli-name: create-event
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "--help"}, context)

	expected := "create-event"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("Stdout did not contain custom operation name, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestParameterWithCustomName(t *testing.T) {
	definition := `
paths:
  /resource:
    get:
      operationId: list
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter 
        schema:
          type: string
        x-uipathcli-name: my-custom-parameter-name
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	expected := "my-custom-parameter-name"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("Stdout did not contain custom parameter name, expected: %v, got: %v", expected, result.StdOut)
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

	result := RunCli([]string{"myservice", "ping", "--filter", "my-filter"}, context)

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

	result := RunCli([]string{"myservice", "ping", "--filter", "my-filter"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myparameter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain request body parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestPropertyWithCustomName(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                myparameter:
                  type: string
                  description: This is my parameter
                  x-uipathcli-name: my-custom-parameter-name
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "create", "--help"}, context)

	expected := "my-custom-parameter-name"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("Stdout did not contain custom parameter name, expected: %v, got: %v", expected, result.StdOut)
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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myparameter string (required)"
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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myparameter string (default: 1)"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain request body parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestHelpShowsParameterWithSingleEnumValueAsDefault(t *testing.T) {
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
                  enum:
                  - '1'
              required:
                - myparameter
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--myparameter string (default: 1)"
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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-data", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--level1"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain myparameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestEnumParameter(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: list
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter 
        schema:
          type: string
          enum:
          - Value1
          - Value2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	if !strings.Contains(result.StdOut, "Allowed values:") {
		t.Errorf("stdout does not contain allowed values, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- Value1") {
		t.Errorf("stdout does not contain first allowed value, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- Value2") {
		t.Errorf("stdout does not contain second allowed value, got: %v", result.StdOut)
	}
}

func TestEnumSchema(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: list
      parameters:
      - name: filter
        in: query
        description: The filter 
        schema:
          $ref: '#/components/schemas/FilterType'
components:
  schemas:
    FilterType:
      type: integer
      enum:
        - 0
        - 1
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	if !strings.Contains(result.StdOut, "Allowed values:") {
		t.Errorf("stdout does not contain allowed values, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- 0") {
		t.Errorf("stdout does not contain first allowed value, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- 1") {
		t.Errorf("stdout does not contain second allowed value, got: %v", result.StdOut)
	}
}

func TestInvalidEnumValueReturnsValidationError(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: list
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter 
        schema:
          type: string
          enum:
          - Value1
          - Value2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--filter", "other-value"}, context)

	expected := "Argument value 'other-value' for --filter is invalid, allowed values: Value1, Value2"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr should show validation error for invalid enum value, got: %v", result.StdErr)
	}
}

func TestInvalidEnumIntegerValueReturnsValidationError(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: list
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter 
        schema:
          type: integer
          enum:
          - 1
          - 2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--filter", "3"}, context)

	expected := "Argument value '3' for --filter is invalid, allowed values: 1, 2"
	if !strings.Contains(result.StdErr, expected) {
		t.Errorf("stderr should show validation error for invalid enum value, got: %v", result.StdErr)
	}
}

func TestEnumAllOf(t *testing.T) {
	definition := `
paths:
  /resource:
    post:
      operationId: list
      parameters:
      - name: filter
        in: query
        description: The filter
        schema:
          allOf:
            - $ref: '#/components/schemas/FilterType'
components:
  schemas:
    FilterType:
      type: string
      enum:
        - Value1
        - Value2
      default: my-default
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "list", "--help"}, context)

	if !strings.Contains(result.StdOut, "Allowed values:") {
		t.Errorf("stdout does not contain allowed values, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- Value1") {
		t.Errorf("stdout does not contain first allowed value, got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "- Value2") {
		t.Errorf("stdout does not contain second allowed value, got: %v", result.StdOut)
	}
}

func TestParameterAllOf(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      parameters:
      - name: filter
        in: query
        required: true
        description: The filter 
        schema:
          allOf:
            - $ref: '#/components/schemas/FilterType'
components:
  schemas:
    FilterType:
      type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--filter"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain filter parameter from allOf schema, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestBodyParameterAllOf(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          application/json:
            schema:
              allOf:
                - $ref: '#/components/schemas/ValidationRequest'
components:
  schemas:
    ValidationRequest:
      type: object
      properties:
        name:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "--name"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain name parameter from allOf schema, expected: %v, got: %v", expected, result.StdOut)
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

	result := RunCli([]string{"myservice", "post-validate", "--help"}, context)

	expected := "The file to upload"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain form parameter description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestRawRequestBodyShowsFileParameter(t *testing.T) {
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

	result := RunCli([]string{"myservice", "upload", "--help"}, context)

	expected := "The file to upload"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain raw request body parameter description, expected: %v, got: %v", expected, result.StdOut)
	}

	expected = "--file"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain input parameter, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestMultipleDefinitions(t *testing.T) {
	definition1 := `
paths:
  /create:
    post:
      summary: Create a resource
`
	definition2 := `
paths:
  /update:
    post:
      summary: Update a resource
`
	context := NewContextBuilder().
		WithDefinition("myservice1", definition1).
		WithDefinition("myservice2", definition2).
		Build()

	result := RunCli([]string{"--help"}, context)

	expected := "myservice1"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain service name from first definition, expected: %v, got: %v", expected, result.StdOut)
	}
	expected = "myservice2"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain service name from second definition, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestMergesMultipleDefinitionsWithSameNamePrefix(t *testing.T) {
	definition1 := `
paths:
  /create:
    post:
      summary: Create a resource
`
	definition2 := `
paths:
  /update:
    post:
      summary: Update a resource
`
	context := NewContextBuilder().
		WithDefinition("myapp.myservice1", definition1).
		WithDefinition("myapp.myservice2", definition2).
		Build()

	result := RunCli([]string{"myapp", "--help"}, context)

	expected := "Create a resource"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain operation from first definition, expected: %v, got: %v", expected, result.StdOut)
	}
	expected = "Update a resource"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain operation from second definition, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestCategoriesFromMultipleDefinitionsWithSameNamePrefix(t *testing.T) {
	definition1 := `
paths:
  /resource:
    post:
      tags:
        - FirstCategory
      summary: Create a resource
`
	definition2 := `
paths:
  /resource:
    delete:
      tags:
        - SecondCategory
      summary: Delete a resource
`
	context := NewContextBuilder().
		WithDefinition("myapp.myservice1", definition1).
		WithDefinition("myapp.myservice2", definition2).
		Build()

	result := RunCli([]string{"myapp", "--help"}, context)

	expected := "first-category"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain category from first definition, expected: %v, got: %v", expected, result.StdOut)
	}
	expected = "second-category"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain category from second definition, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestMergesCategoriesFromMultipleDefinitionsWithSameNamePrefix(t *testing.T) {
	definition1 := `
paths:
  /create:
    post:
      tags:
        - CommonCategory
      summary: Create a resource
`
	definition2 := `
paths:
  /update:
    post:
      tags:
        - CommonCategory
      summary: Update a resource
  /delete:
    post:
      tags:
        - AdditionalCategory
      summary: Delete a resource
`
	context := NewContextBuilder().
		WithDefinition("myapp.myservice1", definition1).
		WithDefinition("myapp.myservice2", definition2).
		Build()

	result := RunCli([]string{"myapp", "common-category", "--help"}, context)

	expected := "Create a resource"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain operation from first definition, expected: %v, got: %v", expected, result.StdOut)
	}
	expected = "Update a resource"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain operation from second definition, expected: %v, got: %v", expected, result.StdOut)
	}
	expected = "Delete a resource"
	if strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout contains operation from wrong category, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestUrlEncodedParameterDescription(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      operationId: validate
      requestBody:
        content:
          application/x-www-form-urlencoded:
            schema:
              properties:
                username:
                  type: string
                  description: The user name
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "validate", "--help"}, context)

	expected := "The user name"
	if !strings.Contains(result.StdOut, expected) {
		t.Errorf("stdout does not contain form parameter description, expected: %v, got: %v", expected, result.StdOut)
	}
}

func TestShowsSpecifiedVersionServiceDefinition(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping-v1
`
	definition2_0 := `
paths:
  /ping:
    get:
      operationId: ping-v2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithDefinitionVersion("myversionedservice", "2.0", definition2_0).
		Build()

	result := RunCli([]string{"--version", "2.0", "--help"}, context)

	if !strings.Contains(result.StdOut, "myversionedservice") {
		t.Errorf("Could not find versioned service definition, but got: %v", result.StdOut)
	}
	if strings.Contains(result.StdOut, "myservice") {
		t.Errorf("Should not parse default service definitions, but got: %v", result.StdOut)
	}
}

func TestShowsDefaultVersionServiceDefinitions(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping-v1
`
	definition2_0 := `
paths:
  /ping:
    get:
      operationId: ping-v2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithDefinitionVersion("myversionedservice", "2.0", definition2_0).
		Build()

	result := RunCli([]string{"--help"}, context)

	if !strings.Contains(result.StdOut, "myservice") {
		t.Errorf("Could not find default service definition, but got: %v", result.StdOut)
	}
	if strings.Contains(result.StdOut, "myversionedservice") {
		t.Errorf("Should not parse versioned service definitions, but got: %v", result.StdOut)
	}
}

func TestShowsSpecifiedVersionServiceOperations(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping-v1
`
	definition2_0 := `
paths:
  /ping:
    get:
      operationId: ping-v2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithDefinitionVersion("myversionedservice", "2.0", definition2_0).
		Build()

	result := RunCli([]string{"myversionedservice", "--version", "2.0", "--help"}, context)

	if !strings.Contains(result.StdOut, "ping-v2") {
		t.Errorf("Could not find versioned service operation, but got: %v", result.StdOut)
	}
	if strings.Contains(result.StdOut, "ping-v1") {
		t.Errorf("Should not parse default service operation, but got: %v", result.StdOut)
	}
}

func TestShowsDefaultVersionServiceOperations(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping-v1
`
	definition2_0 := `
paths:
  /ping:
    get:
      operationId: ping-v2
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithDefinitionVersion("myversionedservice", "2.0", definition2_0).
		Build()

	result := RunCli([]string{"myservice", "--help"}, context)

	if !strings.Contains(result.StdOut, "ping-v1") {
		t.Errorf("Could not find default service operation, but got: %v", result.StdOut)
	}
	if strings.Contains(result.StdOut, "ping-v2") {
		t.Errorf("Should not parse versioned service operation, but got: %v", result.StdOut)
	}
}

func TestShowsExampleOutputForObjects(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateDeploymentRequest'
components:
  schemas:
    CreateDeploymentRequest:
      type: object
      properties:
        config:
          $ref: '#/components/schemas/DeploymentConfigModel'
    DeploymentConfigModel:
      type: object
      properties:
        cpu:
          type: integer
          format: int32
        memory:
          type: number
          format: double
        gpu:
          type: boolean
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "create", "--help"}, context)

	if !strings.Contains(result.StdOut, "Example:") {
		t.Errorf("Could not find example output, but got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "cpu=integer; gpu=boolean; memory=float") {
		t.Errorf("Could not find example command, but got: %v", result.StdOut)
	}
}

func TestShowsExampleOutputForNestedObjects(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateDeploymentRequest'
components:
  schemas:
    CreateDeploymentRequest:
      type: object
      properties:
        instances:
          type: array
          items:
            $ref: '#/components/schemas/DeploymentInstance'
    DeploymentInstance:
      type: object
      properties:
        config:
          $ref: '#/components/schemas/DeploymentConfigModel'
    DeploymentConfigModel:
      type: object
      properties:
        cpu:
          type: integer
          format: int32
        memory:
          type: number
          format: double
        gpu:
          type: boolean
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "create", "--help"}, context)

	if !strings.Contains(result.StdOut, "Example:") {
		t.Errorf("Could not find example output, but got: %v", result.StdOut)
	}
	if !strings.Contains(result.StdOut, "config.cpu=integer; config.gpu=boolean; config.memory=float") {
		t.Errorf("Could not find example command, but got: %v", result.StdOut)
	}
	if strings.Contains(result.StdOut, "config=object") {
		t.Errorf("Should not contain the object itself in the example, but got: %v", result.StdOut)
	}
}
