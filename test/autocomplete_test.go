package commandline

import (
	"testing"
)

func TestAutocompleteNoMatch(t *testing.T) {
	definition := `
paths:
  /ping/{id}:
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice other"}, context)

	if result.StdOut != "" {
		t.Errorf("Should not return any autocomplete words, got: %v", result.StdOut)
	}
}

func TestAutocompletePrefixMatch(t *testing.T) {
	definition := `
paths:
  /ping/{id}:
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice pi"}, context)

	expectedWords := "ping\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}

func TestAutocompleteContainsMatch(t *testing.T) {
	definition := `
paths:
  /ping/{id}:
    get:
      operationId: ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice in"}, context)

	expectedWords := "ping\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}

func TestAutocompleteMultipleMatches(t *testing.T) {
	definition := `
paths:
  /ping/{id}:
    get:
      operationId: ping
  /other-ping/{id}:
    get:
      operationId: other-ping
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice ping"}, context)

	expectedWords := "ping\nother-ping\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}

func TestAutocompleteOrdersMatchWithPrefixFirst(t *testing.T) {
	definition := `
paths:
  /new-create:
    get:
      operationId: new-create
  /create:
    get:
      operationId: create
  /other:
    get:
      operationId: other
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice create"}, context)

	expectedWords := "create\nnew-create\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}

func TestAutocompleteFlagPrefixMatch(t *testing.T) {
	definition := `
paths:
  /ping/{identifier}:
    get:
      operationId: ping
      parameters:
      - name: identifier
        in: path
        required: true
        description: The identifier
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice ping --id"}, context)

	expectedWords := "--identifier\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}

func TestAutocompleteFlagContainsMatch(t *testing.T) {
	definition := `
paths:
  /ping/{identifier}:
    get:
      operationId: ping
      parameters:
      - name: identifier
        in: path
        required: true
        description: The identifier
        schema:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice ping --enti"}, context)

	expectedWords := "--identifier\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}

func TestAutocompleteFlagMultipleMatches(t *testing.T) {
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
        short-description:
          type: string
        description:
          type: string
        other:
          type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := runCli([]string{"complete", "--command", "uipathcli myservice post-validate --desc"}, context)

	expectedWords := "--description\n--short-description\n"
	if result.StdOut != expectedWords {
		t.Errorf("Did not return the expected autocomplete words, expected: %v, got: %v", expectedWords, result.StdOut)
	}
}