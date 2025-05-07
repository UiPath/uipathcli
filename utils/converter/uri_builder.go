package converter

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// UriBuilder takes an Uri and formats it with parameter values.
//
// The builder supports replacing path placeholders like organization and tenant:
// https://cloud.uipath.com/{organization}
// with parameter 'organization' and value 'my-org'
// --> https://cloud.uipath.com/my-org
//
// The builder also supports adding query strings to the uri:
// https://cloud.uipath.com/users
// with parameter 'firstName' and value 'Thomas'
// and parameter 'lastName' and value 'Schmitt'
// --> https://cloud.uipath.com/users?firstName=Thomas&lastName=Schmitt
type UriBuilder struct {
	uri                string
	converter          *StringConverter
	queryStringBuilder *QueryStringBuilder
}

func (b *UriBuilder) FormatPath(name string, value interface{}) *UriBuilder {
	valueString := b.formatPathValue(value)
	b.uri = strings.ReplaceAll(b.uri, "{"+name+"}", valueString)
	return b
}

func (b *UriBuilder) formatPathValue(value interface{}) string {
	switch value := value.(type) {
	case []int, []float64, []bool, []string:
		array := b.converter.ToStringArray(value)
		return b.toCommaSeparatedStringPathEscape(array)
	default:
		str := b.converter.ToString(value)
		return url.PathEscape(str)
	}
}

func (b *UriBuilder) toCommaSeparatedStringPathEscape(array []string) string {
	result := make([]string, len(array))
	for i, value := range array {
		result[i] = url.PathEscape(value)
	}
	return strings.Join(result, ",")
}

func (b *UriBuilder) AddQueryString(name string, value interface{}) *UriBuilder {
	b.queryStringBuilder.Add(name, value)
	return b
}

func (b *UriBuilder) Build() string {
	queryString := b.queryStringBuilder.Build()
	if queryString == "" {
		return b.uri
	}
	return b.uri + "?" + queryString
}

func NewUriBuilder(baseUri url.URL, route string) *UriBuilder {
	normalizedPath := strings.Trim(baseUri.Path, "/")
	normalizedRoute := strings.Trim(route, "/")
	path := path.Join(normalizedPath, normalizedRoute)
	uri := fmt.Sprintf("%s://%s/%s", baseUri.Scheme, baseUri.Host, path)
	return &UriBuilder{uri, NewStringConverter(), NewQueryStringBuilder()}
}
