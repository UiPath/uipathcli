package commandline

import (
	"testing"

	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/utils"
)

func TestConvertReturnsErrorForInvalidBoolean(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("enabled", parser.ParameterTypeBoolean, []parser.Parameter{})
	_, err := converter.Convert("invalid", parameter)

	if err.Error() != "Cannot convert 'enabled' value 'invalid' to boolean" {
		t.Errorf("Should return error, but got: %v", err)
	}
}

func TestConvertStringToBoolean(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("enabled", parser.ParameterTypeBoolean, []parser.Parameter{})
	result, err := converter.Convert("true", parameter)

	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	if result != true {
		t.Errorf("Result should be boolean, but got: %v", result)
	}
}

func TestConvertStringToFileStream(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("file", parser.ParameterTypeBinary, []parser.Parameter{})
	result, err := converter.Convert("test.txt", parameter)

	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	fileStream := result.(*utils.FileStream)
	if fileStream.Name() != "test.txt" {
		t.Errorf("Result should be file stream, but got: %v", result)
	}
}

func TestConvertStringToIntegerArray(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("metrics", parser.ParameterTypeIntegerArray, []parser.Parameter{})
	result, err := converter.Convert("5,2", parameter)

	if err != nil {
		t.Errorf("Should not return error, but got: %v", err)
	}
	array := result.([]int)
	if len(array) != 2 || array[0] != 5 || array[1] != 2 {
		t.Errorf("Result should be integer array, but got: %v", result)
	}
}

func TestConvertExpressionToObject(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("tag", parser.ParameterTypeObject,
		[]parser.Parameter{
			newParameter("name", parser.ParameterTypeString, []parser.Parameter{}),
			newParameter("value", parser.ParameterTypeNumber, []parser.Parameter{}),
		})
	result, _ := converter.Convert("name=hello;value=1.5", parameter)

	name := getValue(result, "name")
	if name != "hello" {
		t.Errorf("Result should be string, but got: %v", name)
	}
	value := getValue(result, "value")
	if value != 1.5 {
		t.Errorf("Result should be float, but got: %v", value)
	}
}

func TestConvertNestedExpressionToObject(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("user", parser.ParameterTypeObject,
		[]parser.Parameter{
			newParameter("profile", parser.ParameterTypeObject,
				[]parser.Parameter{
					newParameter("status", parser.ParameterTypeInteger, []parser.Parameter{}),
				},
			),
		})
	result, _ := converter.Convert("profile.status=1", parameter)

	profile := getValue(result, "profile")
	status := getValue(profile, "status")
	if status != 1 {
		t.Errorf("Result should be integer, but got: %v", status)
	}
}

func TestCustomParameterAddedToObject(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("request", parser.ParameterTypeObject, []parser.Parameter{})
	result, _ := converter.Convert("firstName=Thomas", parameter)

	firstName := getValue(result, "firstName")
	if firstName != "Thomas" {
		t.Errorf("Custom property not found, got: %v", firstName)
	}
}

func TestConvertObjectArray(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("root", parser.ParameterTypeObjectArray, []parser.Parameter{})
	result, _ := converter.Convert("nodes[0].id = 1; nodes[0].value = my-value;", parameter)

	firstNode := getArrayValue(result, "nodes", 0)
	id := getValue(firstNode, "id")
	if id != "1" {
		t.Errorf("Could not find first node id, got: %v", id)
	}
	value := getValue(firstNode, "value")
	if value != "my-value" {
		t.Errorf("Could not find first node value, got: %v", value)
	}
}

func TestConvertRootArray(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("root", parser.ParameterTypeObjectArray, []parser.Parameter{})
	result, _ := converter.Convert("[0].id = 1; [0].value = my-value;", parameter)

	array := result.([]interface{})
	id := getValue(array[0], "id")
	if id != "1" {
		t.Errorf("Could not find first node id, got: %v", id)
	}
	value := getValue(array[0], "value")
	if value != "my-value" {
		t.Errorf("Could not find first node value, got: %v", value)
	}
}

func TestMixingRootObjectAndArrayReturnsError(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("root", parser.ParameterTypeObjectArray, []parser.Parameter{})
	_, err := converter.Convert("[0].id = 1; value = my-value;", parameter)

	if err.Error() != "Cannot convert 'root' value because there is a type mismatch" {
		t.Errorf("Should return error, but got: %v", err)
	}
}

func TestMixingObjectAndArrayReturnsError(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("root", parser.ParameterTypeObjectArray, []parser.Parameter{})
	_, err := converter.Convert("nodes[0].id = 1; nodes.value = my-value;", parameter)

	if err.Error() != "Cannot convert 'root' value because there is a type mismatch for the object key 'nodes'" {
		t.Errorf("Should return error, but got: %v", err)
	}
}

func TestInvalidIndexIsIgnored(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("root", parser.ParameterTypeObjectArray, []parser.Parameter{})
	result, _ := converter.Convert("[invalid].id = 1", parameter)

	value := getValue(result, "[invalid]")
	id := getValue(value, "id")
	if id != "1" {
		t.Errorf("Could not find id value, got: %v", id)
	}
}

func TestNegativeIndexIsIgnored(t *testing.T) {
	converter := newTypeConverter()

	parameter := newParameter("root", parser.ParameterTypeObjectArray, []parser.Parameter{})
	result, _ := converter.Convert("[-1].id = 1", parameter)

	value := getValue(result, "[-1]")
	id := getValue(value, "id")
	if id != "1" {
		t.Errorf("Could not find id value, got: %v", id)
	}
}

func getValue(result interface{}, key string) interface{} {
	return result.(map[string]interface{})[key]
}

func getArrayValue(result interface{}, key string, index int) interface{} {
	return getValue(result, key).([]interface{})[index]
}

func newParameter(name string, t string, parameters []parser.Parameter) parser.Parameter {
	return *parser.NewParameter(name, t, "", "", name, false, nil, []interface{}{}, parameters)
}
