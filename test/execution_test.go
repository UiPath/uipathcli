package test

import (
	"bytes"
	"fmt"
	"net/http"
	"path/filepath"
	"runtime"
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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	if result.Error != nil {
		t.Errorf("Unexpected error, got: %v", result.Error)
	}
	if result.RequestBody != "" {
		t.Errorf("Expected empty request body, got: %v", result.RequestBody)
	}
}

func TestGetRequestShowsResponseBody(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"hello":"world"}`).
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	expectedStdOut := `{
  "hello": "world"
}
`
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestGetRequestWithDebugFlagShowsRequestAndResponseOnStdError(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"hello":"world"}`).
		Build()

	result := RunCli([]string{"myservice", "get-ping", "--debug"}, context)

	stdErr := strings.Split(result.StdErr, "\n")
	expected := "GET http://"
	if !strings.HasPrefix(stdErr[0], expected) {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[0])
	}
	expected = "User-Agent:"
	if !strings.HasPrefix(stdErr[1], expected) {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[1])
	}
	expected = "X-Request-Id:"
	if !strings.HasPrefix(stdErr[2], expected) {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[2])
	}
	expected = "HTTP/1.1 200 OK"
	if stdErr[5] != expected {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[5])
	}
	expected = "Content-Length:"
	if !strings.HasPrefix(stdErr[6], expected) {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[6])
	}
	expected = "Content-Type: text/plain; charset=utf-8"
	if stdErr[7] != expected {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[7])
	}
	expected = "Date:"
	if !strings.HasPrefix(stdErr[8], expected) {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[8])
	}
	expected = `{"hello":"world"}`
	if stdErr[10] != expected {
		t.Errorf("Expected on stderr %v, got: %v", expected, stdErr[10])
	}
}

func TestGetRequestWithDebugFlagShowsOnlyReponseBodyOnStdOut(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"hello":"world"}`).
		Build()

	result := RunCli([]string{"myservice", "get-ping", "--debug"}, context)

	expected := `{
  "hello": "world"
}
`
	if result.StdOut != expected {
		t.Errorf("Expected only response body on stdout %v, got: %v", expected, result.StdOut)
	}
}

func TestUserAgent(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

	userAgent := result.RequestHeader["user-agent"]
	expected := fmt.Sprintf("uipathcli/main (%s; %s)", runtime.GOOS, runtime.GOARCH)
	if userAgent != expected {
		t.Errorf("Could not find user-agent on header, got: %v", userAgent)
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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "get-ping"}, context)

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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "post-ping", "--first-name", "Thomas"}, context)

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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping", "--id", "my-id"}, context)

	expected := "/ping/my-id"
	if !strings.Contains(result.RequestUrl, expected) {
		t.Errorf("Request url did not contain parameter value, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestGetRequestWithCategory(t *testing.T) {
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
      tags:
        - MyCategory
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "my-category", "ping", "--id", "my-id"}, context)

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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping", "--filter", "my-filter"}, context)

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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping", "--filter", "my&filter"}, context)

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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "ping", "--x-uipath-myvalue", "custom-value"}, context)

	value := result.RequestHeader["x-uipath-myvalue"]
	expected := "custom-value"
	if value != expected {
		t.Errorf("Did not set correct custom header, expected: %v, got: %v", expected, value)
	}
}

func TestCommaSeparatedArrayArgument(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		CommaSeparatedArrayArgument(t, "string", "val1,val2", "val1,val2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		CommaSeparatedArrayArgument(t, "integer", "1,4", "1,4")
	})
	t.Run("NumberArray", func(t *testing.T) {
		CommaSeparatedArrayArgument(t, "number", "0.5,0.1,1.3", "0.5,0.1,1.3")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		CommaSeparatedArrayArgument(t, "boolean", "true,false,true", "true,false,true")
	})
}

func CommaSeparatedArrayArgument(t *testing.T, itemsType string, argument string, pathValue string) {
	definition := `
paths:
  /ping/{ids}:
    parameters:
    - name: ids
      in: path
      required: true
      description: The ids
      schema:
        type: array
        items:
          type: ` + itemsType + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--ids", argument}, context)

	expected := "/ping/" + pathValue
	if result.RequestUrl != expected {
		t.Errorf("Invalid request url, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestMultipleArgumentsArray(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		MultipleArgumentsArray(t, "string", []string{"val1", "val2"}, "val1,val2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		MultipleArgumentsArray(t, "integer", []string{"1", "4"}, "1,4")
	})
	t.Run("NumberArray", func(t *testing.T) {
		MultipleArgumentsArray(t, "number", []string{"0.5", "0.1", "1.3"}, "0.5,0.1,1.3")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		MultipleArgumentsArray(t, "boolean", []string{"true", "false", "true"}, "true,false,true")
	})
}

func MultipleArgumentsArray(t *testing.T, itemsType string, arguments []string, pathValue string) {
	definition := `
paths:
  /ping/{ids}:
    parameters:
    - name: ids
      in: path
      required: true
      description: The ids
      schema:
        type: array
        items:
          type: ` + itemsType + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	args := []string{"myservice", "ping"}
	for _, arg := range arguments {
		args = append(args, "--ids", arg)
	}
	result := RunCli(args, context)

	expected := "/ping/" + pathValue
	if result.RequestUrl != expected {
		t.Errorf("Invalid request url, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestRequestWithHeaderParameterArray(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		RequestWithHeaderParameterArray(t, "string", "val1,val2", "val1,val2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		RequestWithHeaderParameterArray(t, "integer", "1,4", "1,4")
	})
	t.Run("NumberArray", func(t *testing.T) {
		RequestWithHeaderParameterArray(t, "number", "0.5,0.1,1.3", "0.5,0.1,1.3")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		RequestWithHeaderParameterArray(t, "boolean", "true,false,true", "true,false,true")
	})
}

func RequestWithHeaderParameterArray(t *testing.T, itemsType string, argument string, headerValue string) {
	definition := `
paths:
  /ping:
    parameters:
    - name: ids
      in: header
      required: true
      description: The ids
      schema:
        type: array
        items:
          type: ` + itemsType + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--ids", argument}, context)

	header := result.RequestHeader["ids"]
	if header != headerValue {
		t.Errorf("Invalid header value, expected: %v, got: %v", headerValue, header)
	}
}

func TestRequestWithQueryStringParameterArray(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "string", "val1,val2", "id=val1&id=val2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "integer", "1,4", "id=1&id=4")
	})
	t.Run("NumberArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "number", "0.5,0.1,1.3", "id=0.5&id=0.1&id=1.3")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "boolean", "true,false,true", "id=true&id=false&id=true")
	})
}

func RequestWithQueryStringParameterArray(t *testing.T, itemsType string, argument string, queryStringValue string) {
	definition := `
paths:
  /ping:
    parameters:
    - name: id
      in: query
      required: true
      description: The id
      schema:
        type: array
        items:
          type: ` + itemsType + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--id", argument}, context)

	expected := "/ping?" + queryStringValue
	if result.RequestUrl != expected {
		t.Errorf("Invalid request url, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestRequestWithPathParameterDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { RequestWithPathParameterDataType(t, "string", "myvalue", "myvalue") })
	t.Run("Integer", func(t *testing.T) { RequestWithPathParameterDataType(t, "integer", "0", "0") })
	t.Run("Number", func(t *testing.T) { RequestWithPathParameterDataType(t, "number", "0.5", "0.5") })
	t.Run("Boolean", func(t *testing.T) { RequestWithPathParameterDataType(t, "boolean", "true", "true") })
}

func RequestWithPathParameterDataType(t *testing.T, datatype string, argument string, value string) {
	definition := `
paths:
  /ping/{id}:
    parameters:
    - name: id
      in: path
      required: true
      description: The id
      schema:
        type: ` + datatype + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--id", argument}, context)

	expected := `/ping/` + value
	if result.RequestUrl != expected {
		t.Errorf("Invalid request url, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestRequestWithQueryStringDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { RequestWithQueryStringDataTypes(t, "string", "myvalue", "myvalue") })
	t.Run("Integer", func(t *testing.T) { RequestWithQueryStringDataTypes(t, "integer", "0", "0") })
	t.Run("Number", func(t *testing.T) { RequestWithQueryStringDataTypes(t, "number", "0.5", "0.5") })
	t.Run("Boolean", func(t *testing.T) { RequestWithQueryStringDataTypes(t, "boolean", "true", "true") })
}

func RequestWithQueryStringDataTypes(t *testing.T, datatype string, argument string, value string) {
	definition := `
paths:
  /ping:
    parameters:
    - name: filter
      in: query
      required: true
      description: The filter
      schema:
        type: ` + datatype + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--filter", argument}, context)

	expected := `/ping?filter=` + value
	if result.RequestUrl != expected {
		t.Errorf("Invalid request url, expected: %v, got: %v", expected, result.RequestUrl)
	}
}

func TestRequestWithHeaderDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { RequestWithHeaderDataTypes(t, "string", "myvalue", "myvalue") })
	t.Run("Integer", func(t *testing.T) { RequestWithHeaderDataTypes(t, "integer", "0", "0") })
	t.Run("Number", func(t *testing.T) { RequestWithHeaderDataTypes(t, "number", "0.5", "0.5") })
	t.Run("Boolean", func(t *testing.T) { RequestWithHeaderDataTypes(t, "boolean", "true", "true") })
}

func RequestWithHeaderDataTypes(t *testing.T, datatype string, argument string, value string) {
	definition := `
paths:
  /ping:
    parameters:
    - name: x-header
      in: header
      required: true
      description: The filter
      schema:
        type: ` + datatype + `
    get:
      operationId: ping
      summary: Simple ping
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "ping", "--x-header", argument}, context)

	header := result.RequestHeader["x-header"]
	if header != value {
		t.Errorf("Invalid header value, expected: %v, got: %v", value, header)
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", argument}, context)

	expected := `{"myparameter":` + value + `}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestArrayDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		PostRequestArrayDataType(t, "string", "val1,val2", "[\"val1\",\"val2\"]")
	})
	t.Run("Integer", func(t *testing.T) {
		PostRequestArrayDataType(t, "integer", "0,1,2", "[0,1,2]")
	})
	t.Run("Number", func(t *testing.T) {
		PostRequestArrayDataType(t, "number", "0.5,0.1", "[0.5,0.1]")
	})
	t.Run("Boolean", func(t *testing.T) {
		PostRequestArrayDataType(t, "boolean", "true", "[true]")
	})
	t.Run("StringWithEscaping", func(t *testing.T) {
		PostRequestArrayDataType(t, "string", "val1,val\\,2", "[\"val1\",\"val,2\"]")
	})
	t.Run("StringWithDoubleEscaping", func(t *testing.T) {
		PostRequestArrayDataType(t, "string", "val1,val\\\\,2", "[\"val1\",\"val\\\\\",\"2\"]")
	})
	t.Run("StringWithMultipleEscaping", func(t *testing.T) {
		PostRequestArrayDataType(t, "string", "val1,val\\,2\\,3", "[\"val1\",\"val,2,3\"]")
	})
	t.Run("IntegerWithSpaces", func(t *testing.T) {
		PostRequestArrayDataType(t, "integer", "0, 1 , 2 ", "[0,1,2]")
	})
	t.Run("NumberWithSpaces", func(t *testing.T) {
		PostRequestArrayDataType(t, "number", "0.5, 0.1", "[0.5,0.1]")
	})
	t.Run("BooleanWithSpaces", func(t *testing.T) {
		PostRequestArrayDataType(t, "boolean", "true, false", "[true,false]")
	})
	t.Run("EmptyArray", func(t *testing.T) {
		PostRequestArrayDataType(t, "boolean", "", "[]")
	})
}

func PostRequestArrayDataType(t *testing.T, datatype string, argument string, value string) {
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
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", argument}, context)

	expected := `{"myparameter":` + value + `}`
	if result.RequestBody != expected {
		t.Errorf("Did not find array in json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestObjectType(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", `hello=world`}, context)

	expected := `{"myparameter":{"hello":"world"}}`
	if result.RequestBody != expected {
		t.Errorf("Did not find object in request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestNestedObjectType(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", `hello.a=world;hello.b=world2;foo=bar`}, context)

	expected := `{"myparameter":{"foo":"bar","hello":{"a":"world","b":"world2"}}}`
	if result.RequestBody != expected {
		t.Errorf("Did not find nested object in json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestObjectTypeConvertsNestedType(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		ObjectTypeConvertsNestedType(t, "string", "val1", "\"val1\"")
	})
	t.Run("Integer", func(t *testing.T) {
		ObjectTypeConvertsNestedType(t, "integer", "1", "1")
	})
	t.Run("Number", func(t *testing.T) {
		ObjectTypeConvertsNestedType(t, "number", "0.5", "0.5")
	})
	t.Run("Boolean", func(t *testing.T) {
		ObjectTypeConvertsNestedType(t, "boolean", "true", "true")
	})
}

func ObjectTypeConvertsNestedType(t *testing.T, datatype string, argument string, value string) {
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
                  $ref: '#/components/schemas/Data'
components:
  schemas:
    Data:
      type: object
      properties:
        myobj:
          $ref: '#/components/schemas/NestedData'
    NestedData:
      type: object
      properties:
        mykey:
          type: ` + datatype + `
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", "myobj.mykey=" + argument}, context)

	expected := `{"myparameter":{"myobj":{"mykey":` + value + `}}}`
	if result.RequestBody != expected {
		t.Errorf("Did not find typed nested object in json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestObjectTypeConvertsNestedArrayType(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		ObjectTypeConvertsNestedArrayType(t, "string", "val1,val2", "[\"val1\",\"val2\"]")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		ObjectTypeConvertsNestedArrayType(t, "integer", "1,4", "[1,4]")
	})
	t.Run("NumberArray", func(t *testing.T) {
		ObjectTypeConvertsNestedArrayType(t, "number", "0.5,0.1,1.3", "[0.5,0.1,1.3]")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		ObjectTypeConvertsNestedArrayType(t, "boolean", "true,false,true", "[true,false,true]")
	})
}

func ObjectTypeConvertsNestedArrayType(t *testing.T, datatype string, argument string, value string) {
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
                  $ref: '#/components/schemas/Data'
components:
  schemas:
    Data:
      type: object
      properties:
        myobj:
          $ref: '#/components/schemas/NestedData'
    NestedData:
      type: object
      properties:
        mykey:
          type: array
          items:
            type: ` + datatype + `
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", "myobj.mykey=" + argument}, context)

	expected := `{"myparameter":{"myobj":{"mykey":` + value + `}}}`
	if result.RequestBody != expected {
		t.Errorf("Did not find typed nested object array in json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestArrayObjectType(t *testing.T) {
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
                    type: object
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", "hello=world", "--myparameter", "other=object"}, context)

	expected := `{"myparameter":[{"hello":"world"},{"other":"object"}]}`
	if result.RequestBody != expected {
		t.Errorf("Did not find object array in json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostFormRequest(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	path := CreateTempFile(t, "hello-world")
	result := RunCli([]string{"myservice", "post-validate", "--file", path}, context)

	contentType := result.RequestHeader["content-type"]
	expected := "multipart/form-data; boundary="
	if !strings.Contains(contentType, expected) {
		t.Errorf("Did not set correct content type, expected: %v, got: %v", expected, contentType)
	}
	expected = fmt.Sprintf(`Content-Disposition: form-data; name="file"; filename="%s"`, filepath.Base(path))
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find Content-Disposition in body, expected: %v, got: %v", expected, result.RequestBody)
	}
	expected = `Content-Type: application/octet-stream`
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find Content-Type in body, expected: %v, got: %v", expected, result.RequestBody)
	}
	expected = `hello-world`
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find content in body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostFormRequestWithUppercaseFile(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      requestBody:
        content:
          multipart/form-data:
            schema:
              properties:
                File:
                  type: string
                  format: binary
                  description: The file to upload
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	path := CreateTempFile(t, "hello-world")
	result := RunCli([]string{"myservice", "post-validate", "--file", path}, context)

	expected := fmt.Sprintf(`Content-Disposition: form-data; name="File"; filename="%s"`, filepath.Base(path))
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find Content-Disposition in body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostFormRequestFromFileReference(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()
	path := CreateTempFile(t, "hello-world")
	result := RunCli([]string{"myservice", "post-validate", "--file", path}, context)

	expected := `Content-Disposition: form-data; name="file"; filename="` + filepath.Base(path) + `"`
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find Content-Disposition in body, expected: %v, got: %v", expected, result.RequestBody)
	}
	expected = `Content-Type: application/octet-stream`
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find Content-Type in body, expected: %v, got: %v", expected, result.RequestBody)
	}
	expected = `hello-world`
	if !strings.Contains(result.RequestBody, expected) {
		t.Errorf("Did not find content in body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostFormRequestWithLargeFile(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	path := CreateTempFileBinary(t, make([]byte, 10*1024*1024))

	result := RunCli([]string{"myservice", "post-validate", "--file", path}, context)

	if len(result.RequestBody) < 10*1024*1024 {
		t.Errorf("Request body does not contain large file, got only %v bytes", len(result.RequestBody))
	}
}

func TestPostRequestWithDefaultValueWhenRequired(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                firstName:
                  type: string
                  default: 'my-name'
              required:
                - firstName
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create"}, context)

	expected := `{"firstName":"my-name"}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestOmitDefaultValueWhenOptional(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                firstName:
                  type: string
                  default: 'first-name'
                lastName:
                  type: string
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create", "--last-name", "last-name"}, context)

	expected := `{"lastName":"last-name"}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestIgnoreDefaultValueWhenProvided(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                firstName:
                  type: string
                  default: 'my-name'
              required:
                - firstName
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create", "--first-name", "provided-name"}, context)

	expected := `{"firstName":"provided-name"}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestUsesStdIn(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                firstName:
                  type: string
                  default: 'my-name'
              required:
                - firstName
`
	stdIn := bytes.Buffer{}
	stdIn.WriteString(`{"firstName":"foo"}`)
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithStdIn(stdIn).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create", "--file", "-"}, context)

	expected := `{"firstName":"foo"}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestWithStdInAndHeaderParameter(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      parameters:
      - name: x-uipath-myvalue
        in: header
        required: true
        schema:
          type: string
`
	stdIn := bytes.Buffer{}
	stdIn.WriteString(`{"foo":"bar"}`)
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithStdIn(stdIn).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create", "--x-uipath-myvalue", "test-value", "--file", "-"}, context)

	expectedBody := `{"foo":"bar"}`
	if result.RequestBody != expectedBody {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expectedBody, result.RequestBody)
	}

	header := result.RequestHeader["x-uipath-myvalue"]
	expectedHeader := "test-value"
	if header != expectedHeader {
		t.Errorf("Did not set correct custom header, expected: %v, got: %v", expectedHeader, header)
	}
}

func TestPostRequestWithStdInAndBodyParameterIgnoresStdIn(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                firstName:
                  type: string
`
	stdIn := bytes.Buffer{}
	stdIn.WriteString(`{"foo":"bar"}`)
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithStdIn(stdIn).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create", "--first-name", "test-value"}, context)

	expectedBody := `{"firstName":"test-value"}`
	if result.RequestBody != expectedBody {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expectedBody, result.RequestBody)
	}
}

func TestPostWithFileAsRawRequestBody(t *testing.T) {
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
		WithResponse(http.StatusOK, "").
		Build()

	path := CreateTempFile(t, "hello-world")
	result := RunCli([]string{"myservice", "upload", "--file", path}, context)

	contentType := result.RequestHeader["content-type"]
	if contentType != "application/octet-stream" {
		t.Errorf("Content-Type is not application/octet-stream, got: %v", contentType)
	}
	contentLength := result.RequestHeader["content-length"]
	if contentLength != "11" {
		t.Errorf("Content-Length is not 11, got: %v", contentLength)
	}
	if result.RequestBody != "hello-world" {
		t.Errorf("Request body is not as expected, got: %v", result.RequestBody)
	}
}

func TestPostWithRelativeFilePathAsRawRequestBody(t *testing.T) {
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
		WithResponse(http.StatusOK, "").
		Build()

	workingDirectory := t.TempDir()
	t.Chdir(workingDirectory)
	path := CreateTempFile(t, "hello-world")
	relativePath, _ := filepath.Rel(workingDirectory, path)
	result := RunCli([]string{"myservice", "upload", "--file", relativePath}, context)

	contentType := result.RequestHeader["content-type"]
	if contentType != "application/octet-stream" {
		t.Errorf("Content-Type is not application/octet-stream, got: %v", contentType)
	}
	if result.RequestBody != "hello-world" {
		t.Errorf("Request body is not as expected, got: %v", result.RequestBody)
	}
}

func TestPostAllOfParameter(t *testing.T) {
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
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--filter", "my-filter"}, context)

	if result.RequestUrl != "/validate?filter=my-filter" {
		t.Errorf("Url does not contain filter from allOf schema, got: %v", result.RequestUrl)
	}
}

func TestPostEnumParameter(t *testing.T) {
	definition := `
paths:
  /validate:
    post:
      parameters:
      - name: type
        in: query
        required: true
        description: The type 
        schema:
          type: string
          enum:
            - username
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--type", "username"}, context)

	if result.RequestUrl != "/validate?type=username" {
		t.Errorf("Url does not contain enum value, got: %v", result.RequestUrl)
	}
}

func TestPostRequestEnumParameterSuccessfully(t *testing.T) {
	definition := `
paths:
  /user:
    post:
      operationId: create-user
      requestBody:
        content:
          application/json:
            schema:
              properties:
                region:
                  type: integer
                  enum:
                    - 1
                    - 2
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create-user", "--region", "2"}, context)

	expected := `{"region":2}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestRequiredEnumParameterWithDefaultValue(t *testing.T) {
	definition := `
paths:
  /user:
    post:
      operationId: create-user
      requestBody:
        content:
          application/json:
            schema:
              properties:
                region:
                  type: integer
                  default: 2
                  enum:
                    - 1
                    - 2
              required:
                - region
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create-user"}, context)

	expected := `{"region":2}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostUrlEncodedRequest(t *testing.T) {
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
                client_id:
                  type: string
                  description: The client id
                client_secret:
                  type: string
                  description: The client secret
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "validate", "--client-id", "my-client-id", "--client-secret", "my-client-secret"}, context)

	contentType := result.RequestHeader["content-type"]
	if contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Did not set x-www-form-urlencoded content type, got: %v", contentType)
	}
	expected := "client_id=my-client-id&client_secret=my-client-secret"
	if result.RequestBody != expected {
		t.Errorf("Did not find url encoded data in body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostUrlEncodedEscapesDataRequest(t *testing.T) {
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
                myparam:
                  type: string
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "validate", "--myparam", "hello & world"}, context)

	expected := "myparam=hello+%26+world"
	if result.RequestBody != expected {
		t.Errorf("Url encoded data is not escaped, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostUrlEncodedDataTypes(t *testing.T) {
	t.Run("String", func(t *testing.T) { PostUrlEncodedDataTypes(t, "string", "myvalue", "myvalue") })
	t.Run("Integer", func(t *testing.T) { PostUrlEncodedDataTypes(t, "integer", "0", "0") })
	t.Run("Number", func(t *testing.T) { PostUrlEncodedDataTypes(t, "number", "0.5", "0.5") })
	t.Run("Boolean", func(t *testing.T) { PostUrlEncodedDataTypes(t, "boolean", "true", "true") })
}

func PostUrlEncodedDataTypes(t *testing.T, datatype string, argument string, value string) {
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
                myparam:
                  type: ` + datatype + `
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "validate", "--myparam", argument}, context)
	expected := "myparam=" + value
	if result.RequestBody != expected {
		t.Errorf("Wrong data type conversion for url encoded data, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestSameParameterNameIsOnlyDefinedOnce(t *testing.T) {
	definition := `
paths:
  /create/{my-param}:
    get:
      operationId: create
      parameters:
        - name: my-param
          in: path
          required: true
          schema:
            type: string
      requestBody:
        content:
          application/json:
            schema:
              properties:
                my-param:
                  type: string
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, "").
		Build()

	result := RunCli([]string{"myservice", "create", "--my-param", "my-value"}, context)

	if result.RequestUrl != "/create/my-value" {
		t.Errorf("Expected parameter in request url, but got: %v", result.RequestUrl)
	}
	if result.RequestBody != `{"my-param":"my-value"}` {
		t.Errorf("Expected parameter in request body, but got: %v", result.RequestBody)
	}
}

func TestPostObjectTypeWithSpaces(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", `foo = a; bar=b `}, context)

	expected := `{"myparameter":{"bar":"b","foo":"a"}}`
	if result.RequestBody != expected {
		t.Errorf("Did not find nested object in json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostObjectTypeUsesRawPropertyName(t *testing.T) {
	definition := `
paths:
  /create:
    post:
      operationId: create
      requestBody:
        content:
          application/json:
            schema:
              properties:
                product:
                  $ref: '#/components/schemas/Product'
components:
  schemas:
    Product:
      type: object
      properties:
        totalPrice:
          type: number
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "create", "--product", `totalPrice=1.3`}, context)

	expected := `{"product":{"totalPrice":1.3}}`
	if result.RequestBody != expected {
		t.Errorf("Object type with custom parameter name not found in request body, expected: %v, but got: %v", expected, result.RequestBody)
	}
}

func TestJsonInputForObjectType(t *testing.T) {
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
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", `{ "hello": "world" }`}, context)

	expected := `{"myparameter":{"hello":"world"}}`
	if result.RequestBody != expected {
		t.Errorf("Did not find json object in request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestMultipleJsonObjectArgumentsForObjectArrayType(t *testing.T) {
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
                    type: object
`
	context := NewContextBuilder().
		WithResponse(http.StatusOK, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"myservice", "post-validate", "--myparameter", `{ "foo": 1 }`, "--myparameter", `{ "bar": 2 }`}, context)

	expected := `{"myparameter":[{"foo":1},{"bar":2}]}`
	if result.RequestBody != expected {
		t.Errorf("Did not find json object in request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestRetriesServerErrorsUntilSuccess(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	callCount := 0
	context := NewContextBuilder().
		WithDefinition("service", definition).
		WithResponseHandler(func(request RequestData) ResponseData {
			callCount++
			if callCount == 3 {
				return ResponseData{Status: http.StatusOK, Body: `{"hello":"world"}`}
			}
			return ResponseData{Status: http.StatusInternalServerError, Body: "Internal Server Error"}
		}).
		Build()

	result := RunCli([]string{"service", "ping", "--debug"}, context)

	if result.Error != nil {
		t.Errorf("Expected no error after retries, but got: %v", result.Error)
	}

	expectedStdOut := `{
  "hello": "world"
}
`
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}

	responseErrorCount := strings.Count(result.StdErr, "HTTP/1.1 500 Internal Server Error")
	if responseErrorCount != 2 {
		t.Errorf("Expected 2 response errors, but got: %v", result.StdErr)
	}
	responseSuccessCount := strings.Count(result.StdErr, "HTTP/1.1 200 OK")
	if responseSuccessCount != 1 {
		t.Errorf("Expected 1 success response, but got: %v", result.StdErr)
	}
}

func TestRetriesServerErrorsThreeTimes(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("service", definition).
		WithResponse(http.StatusInternalServerError, "Internal Server Error").
		Build()

	result := RunCli([]string{"service", "ping", "--debug"}, context)

	if result.Error == nil || result.Error.Error() != "Service returned status code '500' and body 'Internal Server Error'" {
		t.Errorf("Expected service error, but got: %v", result.Error)
	}
	responseErrorCount := strings.Count(result.StdErr, "HTTP/1.1 500 Internal Server Error")
	if responseErrorCount != 3 {
		t.Errorf("Expected 3 response errors, but got: %v", result.StdErr)
	}
}
