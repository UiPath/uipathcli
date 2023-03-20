package test

import (
	"testing"
)

func TestWaitNonBooleanExpressionReturnsError(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, `{"version":1}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--wait", "version"}, context)

	if result.Error.Error() != "Error in wait condition: JMESPath expression needs to return boolean" {
		t.Errorf("Expected error that JMESPath needs to return boolean, but got: %v", result.Error)
	}
}
func TestWaitInvalidExpressionReturnsError(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, `{"version":1}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--wait", "invalid 2& expression"}, context)

	if result.Error.Error() != "Error in query: SyntaxError: Unexpected token at the end of the expression: tNumber" {
		t.Errorf("Expected error for invalid query, but got: %v", result.Error)
	}
}

func TestWaitExpressionIsTrue(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, `{"version":1}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--wait", "version == `1`"}, context)

	if result.Error != nil {
		t.Errorf("Unexpected error, got: %v", result.Error)
	}
	expectedOutput := `{
  "version": 1
}
`
	if result.StdOut != expectedOutput {
		t.Errorf("Expected response body, but got: %v", result.StdOut)
	}
}

func TestWaitForExpressionToBeTrue(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithNextResponse(200, `{"version":1}`).
		WithNextResponse(200, `{"version":1}`).
		WithResponse(200, `{"version":2}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--wait", "version == `2`"}, context)

	if result.Error != nil {
		t.Errorf("Unexpected error, got: %v", result.Error)
	}
	expectedOutput := `{
  "version": 2
}
`
	if result.StdOut != expectedOutput {
		t.Errorf("Expected response body, but got: %v", result.StdOut)
	}
}

func TestWaitForExpressionPrintsStatusMessage(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithNextResponse(200, `{"version":1}`).
		WithNextResponse(200, `{"version":1}`).
		WithResponse(200, `{"version":2}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--wait", "version == `2`"}, context)

	if result.Error != nil {
		t.Errorf("Unexpected error, got: %v", result.Error)
	}
	expected := `Condition is not met yet. Waiting...
Condition is not met yet. Waiting...
`
	if result.StdErr != expected {
		t.Errorf("Expected status message on standard error, but got: %v", result.StdErr)
	}
}

func TestWaitForExpressionTimesOut(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: ping
      summary: Simple ping
`

	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		WithResponse(200, `{"version":1}`).
		Build()

	result := RunCli([]string{"myservice", "ping", "--wait", "version == `2`", "--wait-timeout", "3"}, context)

	if result.Error.Error() != "Timed out waiting for condition" {
		t.Errorf("Expected timeout error, but got: %v", result.Error)
	}
}
