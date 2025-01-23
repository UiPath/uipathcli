package commandline

import (
	"testing"
)

func TestParsesArgOnly(t *testing.T) {
	flags := NewFlagBuilder().
		AddHelpFlag().
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "--help"}, flags)

	value := parser.GetValue("help")
	if value != true {
		t.Errorf("Expected argument value to be true, but got: %v", value)
	}
}

func TestParsesServiceName(t *testing.T) {
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation"}, []*FlagDefinition{})

	serviceName := parser.GetServiceName()
	if serviceName != "my-service" {
		t.Errorf("Expected service name to be my-service, but got: %v", serviceName)
	}
}

func TestParsesResourceName(t *testing.T) {
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation"}, []*FlagDefinition{})

	resourceName := parser.GetResourceName()
	if resourceName != "my-resource" {
		t.Errorf("Expected resource name to be my-resource, but got: %v", resourceName)
	}
}

func TestParsesOperationName(t *testing.T) {
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation"}, []*FlagDefinition{})

	operationName := parser.GetOperationName()
	if operationName != "my-operation" {
		t.Errorf("Expected operation name to be my-operation, but got: %v", operationName)
	}
}

func TestParsesSimpleArgument(t *testing.T) {
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeString)).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--my-arg", "1"}, flags)

	value := parser.GetValue("my-arg")
	if value != "1" {
		t.Errorf("Expected argument value to be 1, but got: %v", value)
	}
}

func TestParsesMultipleArguments(t *testing.T) {
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeString)).
		AddFlag(NewFlag("my-other-arg", "My other argument", FlagTypeString)).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--my-arg", "1", "--my-other-arg", "value2"}, flags)

	value := parser.GetValue("my-arg")
	if value != "1" {
		t.Errorf("Expected argument value to be 1, but got: %v", value)
	}
	value2 := parser.GetValue("my-other-arg")
	if value2 != "value2" {
		t.Errorf("Expected argument value to be value2, but got: %v", value2)
	}
}

func TestParsesArgumentsWithEnvVariableValue(t *testing.T) {
	t.Setenv("UIPATH_MY_ARG", "my-env-value")
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeString).WithEnvVarName("UIPATH_MY_ARG").WithDefaultValue("my-default-value")).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation"}, flags)

	value := parser.GetValue("my-arg")
	if value != "my-env-value" {
		t.Errorf("Expected argument to have env variable value, but got: %v", value)
	}
}

func TestParsesArgumentsWithDefaultValue(t *testing.T) {
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeString).WithDefaultValue("my-default-value")).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation"}, flags)

	value := parser.GetValue("my-arg")
	if value != "my-default-value" {
		t.Errorf("Expected argument to have default value, but got: %v", value)
	}
}

func TestParsesArgumentsWithoutValueAsBoolean(t *testing.T) {
	flags := NewFlagBuilder().
		AddDefaultFlags(false).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--debug"}, flags)

	value := parser.GetValue("debug")
	if value != true {
		t.Errorf("Expected argument value to be true, but got: %v", value)
	}
}

func TestParsesArgumentsWithAndWithoutValues(t *testing.T) {
	flags := NewFlagBuilder().
		AddDefaultFlags(false).
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeInteger)).
		AddFlag(NewFlag("my-other-arg", "My other argument", FlagTypeString)).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--my-arg", "1", "--debug", "--my-other-arg", "value2"}, flags)

	value := parser.GetValue("my-arg")
	if value != 1 {
		t.Errorf("Expected argument value to be 1, but got: %v", value)
	}
	value2 := parser.GetValue("my-other-arg")
	if value2 != "value2" {
		t.Errorf("Expected argument value to be value2, but got: %v", value2)
	}
	valueDebug := parser.GetValue("debug")
	if valueDebug != true {
		t.Errorf("Expected argument value to be true, but got: %v", valueDebug)
	}
}

func TestInvalidArgumentReturnsError(t *testing.T) {
	_, err := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "invalid"}, []*FlagDefinition{})

	if err.Error() != "Unknown argument 'invalid'" {
		t.Errorf("Expected error, but got: %v", err)
	}
}

func TestInvalidMultipleArgumentValuesReturnsError(t *testing.T) {
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeString)).
		Build()
	_, err := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--my-arg", "1", "value2"}, flags)

	if err.Error() != "Unknown argument 'value2'" {
		t.Errorf("Expected error, but got: %v", err)
	}
}

func TestInvalidValueTypeReturnsError(t *testing.T) {
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeInteger)).
		Build()
	_, err := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--my-arg", "my-value"}, flags)

	if err.Error() != "Cannot convert 'my-arg' value 'my-value' to integer" {
		t.Errorf("Expected error, but got: %v", err)
	}
}

func TestIgnoresAdditionalArguments(t *testing.T) {
	flags := NewFlagBuilder().
		AddFlag(NewFlag("my-arg", "My argument", FlagTypeInteger).WithDefaultValue(1)).
		Build()
	parser, _ := NewArgsParser([]string{"uipath", "my-service", "my-resource", "my-operation", "--my-additional-arg", "my-value"}, flags)

	value := parser.GetValue("my-additional-arg")
	if value != nil {
		t.Errorf("Expected argument value to be nil, but got: %v", value)
	}
}
