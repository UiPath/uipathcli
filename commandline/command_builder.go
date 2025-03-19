package commandline

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/output"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/utils/stream"
)

// The CommandBuilder is creating all available operations and arguments for the CLI.
type CommandBuilder struct {
	Input              stream.Stream
	StdIn              io.Reader
	StdOut             io.Writer
	StdErr             io.Writer
	ConfigProvider     config.ConfigProvider
	Executor           executor.Executor
	PluginExecutor     executor.Executor
	DefinitionProvider DefinitionProvider
}

func (b CommandBuilder) sort(commands []*CommandDefinition) {
	sort.Slice(commands, func(i, j int) bool {
		return commands[i].Name < commands[j].Name
	})
}

func (b CommandBuilder) fileInput(context *CommandExecContext, parameters []parser.Parameter) stream.Stream {
	value := context.String(FlagNameFile)
	if value == "" {
		return nil
	}
	if value == FlagValueFromStdIn {
		return b.Input
	}
	for _, param := range parameters {
		if strings.EqualFold(param.FieldName, FlagNameFile) {
			return nil
		}
	}
	return stream.NewFileStream(value)
}

func (b CommandBuilder) createExecutionParameters(context *CommandExecContext, config *config.Config, operation parser.Operation) (executor.ExecutionParameters, error) {
	typeConverter := newTypeConverter()

	parameters := []executor.ExecutionParameter{}
	for _, param := range operation.Parameters {
		if context.IsSet(param.Name) && param.IsArray() {
			value, err := typeConverter.ConvertArray(context.StringSlice(param.Name), param)
			if err != nil {
				return nil, err
			}
			parameter := executor.NewExecutionParameter(param.FieldName, value, param.In)
			parameters = append(parameters, *parameter)
		} else if context.IsSet(param.Name) {
			value, err := typeConverter.Convert(context.String(param.Name), param)
			if err != nil {
				return nil, err
			}
			parameter := executor.NewExecutionParameter(param.FieldName, value, param.In)
			parameters = append(parameters, *parameter)
		} else if configValue, ok := config.Parameter[param.Name]; ok {
			value, err := typeConverter.Convert(configValue, param)
			if err != nil {
				return nil, err
			}
			parameter := executor.NewExecutionParameter(param.FieldName, value, param.In)
			parameters = append(parameters, *parameter)
		} else if param.Required && param.DefaultValue != nil {
			parameter := executor.NewExecutionParameter(param.FieldName, param.DefaultValue, param.In)
			parameters = append(parameters, *parameter)
		}
	}
	parameters = append(parameters, b.createExecutionParametersFromConfigMap(config.Header, parser.ParameterInHeader)...)
	return parameters, nil
}

func (b CommandBuilder) createExecutionParametersFromConfigMap(params map[string]string, in string) executor.ExecutionParameters {
	parameters := []executor.ExecutionParameter{}
	for key, value := range params {
		parameter := executor.NewExecutionParameter(key, value, in)
		parameters = append(parameters, *parameter)
	}
	return parameters
}

func (b CommandBuilder) formatAllowedValues(values []interface{}) string {
	result := ""
	separator := ""
	for _, value := range values {
		result += fmt.Sprintf("%s%v", separator, value)
		separator = ", "
	}
	return result
}

func (b CommandBuilder) createFlags(parameters []parser.Parameter) []*FlagDefinition {
	flags := []*FlagDefinition{}
	for _, parameter := range parameters {
		formatter := newParameterFormatter(parameter)
		flagType := FlagTypeString
		if parameter.IsArray() {
			flagType = FlagTypeStringArray
		}
		flag := NewFlag(parameter.Name, formatter.Description(), flagType)
		flags = append(flags, flag)
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

func (b CommandBuilder) outputFormat(config config.Config, context *CommandExecContext) (string, error) {
	outputFormat := context.String(FlagNameOutputFormat)
	if outputFormat == "" {
		outputFormat = config.Output
	}
	if outputFormat == "" {
		outputFormat = FlagValueOutputFormatJson
	}
	if outputFormat != FlagValueOutputFormatJson && outputFormat != FlagValueOutputFormatText {
		return "", fmt.Errorf("Invalid output format '%s', allowed values: %s, %s", outputFormat, FlagValueOutputFormatJson, FlagValueOutputFormatText)
	}
	return outputFormat, nil
}

func (b CommandBuilder) createBaseUri(operation parser.Operation, config config.Config, context *CommandExecContext) (url.URL, error) {
	uriArgument, err := b.parseUriArgument(context)
	if err != nil {
		return operation.BaseUri, err
	}

	builder := NewUriBuilder(operation.BaseUri)
	builder.OverrideUri(config.Uri)
	builder.OverrideUri(uriArgument)
	return builder.Uri(), nil
}

func (b CommandBuilder) createIdentityUri(context *CommandExecContext, config config.Config, baseUri url.URL) (*url.URL, error) {
	uri := context.String(FlagNameIdentityUri)
	if uri != "" {
		identityUri, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("Error parsing %s argument: %w", FlagNameIdentityUri, err)
		}
		return identityUri, nil
	}

	value := config.Auth.Config["uri"]
	uri, valid := value.(string)
	if valid && uri != "" {
		identityUri, err := url.Parse(uri)
		if err != nil {
			return nil, fmt.Errorf("Error parsing identity uri config: %w", err)
		}
		return identityUri, nil
	}
	identityUri, err := url.Parse(fmt.Sprintf("%s://%s/identity_", baseUri.Scheme, baseUri.Host))
	if err != nil {
		return nil, fmt.Errorf("Error parsing identity uri: %w", err)
	}
	return identityUri, nil
}

func (b CommandBuilder) parseUriArgument(context *CommandExecContext) (*url.URL, error) {
	uriFlag := context.String(FlagNameUri)
	if uriFlag == "" {
		return nil, nil
	}
	uriArgument, err := url.Parse(uriFlag)
	if err != nil {
		return nil, fmt.Errorf("Error parsing %s argument: %w", FlagNameUri, err)
	}
	return uriArgument, nil
}

func (b CommandBuilder) getValue(parameter parser.Parameter, context *CommandExecContext, config config.Config) string {
	value := context.String(parameter.Name)
	if value != "" {
		return value
	}
	value = config.Parameter[parameter.Name]
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

func (b CommandBuilder) validateArguments(context *CommandExecContext, parameters []parser.Parameter, config config.Config) error {
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

func (b CommandBuilder) logger(ctx executor.ExecutionContext, writer io.Writer) log.Logger {
	if ctx.Debug {
		return log.NewDebugLogger(writer)
	}
	return log.NewDefaultLogger(writer)
}

func (b CommandBuilder) outputWriter(writer io.Writer, format string, query string) output.OutputWriter {
	var transformer output.Transformer = output.NewDefaultTransformer()
	if query != "" {
		transformer = output.NewJmesPathTransformer(query)
	}
	if format == FlagValueOutputFormatText {
		return output.NewTextOutputWriter(writer, transformer)
	}
	return output.NewJsonOutputWriter(writer, transformer)
}

func (b CommandBuilder) executeCommand(ctx executor.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
	if ctx.Plugin != nil {
		return b.PluginExecutor.Call(ctx, writer, logger)
	}
	return b.Executor.Call(ctx, writer, logger)
}

func (b CommandBuilder) operationId() string {
	bytes := make([]byte, 16)
	_, _ = rand.Read(bytes)
	return fmt.Sprintf("%x%x%x%x%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:])
}

func (b CommandBuilder) createOperationCommand(operation parser.Operation) *CommandDefinition {
	parameters := operation.Parameters
	b.sortParameters(parameters)

	flags := NewFlagBuilder().
		AddFlags(b.createFlags(parameters)).
		AddDefaultFlags(true).
		AddHelpFlag().
		Build()

	return NewCommand(operation.Name, operation.Summary, operation.Description).
		WithFlags(flags).
		WithHelpTemplate(OperationCommandHelpTemplate).
		WithHidden(operation.Hidden).
		WithAction(func(context *CommandExecContext) error {
			profileName := context.String(FlagNameProfile)
			config := b.ConfigProvider.Config(profileName)
			if config == nil {
				return fmt.Errorf("Could not find profile '%s'", profileName)
			}
			outputFormat, err := b.outputFormat(*config, context)
			if err != nil {
				return err
			}
			query := context.String(FlagNameQuery)
			wait := context.String(FlagNameWait)
			waitTimeout := context.Int(FlagNameWaitTimeout)

			baseUri, err := b.createBaseUri(operation, *config, context)
			if err != nil {
				return err
			}

			input := b.fileInput(context, operation.Parameters)
			if input == nil {
				err = b.validateArguments(context, operation.Parameters, *config)
				if err != nil {
					return err
				}
			}

			parameters, err := b.createExecutionParameters(context, config, operation)
			if err != nil {
				return err
			}

			organization := context.String(FlagNameOrganization)
			if organization == "" {
				organization = config.Organization
			}
			tenant := context.String(FlagNameTenant)
			if tenant == "" {
				tenant = config.Tenant
			}
			insecure := context.Bool(FlagNameInsecure) || config.Insecure
			timeout := time.Duration(context.Int(FlagNameCallTimeout)) * time.Second
			if timeout < 0 {
				return fmt.Errorf("Invalid value for '%s'", FlagNameCallTimeout)
			}
			maxAttempts := context.Int(FlagNameMaxAttempts)
			if maxAttempts < 1 {
				return fmt.Errorf("Invalid value for '%s'", FlagNameMaxAttempts)
			}
			debug := context.Bool(FlagNameDebug) || config.Debug
			identityUri, err := b.createIdentityUri(context, *config, baseUri)
			if err != nil {
				return err
			}
			operationId := b.operationId()

			executionContext := executor.NewExecutionContext(
				organization,
				tenant,
				operation.Method,
				baseUri,
				operation.Route,
				operation.ContentType,
				input,
				parameters,
				config.Auth,
				*identityUri,
				operation.Plugin,
				debug,
				*executor.NewExecutionSettings(operationId, timeout, maxAttempts, insecure),
			)

			if wait != "" {
				return b.executeWait(*executionContext, outputFormat, query, wait, waitTimeout)
			}
			return b.execute(*executionContext, outputFormat, query, nil)
		})
}

func (b CommandBuilder) executeWait(ctx executor.ExecutionContext, outputFormat string, query string, wait string, waitTimeout int) error {
	logger := log.NewDefaultLogger(b.StdErr)
	outputWriter := output.NewMemoryOutputWriter()
	for start := time.Now(); time.Since(start) < time.Duration(waitTimeout)*time.Second; {
		err := b.execute(ctx, "json", "", outputWriter)
		result, evaluationErr := b.evaluateWaitCondition(outputWriter.Response(), wait)
		if evaluationErr != nil {
			return evaluationErr
		}
		if result {
			resultWriter := b.outputWriter(b.StdOut, outputFormat, query)
			_ = resultWriter.WriteResponse(outputWriter.Response())
			return err
		}
		logger.LogError("Condition is not met yet. Waiting...\n")
		time.Sleep(1 * time.Second)
	}
	return errors.New("Timed out waiting for condition")
}

func (b CommandBuilder) evaluateWaitCondition(response output.ResponseInfo, wait string) (bool, error) {
	body, err := io.ReadAll(response.Body)
	if err != nil {
		return false, nil
	}
	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return false, nil
	}
	transformer := output.NewJmesPathTransformer(wait)
	result, err := transformer.Execute(data)
	if err != nil {
		return false, err
	}
	value, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("Error in wait condition: JMESPath expression needs to return boolean")
	}
	return value, nil
}

func (b CommandBuilder) execute(ctx executor.ExecutionContext, outputFormat string, query string, outputWriter output.OutputWriter) error {
	var wg sync.WaitGroup
	wg.Add(3)
	reader, writer := io.Pipe()
	go func() {
		defer wg.Done()
		defer reader.Close()
		_, _ = io.Copy(b.StdOut, reader)
	}()
	errorReader, errorWriter := io.Pipe()
	go func() {
		defer wg.Done()
		defer errorReader.Close()
		_, _ = io.Copy(b.StdErr, errorReader)
	}()

	var err error
	go func() {
		defer wg.Done()
		defer writer.Close()
		defer errorWriter.Close()
		if outputWriter == nil {
			outputWriter = b.outputWriter(writer, outputFormat, query)
		}
		logger := b.logger(ctx, errorWriter)
		err = b.executeCommand(ctx, outputWriter, logger)
	}()

	wg.Wait()
	return err
}

func (b CommandBuilder) createCategoryCommand(operation parser.Operation) *CommandDefinition {
	flags := NewFlagBuilder().
		AddHelpFlag().
		AddServiceVersionFlag(true).
		Build()

	return NewCommand(operation.Category.Name, operation.Category.Summary, operation.Category.Description).
		WithFlags(flags)
}

func (b CommandBuilder) createServiceCommandCategory(operation parser.Operation, categories map[string]*CommandDefinition) (bool, *CommandDefinition) {
	isNewCategory := false
	operationCommand := b.createOperationCommand(operation)
	command, found := categories[operation.Category.Name]
	if !found {
		command = b.createCategoryCommand(operation)
		categories[operation.Category.Name] = command
		isNewCategory = true
	}
	command.Subcommands = append(command.Subcommands, operationCommand)
	return isNewCategory, command
}

func (b CommandBuilder) createServiceCommand(definition parser.Definition) *CommandDefinition {
	categories := map[string]*CommandDefinition{}
	commands := []*CommandDefinition{}
	for _, operation := range definition.Operations {
		if operation.Category == nil {
			command := b.createOperationCommand(operation)
			commands = append(commands, command)
			continue
		}
		isNewCategory, command := b.createServiceCommandCategory(operation, categories)
		if isNewCategory {
			commands = append(commands, command)
		}
	}
	b.sort(commands)
	for _, command := range commands {
		b.sort(command.Subcommands)
	}

	flags := NewFlagBuilder().
		AddHelpFlag().
		AddServiceVersionFlag(true).
		Build()

	return NewCommand(definition.Name, definition.Summary, definition.Description).
		WithFlags(flags).
		WithSubcommands(commands)
}

func (b CommandBuilder) createAutoCompleteEnableCommand() *CommandDefinition {
	const shellFlagName = "shell"
	const fileFlagName = "file"

	flags := NewFlagBuilder().
		AddFlag(NewFlag(shellFlagName, fmt.Sprintf("%s, %s", AutocompletePowershell, AutocompleteBash), FlagTypeString).
			WithRequired(true)).
		AddFlag(NewFlag(fileFlagName, "The profile file path", FlagTypeString).
			WithHidden(true)).
		AddHelpFlag().
		Build()

	return NewCommand("enable", "Enable auto complete", "Enables auto complete in your shell").
		WithFlags(flags).
		WithAction(func(context *CommandExecContext) error {
			shell := context.String(shellFlagName)
			filePath := context.String(fileFlagName)
			handler := newAutoCompleteHandler()
			output, err := handler.EnableCompleter(shell, filePath)
			if err != nil {
				return err
			}
			fmt.Fprintln(b.StdOut, output)
			return nil
		})
}

func (b CommandBuilder) createAutoCompleteCompleteCommand(serviceVersion string) *CommandDefinition {
	const commandFlagName = "command"

	flags := NewFlagBuilder().
		AddFlag(NewFlag(commandFlagName, "The command to autocomplete", FlagTypeString).
			WithRequired(true)).
		AddHelpFlag().
		Build()

	return NewCommand("complete", "Autocomplete suggestions", "Returns the autocomplete suggestions").
		WithFlags(flags).
		WithAction(func(context *CommandExecContext) error {
			commandText := context.String(commandFlagName)
			exclude := []string{}
			for _, flagName := range FlagNamesPredefined {
				exclude = append(exclude, "--"+flagName)
			}
			args := strings.Split(commandText, " ")
			definitions, err := b.loadAutocompleteDefinitions(args, serviceVersion)
			if err != nil {
				return err
			}
			commands := b.createServiceCommands(definitions)
			command := NewCommand("uipath", "", "").
				WithSubcommands(commands)
			handler := newAutoCompleteHandler()
			words := handler.Find(commandText, command, exclude)
			for _, word := range words {
				fmt.Fprintln(b.StdOut, word)
			}
			return nil
		})
}

func (b CommandBuilder) createAutoCompleteCommand(serviceVersion string) *CommandDefinition {
	flags := NewFlagBuilder().
		AddHelpFlag().
		Build()

	subcommands := []*CommandDefinition{
		b.createAutoCompleteEnableCommand(),
		b.createAutoCompleteCompleteCommand(serviceVersion),
	}

	return NewCommand("autocomplete", "Autocompletion", "Commands for autocompletion").
		WithFlags(flags).
		WithSubcommands(subcommands)
}

func (b CommandBuilder) createConfigCommand() *CommandDefinition {
	const flagNameAuth = "auth"

	flags := NewFlagBuilder().
		AddFlag(NewFlag(flagNameAuth, fmt.Sprintf("Authorization type: %s, %s, %s", CredentialsAuth, LoginAuth, PatAuth), FlagTypeString)).
		AddFlag(NewFlag(FlagNameProfile, "Profile to configure", FlagTypeString).
			WithEnvVarName("UIPATH_PROFILE").
			WithDefaultValue(config.DefaultProfile)).
		AddHelpFlag().
		Build()

	subcommands := []*CommandDefinition{
		b.createConfigSetCommand(),
		b.createCacheCommand(),
	}

	return NewCommand("config", "Interactive Configuration", "Interactive command to configure the CLI").
		WithFlags(flags).
		WithSubcommands(subcommands).
		WithAction(func(context *CommandExecContext) error {
			auth := context.String(flagNameAuth)
			profileName := context.String(FlagNameProfile)
			handler := newConfigCommandHandler(b.StdIn, b.StdOut, b.ConfigProvider)
			return handler.Configure(auth, profileName)
		})
}

func (b CommandBuilder) createConfigSetCommand() *CommandDefinition {
	const flagNameKey = "key"
	const flagNameValue = "value"

	flags := NewFlagBuilder().
		AddFlag(NewFlag(flagNameKey, "The key", FlagTypeString).
			WithRequired(true)).
		AddFlag(NewFlag(flagNameValue, "The value to set", FlagTypeString).
			WithRequired(true)).
		AddFlag(NewFlag(FlagNameProfile, "Profile to configure", FlagTypeString).
			WithEnvVarName("UIPATH_PROFILE").
			WithDefaultValue(config.DefaultProfile)).
		AddHelpFlag().
		Build()

	return NewCommand("set", "Set config parameters", "Set config parameters").
		WithFlags(flags).
		WithAction(func(context *CommandExecContext) error {
			profileName := context.String(FlagNameProfile)
			key := context.String(flagNameKey)
			value := context.String(flagNameValue)
			handler := newConfigCommandHandler(b.StdIn, b.StdOut, b.ConfigProvider)
			return handler.Set(key, value, profileName)
		})
}

func (b CommandBuilder) createCacheClearCommand() *CommandDefinition {
	flags := NewFlagBuilder().
		AddHelpFlag().
		Build()

	return NewCommand("clear", "Clears the cache", "Clears the cache").
		WithFlags(flags).
		WithAction(func(context *CommandExecContext) error {
			handler := newCacheClearCommandHandler(b.StdOut)
			return handler.Clear()
		})
}

func (b CommandBuilder) createCacheCommand() *CommandDefinition {
	flags := NewFlagBuilder().
		AddHelpFlag().
		Build()

	subcommands := []*CommandDefinition{
		b.createCacheClearCommand(),
	}

	return NewCommand("cache", "Caching-related commands", "Caching-related commands").
		WithFlags(flags).
		WithSubcommands(subcommands)
}

func (b CommandBuilder) loadDefinitions(args []string, serviceVersion string) ([]parser.Definition, error) {
	if len(args) <= 1 || strings.HasPrefix(args[1], "-") {
		return b.DefinitionProvider.Index(serviceVersion)
	}
	if len(args) > 1 && args[1] == "commands" {
		return b.loadAllDefinitions(serviceVersion)
	}
	definition, err := b.DefinitionProvider.Load(args[1], serviceVersion)
	if definition == nil {
		return nil, err
	}
	return []parser.Definition{*definition}, err
}

func (b CommandBuilder) loadAllDefinitions(serviceVersion string) ([]parser.Definition, error) {
	all, err := b.DefinitionProvider.Index(serviceVersion)
	if err != nil {
		return nil, err
	}
	definitions := []parser.Definition{}
	for _, d := range all {
		definition, err := b.DefinitionProvider.Load(d.Name, serviceVersion)
		if err != nil {
			return nil, err
		}
		if definition != nil {
			definitions = append(definitions, *definition)
		}
	}
	return definitions, nil
}

func (b CommandBuilder) loadAutocompleteDefinitions(args []string, serviceVersion string) ([]parser.Definition, error) {
	if len(args) <= 2 {
		return b.DefinitionProvider.Index(serviceVersion)
	}
	return b.loadDefinitions(args, serviceVersion)
}

func (b CommandBuilder) createShowCommand(definitions []parser.Definition) *CommandDefinition {
	flags := NewFlagBuilder().
		AddHelpFlag().
		Build()

	return NewCommand("show", "Print CLI commands", "Print available uipath CLI commands").
		WithFlags(flags).
		WithHidden(true).
		WithAction(func(context *CommandExecContext) error {
			defaultFlags := NewFlagBuilder().
				AddDefaultFlags(false).
				AddHelpFlag().
				Build()

			handler := newShowCommandHandler()
			output, err := handler.Execute(definitions, defaultFlags)
			if err != nil {
				return err
			}
			fmt.Fprintln(b.StdOut, output)
			return nil
		})
}

func (b CommandBuilder) createInspectCommand(definitions []parser.Definition) *CommandDefinition {
	flags := NewFlagBuilder().
		AddHelpFlag().
		Build()

	subcommands := []*CommandDefinition{
		b.createShowCommand(definitions),
	}

	return NewCommand("commands", "Inspect available CLI operations", "Command to inspect available uipath CLI operations").
		WithFlags(flags).
		WithSubcommands(subcommands).
		WithHidden(true)
}

func (b CommandBuilder) createServiceCommands(definitions []parser.Definition) []*CommandDefinition {
	commands := []*CommandDefinition{}
	for _, e := range definitions {
		command := b.createServiceCommand(e)
		commands = append(commands, command)
	}
	return commands
}

func (b CommandBuilder) parseArgument(args []string, name string) string {
	for i, arg := range args {
		if strings.TrimSpace(arg) == "--"+name {
			if len(args) > i+1 {
				return strings.TrimSpace(args[i+1])
			}
		}
	}
	return ""
}

func (b CommandBuilder) serviceVersionFromProfile(profile string) string {
	config := b.ConfigProvider.Config(profile)
	if config == nil {
		return ""
	}
	return config.ServiceVersion
}

func (b CommandBuilder) Create(args []string) ([]*CommandDefinition, error) {
	serviceVersion := b.parseArgument(args, FlagNameServiceVersion)
	profile := b.parseArgument(args, FlagNameProfile)
	if serviceVersion == "" && profile != "" {
		serviceVersion = b.serviceVersionFromProfile(profile)
	}
	definitions, err := b.loadDefinitions(args, serviceVersion)
	if err != nil {
		return nil, err
	}
	servicesCommands := b.createServiceCommands(definitions)
	autocompleteCommand := b.createAutoCompleteCommand(serviceVersion)
	configCommand := b.createConfigCommand()
	inspectCommand := b.createInspectCommand(definitions)
	commands := append(servicesCommands, autocompleteCommand, configCommand, inspectCommand)
	return commands, nil
}
