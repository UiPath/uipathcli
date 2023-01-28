package test

import (
	"bytes"
	"os"
	"path/filepath"
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

func TestGetRequestShowsResponseBody(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, `{"hello":"world"}`).
		Build()

	result := runCli([]string{"myservice", "get-ping"}, context)

	expectedStdOut := `{
  "hello": "world"
}
`
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestGetRequestWithDebugFlagShowsRequestAndResponse(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, `{"hello":"world"}`).
		Build()

	result := runCli([]string{"myservice", "get-ping", "--debug"}, context)

	stdout := strings.Split(result.StdOut, "\n")
	expected := "GET http://"
	if !strings.HasPrefix(stdout[0], expected) {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[0])
	}
	expected = "X-Request-Id:"
	if !strings.HasPrefix(stdout[1], expected) {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[1])
	}
	expected = "HTTP/1.1 200 OK"
	if stdout[4] != expected {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[4])
	}
	expected = "Content-Length:"
	if !strings.HasPrefix(stdout[5], expected) {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[5])
	}
	expected = "Content-Type: text/plain; charset=utf-8"
	if stdout[6] != expected {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[6])
	}
	expected = "Date:"
	if !strings.HasPrefix(stdout[7], expected) {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[7])
	}
	expected = `{"hello":"world"}`
	if stdout[9] != expected {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[9])
	}
	expected = "{"
	if stdout[12] != expected {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[12])
	}
	expected = `  "hello": "world"`
	if stdout[13] != expected {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[13])
	}
	expected = "}"
	if stdout[14] != expected {
		t.Errorf("Expected on stdout %v, got: %v", expected, stdout[14])
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

func TestRequestWithPathParameterArray(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		RequestWithPathParameterArray(t, "string", "val1,val2", "val1,val2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		RequestWithPathParameterArray(t, "integer", "1,4", "1,4")
	})
	t.Run("NumberArray", func(t *testing.T) {
		RequestWithPathParameterArray(t, "number", "0.5,0.1,1.3", "0.5,0.1,1.3")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		RequestWithPathParameterArray(t, "boolean", "true,false,true", "true,false,true")
	})
}

func RequestWithPathParameterArray(t *testing.T, itemsType string, argument string, pathValue string) {
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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "ping", "--ids", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "ping", "--ids", argument}, context)

	header := result.RequestHeader["ids"]
	if header != headerValue {
		t.Errorf("Invalid header value, expected: %v, got: %v", headerValue, header)
	}
}

func TestRequestWithQueryStringParameterArray(t *testing.T) {
	t.Run("StringArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "string", "val1,val2", "id=val1%2Cid=val2")
	})
	t.Run("IntegerArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "integer", "1,4", "id=1%2Cid=4")
	})
	t.Run("NumberArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "number", "0.5,0.1,1.3", "id=0.5%2Cid=0.1%2Cid=1.3")
	})
	t.Run("BooleanArray", func(t *testing.T) {
		RequestWithQueryStringParameterArray(t, "boolean", "true,false,true", "id=true%2Cid=false%2Cid=true")
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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "ping", "--id", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "ping", "--id", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "ping", "--filter", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "ping", "--x-header", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", `hello=world`}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", `hello.a=world;hello.b=world2;foo=bar`}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", "myobj.mykey=" + argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", "myobj.mykey=" + argument}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--myparameter", `hello=world,other=object`}, context)

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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"myservice", "post-validate", "--file", "hello-world"}, context)

	contentType := result.RequestHeader["content-type"]
	expected := "multipart/form-data; boundary="
	if !strings.Contains(contentType, expected) {
		t.Errorf("Did not set correct content type, expected: %v, got: %v", expected, contentType)
	}
	expected = `Content-Disposition: form-data; name="file"; filename="file"`
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
		WithResponse(200, "{}").
		WithDefinition("myservice", definition).
		Build()
	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)
	result := runCli([]string{"myservice", "post-validate", "--file", "file://" + path}, context)

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
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "create"}, context)

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
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "create", "--last-name", "last-name"}, context)

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
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "create", "--first-name", "provided-name"}, context)

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
	stdIn.Write([]byte(`{"firstName":"foo"}`))
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithStdIn(stdIn).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "create"}, context)

	expected := `{"firstName":"foo"}`
	if result.RequestBody != expected {
		t.Errorf("Invalid json request body, expected: %v, got: %v", expected, result.RequestBody)
	}
}

func TestPostRequestWithStdInAndParameter(t *testing.T) {
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
	stdIn.Write([]byte(`{"foo":"bar"}`))
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithStdIn(stdIn).
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "create", "--x-uipath-myvalue", "test-value"}, context)

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
		WithResponse(200, "").
		Build()

	path := createFile(t)
	os.WriteFile(path, []byte("hello-world"), 0644)
	result := runCli([]string{"myservice", "upload", "--input", "file://" + path}, context)

	contentType := result.RequestHeader["content-type"]
	if contentType != "application/octet-stream" {
		t.Errorf("Content-Type is not application/octet-stream, got: %v", contentType)
	}
	if result.RequestBody != "hello-world" {
		t.Errorf("Request body is not as expected, got: %v", result.RequestBody)
	}
}

func TestPostWithRawRequestBody(t *testing.T) {
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
		WithResponse(200, "").
		Build()

	result := runCli([]string{"myservice", "upload", "--input", "hello-world"}, context)

	contentType := result.RequestHeader["content-type"]
	if contentType != "application/octet-stream" {
		t.Errorf("Content-Type is not application/octet-stream, got: %v", contentType)
	}
	if result.RequestBody != "hello-world" {
		t.Errorf("Request body is not as expected, got: %v", result.RequestBody)
	}
}
