package parser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

const DefaultServerBaseUrl = "https://cloud.uipath.com"
const RawBodyParameterName = "$input"

type OpenApiParser struct{}

func (p OpenApiParser) contains(strs []string, str string) bool {
	for _, s := range strs {
		if s == str {
			return true
		}
	}
	return false
}

func (p OpenApiParser) getTitle(document openapi3.T) string {
	if document.Info == nil {
		return ""
	}
	return document.Info.Title
}

func (p OpenApiParser) getUri(document openapi3.T) (*url.URL, error) {
	server := DefaultServerBaseUrl
	if len(document.Servers) > 0 {
		server = document.Servers[0].URL
	}
	return url.Parse(server)
}

func (p OpenApiParser) formatName(name string) string {
	name = strings.ReplaceAll(name, "$", "")
	name = strings.ReplaceAll(name, "/", "-")
	name = strings.ReplaceAll(name, "_", "-")
	name = ToSnakeCase(name)
	return strings.ToLower(name)
}

func (p OpenApiParser) getName(method string, route string, category *OperationCategory, operation openapi3.Operation) string {
	name := method + route
	if operation.OperationID != "" {
		name = operation.OperationID
	}
	name = p.formatName(name)
	if category != nil {
		name = strings.TrimPrefix(name, category.Name+"-")
	}
	return name
}

func (p OpenApiParser) getSchemaType(schema openapi3.Schema) string {
	if schema.Type == openapi3.TypeArray {
		itemType := schema.Items.Value.Type
		switch itemType {
		case openapi3.TypeBoolean:
			return ParameterTypeBooleanArray
		case openapi3.TypeInteger:
			return ParameterTypeIntegerArray
		case openapi3.TypeNumber:
			return ParameterTypeNumberArray
		case openapi3.TypeObject:
			return ParameterTypeObjectArray
		default:
			return ParameterTypeStringArray
		}
	}

	switch schema.Type {
	case openapi3.TypeBoolean:
		return ParameterTypeBoolean
	case openapi3.TypeInteger:
		return ParameterTypeInteger
	case openapi3.TypeNumber:
		return ParameterTypeNumber
	case openapi3.TypeObject:
		return ParameterTypeObject
	default:
		if schema.Format == "binary" {
			return ParameterTypeBinary
		}
		return ParameterTypeString
	}
}

func (p OpenApiParser) getType(schemaRef *openapi3.SchemaRef) string {
	if schemaRef == nil {
		return ParameterTypeString
	}
	return p.getSchemaType(*schemaRef.Value)
}

func (p OpenApiParser) parseSchema(fieldName string, schemaRef *openapi3.SchemaRef, in string, requiredFieldnames []string) Parameter {
	name := p.formatName(fieldName)
	_type := p.getType(schemaRef)
	required := p.contains(requiredFieldnames, fieldName)
	parameters := []Parameter{}
	description := ""
	var defaultValue interface{}
	if schemaRef != nil {
		description = schemaRef.Value.Description
		defaultValue = schemaRef.Value.Default
		parameters = p.parseSchemas(schemaRef.Value.Properties, in, schemaRef.Value.Required)
	}
	return *NewParameter(name, _type, description, in, fieldName, required, defaultValue, parameters)
}

func (p OpenApiParser) parseSchemas(schemas openapi3.Schemas, in string, requiredFieldnames []string) []Parameter {
	result := []Parameter{}
	for fieldName, schemaRef := range schemas {
		parsed := p.parseSchema(fieldName, schemaRef, in, requiredFieldnames)
		result = append(result, parsed)
	}
	return result
}

func (p OpenApiParser) parseRequestBodyParameters(requestBody *openapi3.RequestBodyRef) (string, []Parameter) {
	parameters := []Parameter{}
	if requestBody == nil {
		return "", parameters
	}
	content := requestBody.Value.Content.Get("application/json")
	if content != nil {
		return "application/json", p.parseSchemas(content.Schema.Value.Properties, ParameterInBody, content.Schema.Value.Required)
	}
	content = requestBody.Value.Content.Get("multipart/form-data")
	if content != nil {
		return "multipart/form-data", p.parseSchemas(content.Schema.Value.Properties, ParameterInForm, content.Schema.Value.Required)
	}
	content = requestBody.Value.Content.Get("application/octet-stream")
	if content != nil {
		parameter := p.parseSchema(RawBodyParameterName, content.Schema, ParameterInBody, []string{RawBodyParameterName})
		return "application/octet-stream", []Parameter{parameter}
	}
	return "", parameters
}

func (p OpenApiParser) parseParameter(param openapi3.Parameter) Parameter {
	fieldName := param.Name
	name := p.formatName(fieldName)
	_type := p.getType(param.Schema)
	required := param.Required
	parameters := []Parameter{}
	var defaultValue interface{}
	if param.Schema != nil {
		defaultValue = param.Schema.Value.Default
		parameters = p.parseSchemas(param.Schema.Value.Properties, param.In, param.Schema.Value.Required)
	}
	return *NewParameter(name, _type, param.Description, param.In, fieldName, required, defaultValue, parameters)
}

func (p OpenApiParser) parseParameters(params openapi3.Parameters) []Parameter {
	parameters := []Parameter{}
	for _, param := range params {
		parameter := p.parseParameter(*param.Value)
		parameters = append(parameters, parameter)
	}
	return parameters
}

func (p OpenApiParser) parseOperationParameters(operation openapi3.Operation, routeParameters openapi3.Parameters) (string, []Parameter) {
	contentType, parameters := p.parseRequestBodyParameters(operation.RequestBody)
	parameters = append(parameters, p.parseParameters(routeParameters)...)
	return contentType, append(parameters, p.parseParameters(operation.Parameters)...)
}

func (p OpenApiParser) getCategory(operation openapi3.Operation, document openapi3.T) *OperationCategory {
	if len(operation.Tags) > 0 {
		name := operation.Tags[0]
		description := ""
		tag := document.Tags.Get(name)
		if tag != nil {
			description = tag.Description
		}
		return NewOperationCategory(p.formatName(name), description)
	}
	return nil
}

func (p OpenApiParser) parseOperation(method string, route string, operation openapi3.Operation, routeParameters openapi3.Parameters, document openapi3.T) Operation {
	category := p.getCategory(operation, document)
	name := p.getName(method, route, category, operation)
	contentType, parameters := p.parseOperationParameters(operation, routeParameters)
	return *NewOperation(name, operation.Summary, method, route, contentType, parameters, nil, false, category)
}

func (p OpenApiParser) parsePath(route string, pathItem openapi3.PathItem, document openapi3.T) []Operation {
	operations := []Operation{}
	for method := range pathItem.Operations() {
		operation := pathItem.GetOperation(method)
		operations = append(operations, p.parseOperation(method, route, *operation, pathItem.Parameters, document))
	}
	return operations
}

func (p OpenApiParser) parse(name string, document openapi3.T) (*Definition, error) {
	operations := []Operation{}
	for path := range document.Paths {
		pathItem := document.Paths.Find(path)
		operations = append(operations, p.parsePath(path, *pathItem, document)...)
	}
	uri, err := p.getUri(document)
	if err != nil {
		return nil, fmt.Errorf("Error parsing server URL: %v", err)
	}
	title := p.getTitle(document)
	return NewDefinition(name, *uri, title, operations), nil
}

func (p OpenApiParser) Parse(name string, data []byte) (*Definition, error) {
	loader := openapi3.NewLoader()
	document, err := loader.LoadFromData(data)
	if err != nil {
		return nil, err
	}
	return p.parse(name, *document)
}
