package commandline

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"sync"

	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/urfave/cli/v2"
)

const insecureFlagName = "insecure"
const debugFlagName = "debug"
const profileFlagName = "profile"
const uriFlagName = "uri"
const helpFlagName = "help"

type CommandBuilder struct {
	StdIn          io.Reader
	StdOut         io.Writer
	ConfigProvider config.ConfigProvider
	Executor       executor.Executor
	PluginExecutor executor.Executor
}

func (b CommandBuilder) createExecutionParameters(context *cli.Context, in string, operation parser.Operation, additionalParameters map[string]string) ([]executor.ExecutionParameter, error) {
	typeConverter := TypeConverter{}

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

func (b CommandBuilder) parameterDescription(parameter parser.Parameter) string {
	required := b.parameterRequired(parameter)
	if parameter.DefaultValue != nil {
		return fmt.Sprintf("%s (default: %v)", parameter.Description, parameter.DefaultValue)
	}
	if required {
		return fmt.Sprintf("%s (required)", parameter.Description)
	}
	return parameter.Description
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

func (b CommandBuilder) overrideUri(uri *url.URL, overrideUri *url.URL, config config.Config) (*url.URL, error) {
	scheme := uri.Scheme
	host := uri.Host
	path := uri.Path

	if overrideUri != nil && overrideUri.Scheme != "" {
		scheme = overrideUri.Scheme
	}
	if overrideUri != nil && overrideUri.Host != "" {
		host = overrideUri.Host
	}
	if overrideUri != nil && overrideUri.Path != "" {
		path = overrideUri.Path
	}
	normalizedPath := strings.Trim(path, "/")
	return url.Parse(fmt.Sprintf("%s://%s/%s", scheme, host, normalizedPath))
}

func (b CommandBuilder) createBaseUri(definition parser.Definition, config config.Config, context *cli.Context) (*url.URL, error) {
	uriArgument, err := b.parseUriArgument(context)
	if err != nil {
		return nil, err
	}

	uri := &definition.BaseUri
	uri, err = b.overrideUri(uri, config.Uri, config)
	if err != nil {
		return nil, err
	}
	uri, err = b.overrideUri(uri, uriArgument, config)
	if err != nil {
		return nil, err
	}
	return uri, nil
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

func (b CommandBuilder) validateArguments(context *cli.Context, parameters []parser.Parameter, config config.Config) error {
	err := errors.New("Invalid arguments:")
	result := true
	for _, parameter := range parameters {
		if parameter.Required &&
			parameter.DefaultValue == nil &&
			context.String(parameter.Name) == "" &&
			config.Path[parameter.Name] == "" &&
			config.Query[parameter.Name] == "" &&
			config.Header[parameter.Name] == "" {
			result = false
			err = fmt.Errorf("%w\n  Argument --%s is missing", err, parameter.Name)
		}
	}
	if result {
		return nil
	}
	return err
}

func (b CommandBuilder) executeCommand(context executor.ExecutionContext, output io.Writer) error {
	if context.Plugin != nil {
		return b.PluginExecutor.Call(context, output)
	}
	return b.Executor.Call(context, output)
}

func (b CommandBuilder) createOperationCommand(definition parser.Definition, operation parser.Operation) *cli.Command {
	flags := b.CreateDefaultFlags(true)
	flags = append(flags, b.HelpFlag())
	flags = append(flags, b.createFlags(operation.Parameters)...)

	return &cli.Command{
		Name:  operation.Name,
		Usage: operation.Description,
		Flags: flags,
		Action: func(context *cli.Context) error {
			profileName := context.String(profileFlagName)
			config := b.ConfigProvider.Config(profileName)
			if config == nil {
				return fmt.Errorf("Could not find profile '%s'", profileName)
			}

			baseUri, err := b.createBaseUri(definition, *config, context)
			if err != nil {
				return err
			}

			err = b.validateArguments(context, operation.Parameters, *config)
			if err != nil {
				return err
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

			insecure := context.Bool(insecureFlagName) || config.Insecure
			debug := context.Bool(debugFlagName) || config.Debug
			executionContext := executor.NewExecutionContext(
				operation.Method,
				*baseUri,
				operation.Route,
				pathParameters,
				queryParameters,
				headerParameters,
				bodyParameters,
				formParameters,
				config.Auth,
				insecure,
				debug,
				operation.Plugin)

			var wg sync.WaitGroup
			wg.Add(2)
			reader, writer := io.Pipe()
			go func(reader *io.PipeReader) {
				defer wg.Done()
				io.Copy(b.StdOut, reader)
			}(reader)

			go func(context executor.ExecutionContext, output io.WriteCloser) {
				defer wg.Done()
				defer output.Close()
				err = b.executeCommand(context, output)
			}(*executionContext, writer)

			wg.Wait()
			return err
		},
		HideHelp: true,
		Hidden:   operation.Hidden,
	}
}

func (b CommandBuilder) createServiceCommand(definition parser.Definition) *cli.Command {
	commands := []*cli.Command{}
	for _, operations := range definition.Operations {
		commands = append(commands, b.createOperationCommand(definition, operations))
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
		Hidden: true,
		Action: func(context *cli.Context) error {
			shell := context.String(shellFlagName)
			filePath := context.String(fileFlagName)
			handler := AutoCompleteHandler{}
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

func (b CommandBuilder) createAutoCompleteCompleteCommand(commands []*cli.Command) *cli.Command {
	return &cli.Command{
		Name:        "complete",
		Description: "Returns the autocomplete suggestions",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:   "command",
				Usage:  "The command to autocomplete",
				Hidden: true,
			},
			b.HelpFlag(),
		},
		Hidden: true,
		Action: func(context *cli.Context) error {
			commandText := context.String("command")
			exclude := []string{
				"--" + insecureFlagName,
				"--" + debugFlagName,
				"--" + profileFlagName,
				"--" + uriFlagName,
				"--" + helpFlagName,
			}
			handler := AutoCompleteHandler{}
			words := handler.Find(commandText, commands, exclude)
			for _, word := range words {
				fmt.Fprintln(b.StdOut, word)
			}
			return nil
		},
		HideHelp: true,
	}
}

func (b CommandBuilder) createAutoCompleteCommand(commands []*cli.Command) *cli.Command {
	return &cli.Command{
		Name: "autocomplete",
		Subcommands: []*cli.Command{
			b.createAutoCompleteEnableCommand(),
			b.createAutoCompleteCompleteCommand(commands),
		},
	}
}

func (b CommandBuilder) createConfigCommand() *cli.Command {
	authFlagName := "auth"
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:  authFlagName,
			Value: CredentialsAuth,
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
		Hidden:      true,
		Flags:       flags,
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

func (b CommandBuilder) Create(definitions []parser.Definition) []*cli.Command {
	commands := []*cli.Command{}
	for _, e := range definitions {
		command := b.createServiceCommand(e)
		commands = append(commands, command)
	}
	autocompleteCommand := b.createAutoCompleteCommand(commands)
	configCommand := b.createConfigCommand()
	return append(commands, autocompleteCommand, configCommand)
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
		&cli.BoolFlag{
			Name:    insecureFlagName,
			Usage:   "Disable HTTPS certificate check",
			EnvVars: []string{"UIPATH_INSECURE"},
			Value:   false,
			Hidden:  hidden,
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
