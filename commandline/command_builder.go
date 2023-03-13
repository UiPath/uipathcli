package commandline

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/utils"
	"github.com/urfave/cli/v2"
)

const insecureFlagName = "insecure"
const debugFlagName = "debug"
const profileFlagName = "profile"
const uriFlagName = "uri"
const organizationFlagName = "organization"
const tenantFlagName = "tenant"
const helpFlagName = "help"

const outputFormatFlagName = "output"
const outputFormatJson = "json"
const outputFormatText = "text"

const queryFlagName = "query"

// The CommandBuilder is creating all available operations and arguments for the CLI.
type CommandBuilder struct {
	Input              utils.Stream
	StdIn              io.Reader
	StdOut             io.Writer
	StdErr             io.Writer
	ConfigProvider     config.ConfigProvider
	Executor           executor.Executor
	PluginExecutor     executor.Executor
	DefinitionProvider DefinitionProvider
}

func (b CommandBuilder) sort(commands []*cli.Command) {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
}

func (b CommandBuilder) getBodyInput(bodyParameters []executor.ExecutionParameter) (utils.Stream, error) {
	if b.Input != nil {
		return b.Input, nil
	}
	if len(bodyParameters) == 1 && bodyParameters[0].Name == parser.RawBodyParameterName {
		switch value := bodyParameters[0].Value.(type) {
		case utils.Stream:
			return value, nil
		default:
			data := []byte(fmt.Sprintf("%v", value))
			return utils.NewMemoryStream(parser.RawBodyParameterName, data), nil
		}
	}
	return nil, nil
}

func (b CommandBuilder) createExecutionParameters(context *cli.Context, in string, operation parser.Operation, additionalParameters map[string]string) ([]executor.ExecutionParameter, error) {
	typeConverter := newTypeConverter()

	parameters := []executor.ExecutionParameter{}
	for _, param := range operation.Parameters {
		if param.In == in && context.IsSet(param.Name) {
			value, err := typeConverter.Convert(context.String(param.Name), param)
			if err != nil {
				return nil, err
			}
			parameter := executor.NewExecutionParameter(param.FieldName, value)
			parameters = append(parameters, *parameter)
		} else if param.In == in && param.Required && param.DefaultValue != nil {
			parameter := executor.NewExecutionParameter(param.FieldName, param.DefaultValue)
			parameters = append(parameters, *parameter)
		}
	}
	for key, value := range additionalParameters {
		parameter := executor.NewExecutionParameter(key, value)
		parameters = append(parameters, *parameter)
	}
	return parameters, nil
}

func (b CommandBuilder) parameterRequired(parameter parser.Parameter) bool {
	return parameter.Required && parameter.DefaultValue == nil
}

func (b CommandBuilder) formatAllowedValues(values []interface{}) string {
	if values == nil {
		return ""
	}
	result := ""
	separator := ""
	for _, value := range values {
		result += fmt.Sprintf("%s%v", separator, value)
		separator = ", "
	}
	return result
}

func (b CommandBuilder) parameterDescription(parameter parser.Parameter) string {
	description := parameter.Description
	if parameter.DefaultValue != nil {
		description = fmt.Sprintf("%s (default: %v)", parameter.Description, parameter.DefaultValue)
	} else if b.parameterRequired(parameter) {
		description = fmt.Sprintf("%s (required)", parameter.Description)
	}

	allowedValues := b.formatAllowedValues(parameter.AllowedValues)
	if allowedValues != "" {
		return fmt.Sprintf("%s \nallowed values: %s", description, allowedValues)
	}
	return description
}

func (b CommandBuilder) createFlags(parameters []parser.Parameter) []cli.Flag {
	flags := []cli.Flag{}
	for _, parameter := range parameters {
		flag := cli.StringFlag{
			Name:  parameter.Name,
			Usage: b.parameterDescription(parameter),
		}
		flags = append(flags, &flag)
	}
	return flags
}

func (b CommandBuilder) sortParameters(parameters []parser.Parameter) {
	sort.Slice(parameters, func(i, j int) bool {
		if parameters[i].Required && !parameters[j].Required {
			return true
		}
		if !parameters[i].Required && parameters[j].Required {
			return false
		}
		return parameters[i].Name < parameters[j].Name
	})
}

func (b CommandBuilder) outputFormat(config config.Config, context *cli.Context) (string, error) {
	outputFormat := context.String(outputFormatFlagName)
	if outputFormat == "" {
		outputFormat = config.Output
	}
	if outputFormat == "" {
		outputFormat = outputFormatJson
	}
	if outputFormat != outputFormatJson && outputFormat != outputFormatText {
		return "", fmt.Errorf("Invalid output format '%s', allowed values: %s, %s", outputFormat, outputFormatJson, outputFormatText)
	}
	return outputFormat, nil
}

func (b CommandBuilder) createBaseUri(operation parser.Operation, config config.Config, context *cli.Context) (url.URL, error) {
	uriArgument, err := b.parseUriArgument(context)
	if err != nil {
		return operation.BaseUri, err
	}

	builder := NewUriBuilder(operation.BaseUri)
	builder.OverrideUri(config.Uri)
	builder.OverrideUri(uriArgument)
	return builder.Uri(), nil
}

func (b CommandBuilder) parseUriArgument(context *cli.Context) (*url.URL, error) {
	uriFlag := context.String(uriFlagName)
	if uriFlag == "" {
		return nil, nil
	}
	uriArgument, err := url.Parse(uriFlag)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s argument: %v", uriFlagName, err)
	}
	return uriArgument, nil
}

func (b CommandBuilder) getValue(parameter parser.Parameter, context *cli.Context, config config.Config) string {
	value := context.String(parameter.Name)
	if value != "" {
		return value
	}
	value = config.Path[parameter.Name]
	if value != "" {
		return value
	}
	value = config.Query[parameter.Name]
	if value != "" {
		return value
	}
	value = config.Header[parameter.Name]
	if value != "" {
		return value
	}
	if parameter.DefaultValue != nil {
		return fmt.Sprintf("%v", parameter.DefaultValue)
	}
	return ""
}

func (b CommandBuilder) validateArguments(context *cli.Context, parameters []parser.Parameter, config config.Config) error {
	err := errors.New("Invalid arguments:")
	result := true
	for _, parameter := range parameters {
		value := b.getValue(parameter, context, config)
		if parameter.Required && value == "" {
			result = false
			err = fmt.Errorf("%w\n  Argument --%s is missing", err, parameter.Name)
		}
		if value != "" && len(parameter.AllowedValues) > 0 {
			valid := false
			for _, allowedValue := range parameter.AllowedValues {
				if fmt.Sprintf("%v", allowedValue) == value {
					valid = true
					break
				}
			}
			if !valid {
				allowedValues := b.formatAllowedValues(parameter.AllowedValues)
				result = false
				err = fmt.Errorf("%w\n  Argument value '%v' for --%s is invalid, allowed values: %s", err, value, parameter.Name, allowedValues)
			}
		}
	}
	if result {
		return nil
	}
	return err
}

func (b CommandBuilder) logger(context executor.ExecutionContext, writer io.Writer, errorWriter io.Writer) log.Logger {
	if context.Debug {
		return log.NewDebugLogger(writer, errorWriter)
	}
	return log.NewDefaultLogger(errorWriter)
}

func (b CommandBuilder) outputWriter(context executor.ExecutionContext, writer io.Writer, format string, query string) output.OutputWriter {
	var transformer output.Transformer = output.NewDefaultTransformer()
	if query != "" {
		transformer = output.NewJmesPathTransformer(query)
	}
	if format == outputFormatText {
		return output.NewTextOutputWriter(writer, transformer)
	}
	return output.NewJsonOutputWriter(writer, transformer)
}

func (b CommandBuilder) executeCommand(context executor.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if context.Plugin != nil {
		return b.PluginExecutor.Call(context, writer, logger)
	}
	return b.Executor.Call(context, writer, logger)
}

func (b CommandBuilder) createOperationCommand(definition parser.Definition, operation parser.Operation) *cli.Command {
	flags := b.CreateDefaultFlags(true)
	flags = append(flags, b.HelpFlag())
	parameters := operation.Parameters
	b.sortParameters(parameters)
	flags = append(flags, b.createFlags(parameters)...)

	return &cli.Command{
		Name:        operation.Name,
		Usage:       operation.Summary,
		Description: operation.Description,
		Flags:       flags,
		Action: func(context *cli.Context) error {
			profileName := context.String(profileFlagName)
			config := b.ConfigProvider.Config(profileName)
			if config == nil {
				return fmt.Errorf("Could not find profile '%s'", profileName)
			}
			outputFormat, err := b.outputFormat(*config, context)
			if err != nil {
				return err
			}
			query := context.String(queryFlagName)

			baseUri, err := b.createBaseUri(operation, *config, context)
			if err != nil {
				return err
			}

			if b.Input == nil {
				err = b.validateArguments(context, operation.Parameters, *config)
				if err != nil {
					return err
				}
			}

			pathParameters, err := b.createExecutionParameters(context, parser.ParameterInPath, operation, config.Path)
			if err != nil {
				return err
			}
			queryParameters, err := b.createExecutionParameters(context, parser.ParameterInQuery, operation, config.Query)
			if err != nil {
				return err
			}
			headerParameters, err := b.createExecutionParameters(context, parser.ParameterInHeader, operation, config.Header)
			if err != nil {
				return err
			}
			bodyParameters, err := b.createExecutionParameters(context, parser.ParameterInBody, operation, map[string]string{})
			if err != nil {
				return err
			}
			formParameters, err := b.createExecutionParameters(context, parser.ParameterInForm, operation, map[string]string{})
			if err != nil {
				return err
			}
			input, err := b.getBodyInput(bodyParameters)
			if err != nil {
				return err
			}
			organization := context.String(organizationFlagName)
			if organization == "" {
				organization = config.Organization
			}
			tenant := context.String(tenantFlagName)
			if tenant == "" {
				tenant = config.Tenant
			}
			insecure := context.Bool(insecureFlagName) || config.Insecure
			debug := context.Bool(debugFlagName) || config.Debug
			parameters := executor.NewExecutionContextParameters(
				pathParameters,
				queryParameters,
				headerParameters,
				bodyParameters,
				formParameters)
			executionContext := executor.NewExecutionContext(
				organization,
				tenant,
				operation.Method,
				baseUri,
				operation.Route,
				operation.ContentType,
				input,
				*parameters,
				config.Auth,
				insecure,
				debug,
				operation.Plugin)

			var wg sync.WaitGroup
			wg.Add(3)
			reader, writer := io.Pipe()
			go func(reader *io.PipeReader) {
				defer wg.Done()
				defer reader.Close()
				io.Copy(b.StdOut, reader)
			}(reader)
			errorReader, errorWriter := io.Pipe()
			go func(errorReader *io.PipeReader) {
				defer wg.Done()
				defer errorReader.Close()
				io.Copy(b.StdErr, errorReader)
			}(errorReader)

			go func(context executor.ExecutionContext, writer *io.PipeWriter, errorWriter *io.PipeWriter) {
				defer wg.Done()
				defer writer.Close()
				defer errorWriter.Close()
				outputWriter := b.outputWriter(context, writer, outputFormat, query)
				logger := b.logger(context, writer, errorWriter)
				err = b.executeCommand(context, outputWriter, logger)
			}(*executionContext, writer, errorWriter)

			wg.Wait()
			return err
		},
		HideHelp: true,
		Hidden:   operation.Hidden,
	}
}

func (b CommandBuilder) createCategoryCommand(operation parser.Operation) *cli.Command {
	return &cli.Command{
		Name:        operation.Category.Name,
		Description: operation.Category.Description,
		Flags: []cli.Flag{
			b.HelpFlag(),
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) createServiceCommandCategory(definition parser.Definition, operation parser.Operation, categories map[string]*cli.Command) (bool, *cli.Command) {
	isNewCategory := false
	operationCommand := b.createOperationCommand(definition, operation)
	command, found := categories[operation.Category.Name]
	if !found {
		command = b.createCategoryCommand(operation)
		categories[operation.Category.Name] = command
		isNewCategory = true
	}
	command.Subcommands = append(command.Subcommands, operationCommand)
	return isNewCategory, command
}

func (b CommandBuilder) createServiceCommand(definition parser.Definition) *cli.Command {
	categories := map[string]*cli.Command{}
	commands := []*cli.Command{}
	for _, operation := range definition.Operations {
		if operation.Category == nil {
			command := b.createOperationCommand(definition, operation)
			commands = append(commands, command)
			continue
		}
		isNewCategory, command := b.createServiceCommandCategory(definition, operation, categories)
		if isNewCategory {
			commands = append(commands, command)
		}
	}
	b.sort(commands)
	for _, command := range commands {
		b.sort(command.Subcommands)
	}

	return &cli.Command{
		Name:        definition.Name,
		Description: definition.Description,
		Flags: []cli.Flag{
			b.HelpFlag(),
		},
		Subcommands: commands,
		HideHelp:    true,
	}
}

func (b CommandBuilder) createAutoCompleteEnableCommand() *cli.Command {
	const shellFlagName = "shell"
	const powershellFlagValue = "powershell"
	const bashFlagValue = "bash"
	const fileFlagName = "file"

	return &cli.Command{
		Name:        "enable",
		Description: "Enables auto complete in your shell",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     shellFlagName,
				Usage:    fmt.Sprintf("%s, %s", powershellFlagValue, bashFlagValue),
				Required: true,
			},
			&cli.StringFlag{
				Name:   fileFlagName,
				Hidden: true,
			},
			b.HelpFlag(),
		},
		Action: func(context *cli.Context) error {
			shell := context.String(shellFlagName)
			filePath := context.String(fileFlagName)
			handler := newAutoCompleteHandler()
			output, err := handler.EnableCompleter(shell, filePath)
			if err != nil {
				return err
			}
			fmt.Fprintln(b.StdOut, output)
			return nil
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) createAutoCompleteCompleteCommand() *cli.Command {
	return &cli.Command{
		Name:        "complete",
		Description: "Returns the autocomplete suggestions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "command",
				Usage:    "The command to autocomplete",
				Required: true,
			},
			b.HelpFlag(),
		},
		Action: func(context *cli.Context) error {
			commandText := context.String("command")
			exclude := []string{
				"--" + insecureFlagName,
				"--" + debugFlagName,
				"--" + profileFlagName,
				"--" + uriFlagName,
				"--" + helpFlagName,
				"--" + outputFormatFlagName,
				"--" + queryFlagName,
				"--" + organizationFlagName,
				"--" + tenantFlagName,
			}
			args := strings.Split(commandText, " ")
			definitions, err := b.loadAutocompleteDefinitions(args)
			if err != nil {
				return err
			}
			commands := b.createServiceCommands(definitions)
			handler := newAutoCompleteHandler()
			words := handler.Find(commandText, commands, exclude)
			for _, word := range words {
				fmt.Fprintln(b.StdOut, word)
			}
			return nil
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) createAutoCompleteCommand() *cli.Command {
	return &cli.Command{
		Name:        "autocomplete",
		Description: "Commands for autocompletion",
		Flags: []cli.Flag{
			b.HelpFlag(),
		},
		Subcommands: []*cli.Command{
			b.createAutoCompleteEnableCommand(),
			b.createAutoCompleteCompleteCommand(),
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) createConfigCommand() *cli.Command {
	authFlagName := "auth"
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  authFlagName,
			Usage: fmt.Sprintf("Authorization type: %s, %s, %s", CredentialsAuth, LoginAuth, PatAuth),
		},
		&cli.StringFlag{
			Name:    profileFlagName,
			Usage:   "Profile to configure",
			EnvVars: []string{"UIPATH_PROFILE"},
			Value:   config.DefaultProfile,
		},
		b.HelpFlag(),
	}

	return &cli.Command{
		Name:        "config",
		Description: "Interactive command to configure the CLI",
		Flags:       flags,
		Subcommands: []*cli.Command{
			b.createConfigSetCommand(),
		},
		Action: func(context *cli.Context) error {
			auth := context.String(authFlagName)
			profileName := context.String(profileFlagName)
			handler := ConfigCommandHandler{
				StdIn:          b.StdIn,
				StdOut:         b.StdOut,
				ConfigProvider: b.ConfigProvider,
			}
			return handler.Configure(auth, profileName)
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) createConfigSetCommand() *cli.Command {
	keyFlagName := "key"
	valueFlagName := "value"
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     keyFlagName,
			Usage:    "The key",
			Required: true,
		},
		&cli.StringFlag{
			Name:     valueFlagName,
			Usage:    "The value to set",
			Required: true,
		},
		&cli.StringFlag{
			Name:    profileFlagName,
			Usage:   "Profile to configure",
			EnvVars: []string{"UIPATH_PROFILE"},
			Value:   config.DefaultProfile,
		},
		b.HelpFlag(),
	}
	return &cli.Command{
		Name:        "set",
		Description: "Set config parameters",
		Flags:       flags,
		Action: func(context *cli.Context) error {
			profileName := context.String(profileFlagName)
			key := context.String(keyFlagName)
			value := context.String(valueFlagName)
			handler := ConfigCommandHandler{
				StdIn:          b.StdIn,
				StdOut:         b.StdOut,
				ConfigProvider: b.ConfigProvider,
			}
			return handler.Set(key, value, profileName)
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) loadDefinitions(args []string) ([]parser.Definition, error) {
	if len(args) <= 1 || strings.HasPrefix(args[1], "--") {
		return b.DefinitionProvider.Index()
	}
	definition, err := b.DefinitionProvider.Load(args[1])
	if definition == nil {
		return nil, err
	}
	return []parser.Definition{*definition}, err
}

func (b CommandBuilder) loadAutocompleteDefinitions(args []string) ([]parser.Definition, error) {
	if len(args) <= 2 {
		return b.DefinitionProvider.Index()
	}
	return b.loadDefinitions(args)
}

func (b CommandBuilder) createServiceCommands(definitions []parser.Definition) []*cli.Command {
	commands := []*cli.Command{}
	for _, e := range definitions {
		command := b.createServiceCommand(e)
		commands = append(commands, command)
	}
	return commands
}

func (b CommandBuilder) Create(args []string) ([]*cli.Command, error) {
	definitions, err := b.loadDefinitions(args)
	if err != nil {
		return nil, err
	}
	servicesCommands := b.createServiceCommands(definitions)
	autocompleteCommand := b.createAutoCompleteCommand()
	configCommand := b.createConfigCommand()
	commands := append(servicesCommands, autocompleteCommand, configCommand)
	return commands, nil
}

func (b CommandBuilder) CreateDefaultFlags(hidden bool) []cli.Flag {
	return []cli.Flag{
		&cli.BoolFlag{
			Name:    debugFlagName,
			Usage:   "Enable debug output",
			EnvVars: []string{"UIPATH_DEBUG"},
			Value:   false,
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    profileFlagName,
			Usage:   "Config profile to use",
			EnvVars: []string{"UIPATH_PROFILE"},
			Value:   config.DefaultProfile,
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    uriFlagName,
			Usage:   "Server Base-URI",
			EnvVars: []string{"UIPATH_URI"},
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    organizationFlagName,
			Usage:   "Organization name",
			EnvVars: []string{"UIPATH_ORGANIZATION"},
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    tenantFlagName,
			Usage:   "Tenant name",
			EnvVars: []string{"UIPATH_TENANT"},
			Hidden:  hidden,
		},
		&cli.BoolFlag{
			Name:    insecureFlagName,
			Usage:   "Disable HTTPS certificate check",
			EnvVars: []string{"UIPATH_INSECURE"},
			Value:   false,
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:    outputFormatFlagName,
			Usage:   fmt.Sprintf("Set output format: %s (default), %s", outputFormatJson, outputFormatText),
			EnvVars: []string{"UIPATH_OUTPUT"},
			Value:   "",
			Hidden:  hidden,
		},
		&cli.StringFlag{
			Name:   queryFlagName,
			Usage:  "Perform JMESPath query on output",
			Value:  "",
			Hidden: hidden,
		},
	}
}

func (b CommandBuilder) HelpFlag() cli.Flag {
	return &cli.BoolFlag{
		Name:   helpFlagName,
		Usage:  "Show help",
		Value:  false,
		Hidden: true,
	}
}
