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
	valueString := b.converter.ToString(value)
	b.uri = strings.ReplaceAll(b.uri, "{"+name+"}", valueString)
	return b
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

func NewUriBuilder(baseUri string, route string) *UriBuilder {
	uri := strings.Trim(baseUri, "/")
	normalizedRoute := strings.Trim(route, "/")
	if normalizedRoute != "" {
		uri += "/" + normalizedRoute
	}
	return &UriBuilder{uri, NewStringConverter(), NewQueryStringBuilder()}
}

func NewUriBuilderFromUrl(baseUri url.URL, route string) *UriBuilder {
	normalizedPath := strings.Trim(baseUri.Path, "/")
	normalizedRoute := strings.Trim(route, "/")
	path := path.Join(normalizedPath, normalizedRoute)
	uri := fmt.Sprintf("%s://%s/%s", baseUri.Scheme, baseUri.Host, path)
	return &UriBuilder{uri, NewStringConverter(), NewQueryStringBuilder()}
}
