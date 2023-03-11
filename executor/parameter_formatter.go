package executor

import (
	"fmt"
	"strings"
)

// parameterFormatter converts ExecutionParameter into a string.
// Depending on the type of the parameter, the formatter converts the value to
// the proper format:
// - Integers, Float, etc.. are simply converted to a string
// - Arrays are formatted comma-separated
// - Booleans are converted to true or false
type parameterFormatter struct{}

func (f parameterFormatter) Format(parameter ExecutionParameter) string {
	return f.formatParameter(parameter)
}

func (f parameterFormatter) formatParameter(parameter ExecutionParameter) string {
	switch value := parameter.Value.(type) {
	case []int:
		return f.arrayToCommaSeparatedString(value)
	case []float64:
		return f.arrayToCommaSeparatedString(value)
	case []bool:
		return f.arrayToCommaSeparatedString(value)
	case []string:
		return f.arrayToCommaSeparatedString(value)
	default:
		return fmt.Sprintf("%v", value)
	}
}

func (f parameterFormatter) arrayToCommaSeparatedString(array interface{}) string {
	switch value := array.(type) {
	case []string:
		return strings.Join(value, ",")
	default:
		return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
	}
}

func newParameterFormatter() *parameterFormatter {
	return &parameterFormatter{}
}
