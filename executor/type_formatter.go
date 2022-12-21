package executor

import (
	"fmt"
	"net/url"
	"strings"
)

type TypeFormatter struct{}

func (f TypeFormatter) FormatPath(parameter ExecutionParameter) string {
	return f.formatParameter(parameter)
}

func (f TypeFormatter) FormatHeader(parameter ExecutionParameter) string {
	return f.formatParameter(parameter)
}

func (f TypeFormatter) formatParameter(parameter ExecutionParameter) string {
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

func (f TypeFormatter) arrayToCommaSeparatedString(array interface{}) string {
	switch value := array.(type) {
	case []string:
		return strings.Join(value, ",")
	default:
		return strings.Trim(strings.Join(strings.Fields(fmt.Sprint(value)), ","), "[]")
	}
}

func (f TypeFormatter) FormatQueryString(parameter ExecutionParameter) string {
	switch value := parameter.Value.(type) {
	case []int:
		return f.integerArrayToQueryString(parameter.Name, value)
	case []float64:
		return f.numberArrayToQueryString(parameter.Name, value)
	case []bool:
		return f.booleanArrayToQueryString(parameter.Name, value)
	case []string:
		return f.stringArrayToQueryString(parameter.Name, value)
	default:
		return f.toQueryString(parameter.Name, value)
	}
}

func (f TypeFormatter) integerArrayToQueryString(key string, value []int) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f TypeFormatter) numberArrayToQueryString(key string, value []float64) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f TypeFormatter) booleanArrayToQueryString(key string, value []bool) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f TypeFormatter) stringArrayToQueryString(key string, value []string) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f TypeFormatter) arrayToQueryString(key string, value []interface{}) string {
	result := make([]string, len(value))
	for i, v := range value {
		result[i] = f.toQueryString(key, v)
	}
	return strings.Join(result, url.QueryEscape(","))
}

func (f TypeFormatter) toQueryString(key string, value interface{}) string {
	stringValue := fmt.Sprintf("%v", value)
	return fmt.Sprintf("%s=%v", key, url.QueryEscape(stringValue))
}
