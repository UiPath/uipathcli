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
