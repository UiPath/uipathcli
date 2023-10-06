package commandline

import (
	"fmt"
	"strings"

	"github.com/UiPath/uipathcli/parser"
)

type parameterFormatter struct {
	parameter parser.Parameter
}

func (f parameterFormatter) Description() string {
	return f.description(f.parameter)
}

func (f parameterFormatter) UsageExample() string {
	return f.usageExample(f.parameter)
}

func (f parameterFormatter) description(parameter parser.Parameter) string {
	builder := strings.Builder{}

	builder.WriteString(f.humanReadableType(parameter.Type))

	fields := f.descriptionFields(parameter)
	if len(fields) > 0 {
		builder.WriteString(" (")
		builder.WriteString(f.commaSeparatedValues(fields))
		builder.WriteString(")")
	}

	if parameter.Description != "" {
		builder.WriteString("\n")
		builder.WriteString(parameter.Description)
	}

	if len(parameter.AllowedValues) > 0 {
		f.writeSeparator(&builder, "\n\n")
		builder.WriteString("Allowed values:")
		for _, value := range parameter.AllowedValues {
			builder.WriteString(fmt.Sprintf("\n- %v", value))
		}
	}

	example := f.usageExample(parameter)
	if example != "" {
		f.writeSeparator(&builder, "\n\n")
		builder.WriteString("Example:\n")
		builder.WriteString("   " + example)
	}

	return builder.String()
}

func (f parameterFormatter) descriptionFields(parameter parser.Parameter) []interface{} {
	fields := []interface{}{}
	if parameter.Required && parameter.DefaultValue == nil {
		fields = append(fields, "required")
	}
	if parameter.DefaultValue != nil {
		fields = append(fields, fmt.Sprintf("default: %v", parameter.DefaultValue))
	}
	return fields
}

func (f parameterFormatter) usageExample(parameter parser.Parameter) string {
	parameters := map[string]string{}
	f.collectUsageParameters(parameter, "", parameters)

	builder := strings.Builder{}
	for key, value := range parameters {
		f.writeSeparator(&builder, "; ")
		builder.WriteString(fmt.Sprintf("%s=%s", key, value))
	}
	return builder.String()
}

func (f parameterFormatter) collectUsageParameters(parameter parser.Parameter, prefix string, result map[string]string) {
	for _, p := range parameter.Parameters {
		if p.Type == parser.ParameterTypeObjectArray {
			f.collectUsageParameters(p, prefix+p.FieldName+"[0].", result)
			continue
		}
		if p.Type == parser.ParameterTypeObject {
			f.collectUsageParameters(p, prefix+p.FieldName+".", result)
		}
		result[prefix+p.FieldName] = f.humanReadableType(p.Type)
	}
}

func (f parameterFormatter) commaSeparatedValues(values []interface{}) string {
	builder := strings.Builder{}
	for _, value := range values {
		f.writeSeparator(&builder, ", ")
		builder.WriteString(fmt.Sprintf("%v", value))
	}
	return builder.String()
}

func (f parameterFormatter) writeSeparator(builder *strings.Builder, separator string) {
	if builder.Len() > 0 {
		builder.WriteString(separator)
	}
}

func (f parameterFormatter) humanReadableType(_type string) string {
	switch _type {
	case parser.ParameterTypeString:
		return "string"
	case parser.ParameterTypeBinary:
		return "binary"
	case parser.ParameterTypeInteger:
		return "integer"
	case parser.ParameterTypeNumber:
		return "float"
	case parser.ParameterTypeBoolean:
		return "boolean"
	case parser.ParameterTypeStringArray:
		return "string,string,..."
	case parser.ParameterTypeIntegerArray:
		return "integer,integer,..."
	case parser.ParameterTypeNumberArray:
		return "float,float,..."
	case parser.ParameterTypeBooleanArray:
		return "boolean,boolean,..."
	case parser.ParameterTypeObjectArray:
		return "object (multiple)"
	default:
		return "object"
	}
}

func newParameterFormatter(parameter parser.Parameter) *parameterFormatter {
	return &parameterFormatter{parameter}
}
