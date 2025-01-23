package converter

import (
	"fmt"
	"net/url"
	"strings"
)

// QueryStringBuilder converts a list of parameters into a query string.
//
// Depending on the type of the parameter, the formatter converts the value to
// the proper format and makes sure the query string is properly escaped.
//
// Example:
// - parameter 'username' and value 'tschmitt'
// - parameter 'message' and value 'Hello World'
// --> username=tschmitt&message=Hello+World
type QueryStringBuilder struct {
	querystring string
}

func (f *QueryStringBuilder) Add(name string, value interface{}) {
	param := f.formatQueryStringParam(name, value)
	if f.querystring == "" {
		f.querystring = param
	} else {
		f.querystring = f.querystring + "&" + param
	}
}

func (f QueryStringBuilder) Build() string {
	return f.querystring
}

func (f QueryStringBuilder) formatQueryStringParam(name string, value interface{}) string {
	switch value := value.(type) {
	case []int:
		return f.integerArrayToQueryString(name, value)
	case []float64:
		return f.numberArrayToQueryString(name, value)
	case []bool:
		return f.booleanArrayToQueryString(name, value)
	case []string:
		return f.stringArrayToQueryString(name, value)
	default:
		return f.toQueryString(name, value)
	}
}

func (f QueryStringBuilder) integerArrayToQueryString(key string, value []int) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f QueryStringBuilder) numberArrayToQueryString(key string, value []float64) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f QueryStringBuilder) booleanArrayToQueryString(key string, value []bool) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f QueryStringBuilder) stringArrayToQueryString(key string, value []string) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return f.arrayToQueryString(key, result)
}

func (f QueryStringBuilder) arrayToQueryString(key string, value []interface{}) string {
	result := make([]string, len(value))
	for i, v := range value {
		result[i] = f.toQueryString(key, v)
	}
	return strings.Join(result, "&")
}

func (f QueryStringBuilder) toQueryString(key string, value interface{}) string {
	stringValue := fmt.Sprintf("%v", value)
	return fmt.Sprintf("%s=%v", key, url.QueryEscape(stringValue))
}

func NewQueryStringBuilder() *QueryStringBuilder {
	return &QueryStringBuilder{}
}
