package test

import (
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
		WithResponse(200, `{"a":"foo","b":"bar"}`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `{"b":"bar","a":"foo","c":"baz"}`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `[{"a":"foo1","b":"bar1"}, {"a":"foo2","b":"bar2"}]`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `[1, 2, 3]`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `[{"a":"foo1","b":"bar1"}, {"b":"bar2"}]`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `{"outer":{"inner":"value"}}`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `{"a":"my-value","outer":{"inner":"value"},"z":"my-last-value"}`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `{"a":null,"z":"my-last-value"}`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

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
		WithResponse(200, `{"results":[{"id":1,"name":"test"},{"id":2,"name":"test2"}]}`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text", "--query", "results[]"}, context)

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
		WithResponse(200, `[[1, "my-name"], [2, "other-name"]]`).
		Build()

	result := runCli([]string{"myservice", "ping", "--output", "text"}, context)

	expectedStdOut := "1\tmy-name\n2\tother-name\n"
	if result.StdOut != expectedStdOut {
		t.Errorf("Expected response body on stdout %v, got: %v", expectedStdOut, result.StdOut)
	}
}
