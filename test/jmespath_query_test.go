package test

import (
	"net/http"
	"testing"
)

func TestQuerySelectsField(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"hello":"world"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--query", "hello"}, context)

	expectedStdOut := "\"world\"\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestQuerySelectsMultipleFields(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"myfield":"foo","otherfield":"bar"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--query", "{a:myfield, b:otherfield}"}, context)

	expectedStdOut := `{
  "a": "foo",
  "b": "bar"
}
`
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestInvalidQueryReturnsError(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--query", ";"}, context)

	expectedStdErr := "Error in query: SyntaxError: Unknown char: ';'\n"
	if result.StdErr != expectedStdErr {
		t.Errorf("Expected query error on stdout '%v', got: '%v'", expectedStdErr, result.StdErr)
	}
}

func TestQuerySelectsArrayFields(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"results":[{"a": 1}, {"a": 2}]}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--query", "results[].a"}, context)

	expectedStdOut := `[
  1,
  2
]
`
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}
