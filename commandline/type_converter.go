package commandline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
)

const filePrefix = "file://"

type TypeConverter struct{}

func (c TypeConverter) trim(value string) string {
	return strings.TrimSpace(value)
}

func (c TypeConverter) convertToInteger(value string, parameter parser.Parameter) (int, error) {
	result, err := strconv.Atoi(c.trim(value))
	if err != nil {
		return 0, fmt.Errorf("Cannot convert '%s' value '%s' to integer", parameter.Name, value)
	}
	return result, nil
}

func (c TypeConverter) convertToNumber(value string, parameter parser.Parameter) (float64, error) {
	result, err := strconv.ParseFloat(c.trim(value), 64)
	if err != nil {
		return 0, fmt.Errorf("Cannot convert '%s' value '%s' to number", parameter.Name, value)
	}
	return result, nil
}

func (c TypeConverter) convertToBoolean(value string, parameter parser.Parameter) (bool, error) {
	trimmedValue := c.trim(value)
	if strings.EqualFold(trimmedValue, "true") {
		return true, nil
	} else if strings.EqualFold(trimmedValue, "false") {
		return false, nil
	}
	return false, fmt.Errorf("Cannot convert '%s' value '%s' to boolean", parameter.Name, value)
}

func (c TypeConverter) readFile(path string) (executor.FileReference, error) {
	data, err := os.ReadFile(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return executor.FileReference{}, fmt.Errorf("File '%s' not found", path)
	}
	if err != nil {
		return executor.FileReference{}, fmt.Errorf("Error reading file '%s': %v", path, err)
	}
	filename := filepath.Base(path)
	return *executor.NewFileReference(filename, data), nil
}

func (c TypeConverter) convertToBinary(value string, parameter parser.Parameter) (executor.FileReference, error) {
	if strings.HasPrefix(value, filePrefix) {
		path := strings.TrimPrefix(value, filePrefix)
		return c.readFile(path)
	}
	return *executor.NewFileReference(parameter.Name, []byte(value)), nil
}

func (c TypeConverter) findParameter(parameter *parser.Parameter, name string) *parser.Parameter {
	if parameter == nil {
		return nil
	}
	for _, param := range parameter.Parameters {
		if param.Name == name {
			return &param
		}
	}
	return nil
}

func (c TypeConverter) tryConvert(value string, parameter *parser.Parameter) (interface{}, error) {
	if parameter == nil {
		return value, nil
	}
	return c.Convert(value, *parameter)
}

func (c TypeConverter) assignToObject(obj map[string]interface{}, keys []string, value string, parameter parser.Parameter) error {
	current := obj
	currentParameter := &parameter
	for i := 0; i < len(keys); i++ {
		key := keys[i]
		lastKey := i == len(keys)-1
		if lastKey && current[key] != nil {
			return fmt.Errorf("Cannot convert '%s' value because object key '%s' is already defined", parameter.Name, key)
		}

		currentParameter = c.findParameter(currentParameter, key)
		if current[key] == nil {
			current[key] = map[string]interface{}{}
		}
		if lastKey {
			parsedValue, err := c.tryConvert(value, currentParameter)
			if err != nil {
				return err
			}
			current[key] = parsedValue
			break
		}
		current = current[key].(map[string]interface{})
	}
	return nil
}

func (c TypeConverter) convertToObject(value string, parameter parser.Parameter) (interface{}, error) {
	obj := map[string]interface{}{}
	assigns := c.splitEscaped(value, ';')
	for _, assign := range assigns {
		keyValue := c.splitEscaped(assign, '=')
		if len(keyValue) < 2 {
			keyValue = append(keyValue, "")
		}
		keys := c.splitEscaped(keyValue[0], '.')
		value := keyValue[1]
		err := c.assignToObject(obj, keys, value, parameter)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func (c TypeConverter) convertToStringArray(value string, parameter parser.Parameter) ([]string, error) {
	return c.splitEscaped(value, ','), nil
}

func (c TypeConverter) convertToIntegerArray(value string, parameter parser.Parameter) ([]int, error) {
	splitted := c.splitEscaped(value, ',')

	result := []int{}
	for _, itemStr := range splitted {
		item, err := c.convertToInteger(itemStr, parameter)
		if err != nil {
			return nil, fmt.Errorf("Cannot convert '%s' values '%s' to integer array", parameter.Name, value)
		}
		result = append(result, item)
	}
	return result, nil
}

func (c TypeConverter) convertToNumberArray(value string, parameter parser.Parameter) ([]float64, error) {
	splitted := c.splitEscaped(value, ',')

	result := []float64{}
	for _, itemStr := range splitted {
		item, err := c.convertToNumber(itemStr, parameter)
		if err != nil {
			return nil, fmt.Errorf("Cannot convert '%s' values '%s' to number array", parameter.Name, value)
		}
		result = append(result, item)
	}
	return result, nil
}

func (c TypeConverter) convertToBooleanArray(value string, parameter parser.Parameter) ([]bool, error) {
	splitted := c.splitEscaped(value, ',')

	result := []bool{}
	for _, itemStr := range splitted {
		item, err := c.convertToBoolean(itemStr, parameter)
		if err != nil {
			return nil, fmt.Errorf("Cannot convert '%s' values '%s' to boolean array", parameter.Name, value)
		}
		result = append(result, item)
	}
	return result, nil
}

func (c TypeConverter) convertToObjectArray(value string, parameter parser.Parameter) ([]interface{}, error) {
	splitted := c.splitEscaped(value, ',')

	result := []interface{}{}
	for _, itemStr := range splitted {
		item, err := c.convertToObject(itemStr, parameter)
		if err != nil {
			return nil, fmt.Errorf("Cannot convert '%s' values '%s' to object array", parameter.Name, value)
		}
		result = append(result, item)
	}
	return result, nil
}

func (c TypeConverter) splitEscaped(str string, separator byte) []string {
	var result []string
	var item []byte
	var escaping = false
	for i := 0; i < len(str); i++ {
		char := str[i]
		if char == '\\' && !escaping {
			escaping = true
		} else if char == '\\' && escaping {
			escaping = false
			item = append(item, char)
		} else if char == separator && !escaping {
			result = append(result, string(item))
			item = []byte{}
		} else {
			escaping = false
			item = append(item, char)
		}
	}
	if len(item) > 0 {
		result = append(result, string(item))
	}
	return result
}

func (c TypeConverter) Convert(value string, parameter parser.Parameter) (interface{}, error) {
	switch parameter.Type {
	case parser.ParameterTypeInteger:
		return c.convertToInteger(value, parameter)
	case parser.ParameterTypeNumber:
		return c.convertToNumber(value, parameter)
	case parser.ParameterTypeBoolean:
		return c.convertToBoolean(value, parameter)
	case parser.ParameterTypeBinary:
		return c.convertToBinary(value, parameter)
	case parser.ParameterTypeObject:
		return c.convertToObject(value, parameter)
	case parser.ParameterTypeStringArray:
		return c.convertToStringArray(value, parameter)
	case parser.ParameterTypeIntegerArray:
		return c.convertToIntegerArray(value, parameter)
	case parser.ParameterTypeNumberArray:
		return c.convertToNumberArray(value, parameter)
	case parser.ParameterTypeBooleanArray:
		return c.convertToBooleanArray(value, parameter)
	case parser.ParameterTypeObjectArray:
		return c.convertToObjectArray(value, parameter)
	default:
		return value, nil
	}
}
