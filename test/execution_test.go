package commandline

import (
	"strings"
	"testing"
)

func TestGetRequestSuccessfully(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "get-ping"}, context)

	if result.Error != nil {
		t.Errorf("Unexpected error, got: %v", result.Error)
	}
	if result.RequestBody != "" {
		t.Errorf("Expected empty request body, got: %v", result.RequestBody)
	}
}

func TestRequestId(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "get-ping"}, context)

	requestId := result.RequestHeader["x-request-id"]
	if len(requestId) != 32 {
		t.Errorf("Could not find request id on header, got: %v", requestId)
	}
}

func TestPostRequestSuccessfully(t *testing.T) {
	definition := `
paths:
  /ping:
    post:
      summary: Simple ping
      requestBody:
        content:	  
          application/json:
            schema:
              properties:
                firstName:
                  type: string
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "post-ping", "--first-name", "Thomas"}, context)

	contentType := result.RequestHeader["content-type"]
	expected := "application/json"
	if contentType != expected {
		t.Errorf("Did not set correct content type, expected: %v, got: %v", expected, contentType)
	}

	expected = `{"firstName":"Thomas"}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestGetRequestWithPathParameter(t *testing.T) {
	definition := `
paths:
  /ping/{id}:
    parameters:
    - name: id
      in: path
      required: true
      description: The id
      schema:
        type: string
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "ping", "--id", "my-id"}, context)

	expected := "/ping/my-id"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Request url did not contain parameter value, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestGetRequestWithQueryParameter(t *testing.T) {
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

func TestGetRequestWithEscapedQueryParameter(t *testing.T) {
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

	result := runCli([]string{"myservice", "ping", "--filter", "my&filter"}, context)

	expected := "/ping?filter=my%26filter"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Request url did not contain query string, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestGetRequestWithHeaderParameter(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
      parameters:
      - name: x-uipath-myvalue
        in: header
        required: true
        description: The filter
        schema:
          type: string
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "ping", "--x-uipath-myvalue", "custom-value"}, context)

	value := result.RequestHeader["x-uipath-myvalue"]
	expected := "custom-value"
	if value != expected {
		t.Errorf("Did not set correct custom header, expected: %v, got: %v", expected, value)
	}
}

func TestPostRequestDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { PostRequestDataType(t, "string", "myvalue", "\"myvalue\"") })
	t.Run("Integer", func(t *testing.T) { PostRequestDataType(t, "integer", "0", "0") })
	t.Run("Number", func(t *testing.T) { PostRequestDataType(t, "number", "0.5", "0.5") })
	t.Run("Boolean", func(t *testing.T) { PostRequestDataType(t, "boolean", "true", "true") })
}

func PostRequestDataType(t *testing.T, datatype string, argument string, value string) {
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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", argument}, context)

	expected := `{"myparameter":` + value + `}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}
