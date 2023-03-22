package commandline

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/utils"
)

// The typeConverter converts the string value from the command-line argument into the type
// the definition declared. CLI arguments are always passed as strings and need to be converted
// to their respective type.
type typeConverter struct{}

func (c typeConverter) trim(value string) string {
	return strings.TrimSpace(value)
}

func (c typeConverter) convertToInteger(value string, parameter parser.Parameter) (int, error) {
	result, err := strconv.Atoi(c.trim(value))
	if err != nil {
		return 0, fmt.Errorf("Cannot convert '%s' value '%s' to integer", parameter.Name, value)
	}
	return result, nil
}

func (c typeConverter) convertToNumber(value string, parameter parser.Parameter) (float64, error) {
	result, err := strconv.ParseFloat(c.trim(value), 64)
	if err != nil {
		return 0, fmt.Errorf("Cannot convert '%s' value '%s' to number", parameter.Name, value)
	}
	return result, nil
}

func (c typeConverter) convertToBoolean(value string, parameter parser.Parameter) (bool, error) {
	trimmedValue := c.trim(value)
	if strings.EqualFold(trimmedValue, "true") {
		return true, nil
	} else if strings.EqualFold(trimmedValue, "false") {
		return false, nil
	}
	return false, fmt.Errorf("Cannot convert '%s' value '%s' to boolean", parameter.Name, value)
}

func (c typeConverter) convertToBinary(value string, parameter parser.Parameter) (utils.Stream, error) {
	return utils.NewFileStream(value), nil
}

func (c typeConverter) findParameter(parameter *parser.Parameter, name string) *parser.Parameter {
	if parameter == nil {
		return nil
	}
	for _, param := range parameter.Parameters {
		if param.FieldName == name {
			return &param
		}
	}
	return nil
}

func (c typeConverter) tryConvert(value string, parameter *parser.Parameter) (interface{}, error) {
	if parameter == nil {
		return value, nil
	}
	return c.Convert(value, *parameter)
}

func (c typeConverter) assignToObject(obj map[string]interface{}, keys []string, value string, parameter parser.Parameter) error {
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

func (c typeConverter) convertToObject(value string, parameter parser.Parameter) (interface{}, error) {
	obj := map[string]interface{}{}
	assigns := c.splitEscaped(value, ';')
	for _, assign := range assigns {
		keyValue := c.splitEscaped(assign, '=')
		if len(keyValue) < 2 {
			keyValue = append(keyValue, "")
		}
		keys := c.splitEscaped(strings.Trim(keyValue[0], " "), '.')
		value := strings.Trim(keyValue[1], " ")
		err := c.assignToObject(obj, keys, value, parameter)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func (c typeConverter) convertToStringArray(value string, parameter parser.Parameter) ([]string, error) {
	return c.splitEscaped(value, ','), nil
}

func (c typeConverter) convertToIntegerArray(value string, parameter parser.Parameter) ([]int, error) {
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

func (c typeConverter) convertToNumberArray(value string, parameter parser.Parameter) ([]float64, error) {
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

func (c typeConverter) convertToBooleanArray(value string, parameter parser.Parameter) ([]bool, error) {
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

func (c typeConverter) convertToObjectArray(value string, parameter parser.Parameter) ([]interface{}, error) {
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

func (c typeConverter) splitEscaped(str string, separator byte) []string {
	result := []string{}
	item := []byte{}
	escaping := false
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

func (c typeConverter) Convert(value string, parameter parser.Parameter) (interface{}, error) {
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

func newTypeConverter() *typeConverter {
	return &typeConverter{}
}
