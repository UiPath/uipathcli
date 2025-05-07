// Package converter provides helper classes for transformating data
// from one representation to another.
package converter

import (
	"fmt"
	"strconv"
	"strings"
)

// StringConverter converts interface{} values into a string.
// Depending on the type of the parameter, it converts the value to
// the proper format:
// - Integers, Float, etc.. are simply converted to a string
// - Arrays are formatted comma-separated
// - Booleans are converted to true or false
type StringConverter struct{}

func (c StringConverter) ToString(value interface{}) string {
	switch value := value.(type) {
	case []int, []float64, []bool, []string:
		array := c.ToStringArray(value)
		return strings.Join(array, ",")
	case int:
		return strconv.Itoa(value)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case string:
		return value
	default:
		return fmt.Sprint(value)
	}
}

func (c StringConverter) ToStringArray(value interface{}) []string {
	switch value := value.(type) {
	case []int:
		return c.intArrayToStringArray(value)
	case []float64:
		return c.floatArrayToStringArray(value)
	case []bool:
		return c.boolArrayToStringArray(value)
	case []string:
		return value
	default:
		return strings.Fields(fmt.Sprint(value))
	}
}

func (c StringConverter) intArrayToStringArray(array []int) []string {
	result := make([]string, len(array))
	for i, value := range array {
		result[i] = strconv.Itoa(value)
	}
	return result
}

func (c StringConverter) floatArrayToStringArray(array []float64) []string {
	result := make([]string, len(array))
	for i, value := range array {
		result[i] = strconv.FormatFloat(value, 'f', -1, 64)
	}
	return result
}

func (c StringConverter) boolArrayToStringArray(array []bool) []string {
	result := make([]string, len(array))
	for i, value := range array {
		result[i] = strconv.FormatBool(value)
	}
	return result
}

func NewStringConverter() *StringConverter {
	return &StringConverter{}
}
