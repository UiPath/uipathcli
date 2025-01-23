package commandline

import (
	"fmt"
	"os"
	"strings"

	"github.com/UiPath/uipathcli/parser"
)

const longNamedPrefix = "--"

type ArgsParser struct {
	serviceName   string
	resourceName  string
	operationName string
	values        map[string]any
}

func (p ArgsParser) GetServiceName() any {
	return p.serviceName
}

func (p ArgsParser) GetResourceName() any {
	return p.resourceName
}

func (p ArgsParser) GetOperationName() any {
	return p.operationName
}

func (p ArgsParser) GetValue(name string) any {
	return p.values[name]
}

func parseNameArgs(args []string) (string, string, string, []string) {
	serviceName := ""
	resourceName := ""
	operationName := ""

	longNamedArgs := args[1:]
	for i, arg := range args {
		if i > 3 || strings.HasPrefix(arg, longNamedPrefix) {
			break
		}

		switch i {
		case 1:
			serviceName = arg
		case 2:
			resourceName = arg
		case 3:
			operationName = arg
		}
		longNamedArgs = args[i+1:]
	}
	return serviceName, resourceName, operationName, longNamedArgs
}

func parseLongNamedArgValue(value string, flag *FlagDefinition) (any, error) {
	var flagType string
	switch flag.Type {
	case FlagTypeBoolean:
		flagType = parser.ParameterTypeBoolean
	case FlagTypeInteger:
		flagType = parser.ParameterTypeInteger
	case FlagTypeStringArray:
		flagType = parser.ParameterTypeStringArray
	default:
		flagType = parser.ParameterTypeString
	}
	typeConverter := newTypeConverter()
	return typeConverter.Convert(value, *parser.NewParameter(flag.Name, flagType, "", "", "", true, nil, nil, nil))
}

func parseLongNamedArgValues(args []string) (map[string]string, error) {
	values := map[string]string{}
	name := ""
	for _, arg := range append(args, longNamedPrefix) {
		if strings.HasPrefix(arg, longNamedPrefix) {
			if name != "" {
				values[name] = "true"
			}
			name = arg[2:]
		} else if name != "" {
			values[name] = arg
			name = ""
		} else {
			return map[string]string{}, fmt.Errorf("Unknown argument '%s'", arg)
		}
	}
	return values, nil
}

func parseLongNamedFlagValues(args []string, flags []*FlagDefinition) (map[string]any, error) {
	values, err := parseLongNamedArgValues(args)
	if err != nil {
		return map[string]any{}, err
	}

	result := map[string]any{}
	for _, flag := range flags {
		value, ok := values[flag.Name]
		if !ok {
			value, ok = os.LookupEnv(flag.EnvVarName)
		}

		if ok {
			parsedValue, err := parseLongNamedArgValue(value, flag)
			if err != nil {
				return result, err
			}
			result[flag.Name] = parsedValue
		} else if flag.DefaultValue != nil {
			result[flag.Name] = flag.DefaultValue
		}
	}
	return result, nil
}

func NewArgsParser(args []string, flags []*FlagDefinition) (*ArgsParser, error) {
	serviceName, resourceName, operationName, longNamedArgs := parseNameArgs(args)
	values, err := parseLongNamedFlagValues(longNamedArgs, flags)
	return &ArgsParser{serviceName, resourceName, operationName, values}, err
}
