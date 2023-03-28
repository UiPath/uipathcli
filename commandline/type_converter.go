package commandline

import (
	"encoding/json"
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

func (c typeConverter) trimAll(values []string) []string {
	result := []string{}
	for _, value := range values {
		result = append(result, c.trim(value))
	}
	return result
}

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

func (c typeConverter) isArray(key string) (bool, string, int) {
	if len(key) < 3 {
		return false, "", -1
	}
	startIndex := strings.LastIndexAny(key, "[")
	if startIndex == -1 || !strings.HasSuffix(key, "]") {
		return false, "", -1
	}
	property := key[0:startIndex]
	indexString := key[startIndex+1 : len(key)-1]
	index, err := strconv.Atoi(indexString)
	if index < 0 || err != nil {
		return false, "", -1
	}
	return true, property, index
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

		isArray, arrayKey, index := c.isArray(key)
		if isArray {
			array, item := c.initArray(parameter.Name, current[arrayKey], index)
			if array == nil || item == nil {
				return fmt.Errorf("Cannot convert '%s' value because object key '%s' is already defined", parameter.Name, arrayKey)
			}
			current[arrayKey] = array
			current = item
			continue
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
		nextItem, ok := current[key].(map[string]interface{})
		if !ok {
			return fmt.Errorf("Cannot convert '%s' value because there is a type mismatch for the object key '%s'", parameter.Name, key)
		}
		current = nextItem
	}
	return nil
}

func (c typeConverter) initArrayItem(array []interface{}, index int) ([]interface{}, map[string]interface{}) {
	if index > len(array)-1 {
		a := make([]interface{}, index+1)
		copy(a, array)
		array = a
	}
	obj := array[index]
	if obj == nil {
		obj = map[string]interface{}{}
		array[index] = obj
	}
	item, ok := obj.(map[string]interface{})
	if !ok {
		return nil, nil
	}
	return array, item
}

func (c typeConverter) initArray(name string, value interface{}, index int) ([]interface{}, map[string]interface{}) {
	if value == nil {
		value = []interface{}{}
	}
	array, ok := value.([]interface{})
	if !ok {
		return nil, nil
	}
	return c.initArrayItem(array, index)
}

func (c typeConverter) convertJsonToObject(value string) (interface{}, error) {
	var data interface{}
	err := json.Unmarshal([]byte(value), &data)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (c typeConverter) convertToObject(value string, parameter parser.Parameter) (interface{}, error) {
	data, err := c.convertJsonToObject(value)
	if err == nil {
		return data, nil
	}

	obj := map[string]interface{}{}
	assigns := c.splitEscaped(value, ';')
	for _, assign := range assigns {
		keyValue := c.splitEscaped(assign, '=')
		if len(keyValue) < 2 {
			keyValue = append(keyValue, "")
		}
		keys := c.trimAll(c.splitEscaped(keyValue[0], '.'))
		value := c.trim(keyValue[1])
		err := c.assignToObject(obj, keys, value, parameter)
		if err != nil {
			return nil, err
		}
	}
	return obj, nil
}

func (c typeConverter) convertValueToStringArray(value string, parameter parser.Parameter) ([]string, error) {
	return c.splitEscaped(value, ','), nil
}

func (c typeConverter) convertValueToIntegerArray(value string, parameter parser.Parameter) ([]int, error) {
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

func (c typeConverter) convertValueToNumberArray(value string, parameter parser.Parameter) ([]float64, error) {
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

func (c typeConverter) convertValueToBooleanArray(value string, parameter parser.Parameter) ([]bool, error) {
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
		return c.convertValueToStringArray(value, parameter)
	case parser.ParameterTypeIntegerArray:
		return c.convertValueToIntegerArray(value, parameter)
	case parser.ParameterTypeNumberArray:
		return c.convertValueToNumberArray(value, parameter)
	case parser.ParameterTypeBooleanArray:
		return c.convertValueToBooleanArray(value, parameter)
	case parser.ParameterTypeObjectArray:
		return c.convertToObject(value, parameter)
	default:
		return value, nil
	}
}

func (c typeConverter) convertToObjectArray(values []string, parameter parser.Parameter) ([]interface{}, error) {
	result := []interface{}{}
	for _, value := range values {
		item, err := c.convertToObject(value, parameter)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func (c typeConverter) convertToStringArray(values []string, parameter parser.Parameter) ([]string, error) {
	result := []string{}
	for _, value := range values {
		items, err := c.convertValueToStringArray(value, parameter)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (c typeConverter) convertToIntegerArray(values []string, parameter parser.Parameter) ([]int, error) {
	result := []int{}
	for _, value := range values {
		items, err := c.convertValueToIntegerArray(value, parameter)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (c typeConverter) convertToNumberArray(values []string, parameter parser.Parameter) ([]float64, error) {
	result := []float64{}
	for _, value := range values {
		items, err := c.convertValueToNumberArray(value, parameter)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (c typeConverter) convertToBooleanArray(values []string, parameter parser.Parameter) ([]bool, error) {
	result := []bool{}
	for _, value := range values {
		items, err := c.convertValueToBooleanArray(value, parameter)
		if err != nil {
			return nil, err
		}
		result = append(result, items...)
	}
	return result, nil
}

func (c typeConverter) ConvertArray(values []string, parameter parser.Parameter) (interface{}, error) {
	switch parameter.Type {
	case parser.ParameterTypeObjectArray:
		return c.convertToObjectArray(values, parameter)
	case parser.ParameterTypeStringArray:
		return c.convertToStringArray(values, parameter)
	case parser.ParameterTypeIntegerArray:
		return c.convertToIntegerArray(values, parameter)
	case parser.ParameterTypeNumberArray:
		return c.convertToNumberArray(values, parameter)
	case parser.ParameterTypeBooleanArray:
		return c.convertToBooleanArray(values, parameter)
	default:
		return values, nil
	}
}

func newTypeConverter() *typeConverter {
	return &typeConverter{}
}
