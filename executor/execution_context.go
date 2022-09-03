package executor

import "net/url"

type ExecutionContext struct {
	Method           string
	BaseUri          url.URL
	Route            string
	PathParameters   []ExecutionParameter
	QueryParameters  []ExecutionParameter
	HeaderParameters []ExecutionParameter
	BodyParameters   []ExecutionParameter
	FormParameters   []ExecutionParameter
	ClientId         string
	ClientSecret     string
	Insecure         bool
	Debug            bool
}

func NewExecutionContext(
	method string,
	uri url.URL,
	route string,
	pathParameters []ExecutionParameter,
	queryParameters []ExecutionParameter,
	headerParameters []ExecutionParameter,
	bodyParameters []ExecutionParameter,
	formParameters []ExecutionParameter,
	clientId string,
	clientSecret string,
	insecure bool,
	debug bool) *ExecutionContext {
	return &ExecutionContext{method, uri, route, pathParameters, queryParameters, headerParameters, bodyParameters, formParameters, clientId, clientSecret, insecure, debug}
}
