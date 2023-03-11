package executor

import (
	"fmt"
	"net/url"
	"path"
	"strings"
)

// uriFormatter takes an Uri and formats it with ExecutionParameter values.
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
type uriFormatter struct {
	uri         string
	queryString string
}

func (f *uriFormatter) FormatPath(parameter ExecutionParameter) {
	formatter := newParameterFormatter()
	value := formatter.Format(parameter)
	f.uri = strings.ReplaceAll(f.uri, "{"+parameter.Name+"}", value)
}

func (f *uriFormatter) AddQueryString(parameters []ExecutionParameter) {
	formatter := newQueryStringFormatter()
	f.queryString = formatter.Format(parameters)
}

func (f uriFormatter) Uri() string {
	if f.queryString == "" {
		return f.uri
	}
	return f.uri + "?" + f.queryString
}

func newUriFormatter(baseUri url.URL, route string) *uriFormatter {
	normalizedPath := strings.Trim(baseUri.Path, "/")
	normalizedRoute := strings.Trim(route, "/")
	path := path.Join(normalizedPath, normalizedRoute)
	uri := fmt.Sprintf("%s://%s/%s", baseUri.Scheme, baseUri.Host, path)
	return &uriFormatter{uri, ""}
}
