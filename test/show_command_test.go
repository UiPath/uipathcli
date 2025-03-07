package test

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestCommandReturnedSuccessfully(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
      tags:
        - health
      parameters:
        - name: MyParam
          in: query
          schema:
            type: string
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"commands", "show"}, context)

	command := GetCommand(t, result)
	name := command["name"]
	if name != "uipath" {
		t.Errorf("Unexpected executable name in output, got: %v", name)
	}

	serviceCommand := GetSubcommands(command)[0]
	serviceName := serviceCommand["name"]
	if serviceName != "myservice" {
		t.Errorf("Unexpected service name in output, got: %v", serviceName)
	}

	categoryCommand := GetSubcommands(serviceCommand)[0]
	categoryName := categoryCommand["name"]
	if categoryName != "health" {
		t.Errorf("Unexpected category name in output, got: %v", categoryName)
	}

	operationCommand := GetSubcommands(categoryCommand)[0]
	operationName := operationCommand["name"]
	if operationName != "ping" {
		t.Errorf("Unexpected operation name in output, got: %v", operationName)
	}

	parameters := GetParameters(operationCommand)
	if len(parameters) != 1 {
		t.Errorf("Expected single parameter, got: %v", len(parameters))
	}
	parameterName := parameters[0]["name"]
	if parameterName != "my-param" {
		t.Errorf("Expected my-param parameter, got: %v", parameterName)
	}
}

func TestCommandWithoutCategoryReturnedSuccessfully(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
      parameters:
        - name: MyParam
          in: query
          schema:
            type: number
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"commands", "show"}, context)

	command := GetCommand(t, result)
	serviceCommand := GetSubcommands(command)[0]
	operationCommand := GetSubcommands(serviceCommand)[0]
	operationName := operationCommand["name"]
	if operationName != "ping" {
		t.Errorf("Unexpected operation name in output, got: %v", operationName)
	}
}

func TestMultipleCommandsReturnedAlphabetically(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      operationId: get
    post:
      operationId: create
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"commands", "show"}, context)

	command := GetCommand(t, result)
	serviceCommand := GetSubcommands(command)[0]
	operationCommands := GetSubcommands(serviceCommand)
	operationName := operationCommands[0]["name"]
	if operationName != "create" {
		t.Errorf("Expected operation create first, got: %v", operationName)
	}
	operationName2 := operationCommands[1]["name"]
	if operationName2 != "get" {
		t.Errorf("Expected operation get second, got: %v", operationName2)
	}
}

func TestCommandGlobalFlags(t *testing.T) {
	definition := `
paths:
  /ping:
    get:
      summary: Simple ping
      operationId: ping
      tags:
        - health
`
	context := NewContextBuilder().
		WithDefinition("myservice", definition).
		Build()

	result := RunCli([]string{"commands", "show"}, context)

	command := GetCommand(t, result)
	parameters := GetParameters(command)

	names := []string{}
	for _, parameter := range parameters {
		names = append(names, parameter["name"].(string))
	}

	expectedNames := []string{
		"debug",
		"profile",
		"uri",
		"organization",
		"tenant",
		"insecure",
		"call-timeout",
		"max-attempts",
		"output",
		"query",
		"wait",
		"wait-timeout",
		"file",
		"identity-uri",
		"service-version",
		"help"}
	if !reflect.DeepEqual(names, expectedNames) {
		t.Errorf("Unexpected global parameters in output, expected: %v but got: %v", expectedNames, names)
	}
}

func GetCommand(t *testing.T, result Result) map[string]interface{} {
	command := map[string]interface{}{}
	err := json.Unmarshal([]byte(result.StdOut), &command)
	if err != nil {
		t.Errorf("Failed to deserialize show commands result %v", err)
	}
	return command
}

func GetSubcommands(command map[string]interface{}) []map[string]interface{} {
	return GetArray(command, "subcommands")
}

func GetParameters(command map[string]interface{}) []map[string]interface{} {
	return GetArray(command, "parameters")
}

func GetArray(section map[string]interface{}, name string) []map[string]interface{} {
	array := section[name].([]interface{})
	result := []map[string]interface{}{}
	for _, item := range array {
		result = append(result, item.(map[string]interface{}))
	}
	return result
}
