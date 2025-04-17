package test

import (
	"net/http"
	"testing"
)

func TestTextOutputPrintsFieldsTabSeparated(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"a":"foo","b":"bar"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "foo\tbar\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputPrintsFieldsSortedAndTabSeparated(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"b":"bar","a":"foo","c":"baz"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "foo\tbar\tbaz\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputPrintsObjectArrayOnMultipleLines(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `[{"a":"foo1","b":"bar1"}, {"a":"foo2","b":"bar2"}]`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "foo1\tbar1\nfoo2\tbar2\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputPrintsArrayOnMultipleLines(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `[1, 2, 3]`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "1\n2\n3\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputPrintsObjectAndSkipsMissingField(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `[{"a":"foo1","b":"bar1"}, {"b":"bar2"}]`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "foo1\tbar1\n\tbar2\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputIgnoresComplexObjects(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"outer":{"inner":"value"}}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputPrintsAllButComplexObjects(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"a":"my-value","outer":{"inner":"value"},"z":"my-last-value"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "my-value\tmy-last-value\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputIgnoresNullValue(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"a":null,"z":"my-last-value"}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "my-last-value\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputSupportsQuery(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"results":[{"id":1,"name":"test"},{"id":2,"name":"test2"}]}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text", "--query", "results[]"}, context)

	expectedStdOut := "1\ttest\n2\ttest2\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextSupportsNestedArrays(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `[[1, "my-name"], [2, "other-name"]]`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "1\tmy-name\n2\tother-name\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputFormatsIntegers(t *testing.T) {
	definition := `
paths:
  /resource:
    get:
      operationId: list
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"id":2600000584}`).
		Build()

	result := RunCli([]string{"myservice", "list", "--output", "text"}, context)

	expectedStdOut := "2600000584\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected integer on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}

func TestTextOutputFormatsBooleans(t *testing.T) {
	definition := `
paths:
  /resource:
    get:
      operationId: list
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(http.StatusOK, `{"active":true}`).
		Build()

	result := RunCli([]string{"myservice", "list", "--output", "text"}, context)

	expectedStdOut := "true\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected boolean on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}
