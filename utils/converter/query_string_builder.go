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

func (b *QueryStringBuilder) Add(name string, value interface{}) *QueryStringBuilder {
	param := b.formatQueryStringParam(name, value)
	if b.querystring == "" {
		b.querystring = param
	} else {
		b.querystring = b.querystring + "&" + param
	}
	return b
}

func (b *QueryStringBuilder) Build() string {
	return b.querystring
}

func (b *QueryStringBuilder) formatQueryStringParam(name string, value interface{}) string {
	switch value := value.(type) {
	case []int:
		return b.integerArrayToQueryString(name, value)
	case []float64:
		return b.numberArrayToQueryString(name, value)
	case []bool:
		return b.booleanArrayToQueryString(name, value)
	case []string:
		return b.stringArrayToQueryString(name, value)
	default:
		return b.toQueryString(name, value)
	}
}

func (b *QueryStringBuilder) integerArrayToQueryString(key string, value []int) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return b.arrayToQueryString(key, result)
}

func (b *QueryStringBuilder) numberArrayToQueryString(key string, value []float64) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return b.arrayToQueryString(key, result)
}

func (b *QueryStringBuilder) booleanArrayToQueryString(key string, value []bool) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return b.arrayToQueryString(key, result)
}

func (b *QueryStringBuilder) stringArrayToQueryString(key string, value []string) string {
	result := make([]interface{}, len(value))
	for i, v := range value {
		result[i] = v
	}
	return b.arrayToQueryString(key, result)
}

func (b *QueryStringBuilder) arrayToQueryString(key string, value []interface{}) string {
	result := make([]string, len(value))
	for i, v := range value {
		result[i] = b.toQueryString(key, v)
	}
	return strings.Join(result, "&")
}

func (b *QueryStringBuilder) toQueryString(key string, value interface{}) string {
	stringValue := fmt.Sprint(value)
	return fmt.Sprintf("%s=%v", url.QueryEscape(key), url.QueryEscape(stringValue))
}

func NewQueryStringBuilder() *QueryStringBuilder {
	return &QueryStringBuilder{}
}
