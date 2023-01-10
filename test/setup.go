package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
)

type ContextBuilder struct {
	context Context
}

func NewContextBuilder() *ContextBuilder {
	return &ContextBuilder{
		context: Context{},
	}
}

func (b *ContextBuilder) WithDefinition(name string, data string) *ContextBuilder {
	b.context.DefinitionName = name
	b.context.DefinitionData = data
	return b
}

func (b *ContextBuilder) WithConfig(config string) *ContextBuilder {
	b.context.Config = config
	return b
}

func (b *ContextBuilder) WithConfigFile(configFile string) *ContextBuilder {
	b.context.ConfigFile = configFile
	return b
}

func (b *ContextBuilder) WithStdIn(input bytes.Buffer) *ContextBuilder {
	b.context.StdIn = &input
	return b
}

func (b *ContextBuilder) WithResponse(statusCode int, body string) *ContextBuilder {
	b.context.ResponseStatus = statusCode
	b.context.ResponseBody = body
	return b
}

func (b *ContextBuilder) WithIdentityResponse(statusCode int, body string) *ContextBuilder {
	b.context.Identity = IdentityContext{statusCode, body}
	return b
}

func (b *ContextBuilder) WithResponseHeader(header map[string]string) *ContextBuilder {
	b.context.ResponseHeader = header
	return b
}

func (b *ContextBuilder) WithCommandPlugin(commandPlugin plugin.CommandPlugin) *ContextBuilder {
	b.context.CommandPlugin = commandPlugin
	return b
}

func (b *ContextBuilder) Build() Context {
	return b.context
}

type IdentityContext struct {
	ResponseStatus int
	ResponseBody   string
}

type Context struct {
	Config         string
	ConfigFile     string
	StdIn          *bytes.Buffer
	DefinitionName string
	DefinitionData string
	ResponseStatus int
	ResponseHeader map[string]string
	ResponseBody   string
	Identity       IdentityContext
	CommandPlugin  plugin.CommandPlugin
}

type Result struct {
	Error         error
	StdOut        string
	StdErr        string
	RequestUrl    string
	RequestHeader map[string]string
	RequestBody   string
}

func handleIdentityTokenRequest(context Context, request *http.Request, response http.ResponseWriter) {
	body, _ := io.ReadAll(request.Body)
	requestBody := string(body)
	data, _ := url.ParseQuery(requestBody)
	if len(data["client_id"]) != 1 || data["client_id"][0] == "" {
		response.WriteHeader(400)
		response.Write([]byte("client_id is missing"))
		return
	}
	if len(data["client_secret"]) != 1 || data["client_secret"][0] == "" {
		response.WriteHeader(400)
		response.Write([]byte("client_secret is missing"))
		return
	}
	if len(data["grant_type"]) != 1 || data["grant_type"][0] != "client_credentials" {
		response.WriteHeader(400)
		response.Write([]byte("Invalid grant_type"))
		return
	}
	response.WriteHeader(context.Identity.ResponseStatus)
	response.Write([]byte(context.Identity.ResponseBody))
}

func runCli(args []string, context Context) Result {
	requestUrl := ""
	requestHeader := map[string]string{}
	requestBody := ""

	if context.ResponseStatus != 0 {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.String() == "/identity_/connect/token" {
				handleIdentityTokenRequest(context, r, w)
				return
			}

			requestUrl = r.URL.String()
			body, _ := io.ReadAll(r.Body)
			requestBody = string(body)
			for key, values := range r.Header {
				for _, value := range values {
					requestHeader[strings.ToLower(key)] = value
				}
			}

			w.WriteHeader(context.ResponseStatus)
			for key, value := range context.ResponseHeader {
				w.Header().Add(strings.ToLower(key), value)
			}
			w.Write([]byte(context.ResponseBody))
		}))
		defer srv.Close()
		args = append(args, "--uri", srv.URL)
	}

	if context.ConfigFile != "" && context.Config != "" {
		os.WriteFile(context.ConfigFile, []byte(context.Config), 0600)
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	authenticators := []auth.Authenticator{
		auth.PatAuthenticator{},
		auth.OAuthAuthenticator{
			Cache: cache.FileCache{},
		},
		auth.BearerAuthenticator{
			Cache: cache.FileCache{},
		},
	}
	commandPlugins := []plugin.CommandPlugin{}
	if context.CommandPlugin != nil {
		commandPlugins = append(commandPlugins, context.CommandPlugin)
	}

	cli := commandline.Cli{
		StdIn:  context.StdIn,
		StdOut: stdout,
		StdErr: stderr,
		Parser: parser.OpenApiParser{},
		ConfigProvider: config.ConfigProvider{
			ConfigFileName: context.ConfigFile,
		},
		Executor: executor.HttpExecutor{
			Authenticators: authenticators,
		},
		PluginExecutor: executor.PluginExecutor{
			Authenticators: authenticators,
		},
		CommandPlugins: commandPlugins,
	}
	data := []commandline.DefinitionData{
		*commandline.NewDefinitionData(context.DefinitionName, []byte(context.DefinitionData)),
	}
	args = append([]string{"uipath"}, args...)
	err := cli.Run(args, []byte(context.Config), data)

	return Result{
		Error:         err,
		StdOut:        stdout.String(),
		StdErr:        stderr.String(),
		RequestUrl:    requestUrl,
		RequestHeader: requestHeader,
		RequestBody:   requestBody,
	}
}
