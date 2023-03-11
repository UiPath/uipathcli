package executor

import (
	"fmt"
	"net/url"
	"strings"
)

// queryStringFormatter converts ExecutionParameter's into a query string.
//
// Depending on the type of the parameter, the formatter converts the value to
// the proper format and makes sure the query string is properly escaped.
//
// Example:
// - parameter 'username' and value 'tschmitt'
// - parameter 'message' and value 'Hello World'
// --> username=tschmitt&message=Hello+World
type queryStringFormatter struct{}

func (f queryStringFormatter) Format(parameters []ExecutionParameter) string {
	result := ""
	for _, parameter := range parameters {
		param := f.formatQueryStringParam(parameter)
		if result == "" {
			result = param
		} else {
			result = result + "&" + param
		}
	}
	return result
}

func (f queryStringFormatter) formatQueryStringParam(parameter ExecutionParameter) string {
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

func (f queryStringFormatter) integerArrayToQueryString(key string, value []int) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f queryStringFormatter) numberArrayToQueryString(key string, value []float64) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f queryStringFormatter) booleanArrayToQueryString(key string, value []bool) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f queryStringFormatter) stringArrayToQueryString(key string, value []string) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f queryStringFormatter) arrayToQueryString(key string, value []interface{}) string {
	result := make([]string, len(value))
	for i, v := range value {
		result[i] = f.toQueryString(key, v)
	}
	return strings.Join(result, "&")
}

func (f queryStringFormatter) toQueryString(key string, value interface{}) string {
	stringValue := fmt.Sprintf("%v", value)
	return fmt.Sprintf("%s=%v", key, url.QueryEscape(stringValue))
}

func newQueryStringFormatter() *queryStringFormatter {
	return &queryStringFormatter{}
}
