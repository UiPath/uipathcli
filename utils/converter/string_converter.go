// Package converter provides helper classes for transformating data
// from one representation to another.
package converter

import (
	"fmt"
	"strings"
)

// StringConverter converts interface{} values into a string.
// Depending on the type of the parameter, the formatter converts the value to
// the proper format:
// - Integers, Float, etc.. are simply converted to a string
// - Arrays are formatted comma-separated
// - Booleans are converted to true or false
type StringConverter struct{}

func (c StringConverter) ToString(value interface{}) string {
	return c.formatParameter(value)
}

func (c StringConverter) formatParameter(value interface{}) string {
	switch value := value.(type) {
	case []int:
		return c.arrayToCommaSeparatedString(value)
	case []float64:
		return c.arrayToCommaSeparatedString(value)
	case []bool:
		return c.arrayToCommaSeparatedString(value)
	case []string:
		return c.arrayToCommaSeparatedString(value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func (c StringConverter) arrayToCommaSeparatedString(array interface{}) string {
	switch value := array.(type) {
	case []string:
		return strings.Join(value, ",")
	default:
		return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
	}
}

func NewStringConverter() *StringConverter {
	return &StringConverter{}
}
