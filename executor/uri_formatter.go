package executor

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// The UriFormatter takes an Uri and formats it with ExecutionParameter values.
//
// The formatter supports replacing path placeholders like organization and tenant:
// https://cloud.uipath.com/{organization}
// with parameter 'organization' and value 'my-org'
// --> https://cloud.uipath.com/my-org
//
// The formatter also supports adding query strings to the uri:
// https://cloud.uipath.com/users
// with parameter 'firstName' and value 'Thomas'
// and parameter 'lastName' and value 'Schmitt'
// --> https://cloud.uipath.com/users?firstName=Thomas&lastName=Schmitt
type UriFormatter struct {
	uri         string
	queryString string
}

func (f *UriFormatter) FormatPath(parameter ExecutionParameter) {
	formatter := NewParameterFormatter()
	value := formatter.Format(parameter)
	f.uri = strings.ReplaceAll(f.uri, "{"+parameter.Name+"}", value)
}

func (f *UriFormatter) AddQueryString(parameters []ExecutionParameter) {
	formatter := NewQueryStringFormatter()
	f.queryString = formatter.Format(parameters)
}

func (f UriFormatter) Uri() string {
	if f.queryString == "" {
		return f.uri
	}
	return f.uri + "?" + f.queryString
}

func NewUriFormatter(baseUri url.URL, route string) *UriFormatter {
	normalizedPath := strings.Trim(baseUri.Path, "/")
	normalizedRoute := strings.Trim(route, "/")
	path := path.Join(normalizedPath, normalizedRoute)
	uri := fmt.Sprintf("%s://%s/%s", baseUri.Scheme, baseUri.Host, path)
	return &UriFormatter{uri, ""}
}
