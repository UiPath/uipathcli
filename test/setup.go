package test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

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
		context: Context{
			Definitions: []commandline.DefinitionData{},
			Responses:   map[string]ResponseData{},
		},
	}
}

func (b *ContextBuilder) WithDefinition(name string, data string) *ContextBuilder {
	definitionData := commandline.NewDefinitionData(name, []byte(data))
	b.context.Definitions = append(b.context.Definitions, *definitionData)
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
	b.context.Responses["*"] = ResponseData{statusCode, body}
	return b
}

func (b *ContextBuilder) WithUrlResponse(url string, statusCode int, body string) *ContextBuilder {
	b.context.Responses[url] = ResponseData{statusCode, body}
	return b
}

func (b *ContextBuilder) WithIdentityResponse(statusCode int, body string) *ContextBuilder {
	b.context.IdentityResponse = ResponseData{statusCode, body}
	return b
}

func (b *ContextBuilder) WithCommandPlugin(commandPlugin plugin.CommandPlugin) *ContextBuilder {
	b.context.CommandPlugin = commandPlugin
	return b
}

func (b *ContextBuilder) Build() Context {
	return b.context
}

type ResponseData struct {
	Status int
	Body   string
}

type Context struct {
	Config           string
	ConfigFile       string
	StdIn            *bytes.Buffer
	Definitions      []commandline.DefinitionData
	Responses        map[string]ResponseData
	IdentityResponse ResponseData
	CommandPlugin    plugin.CommandPlugin
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
	response.WriteHeader(context.IdentityResponse.Status)
	response.Write([]byte(context.IdentityResponse.Body))
}

func runCli(args []string, context Context) Result {
	requestUrl := ""
	requestHeader := map[string]string{}
	requestBody := ""

	if len(context.Responses) > 0 {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestUrl = r.URL.String()
			if requestUrl == "/identity_/connect/token" {
				handleIdentityTokenRequest(context, r, w)
				return
			}

			body, _ := io.ReadAll(r.Body)
			requestBody = string(body)
			for key, values := range r.Header {
				for _, value := range values {
					requestHeader[strings.ToLower(key)] = value
				}
			}

			response, found := context.Responses[requestUrl]
			if !found {
				response = context.Responses["*"]
			}
			w.WriteHeader(response.Status)
			w.Write([]byte(response.Body))
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
			Cache:           cache.FileCache{},
			BrowserLauncher: auth.ExecBrowserLauncher{},
		},
		auth.BearerAuthenticator{
			Cache: cache.FileCache{},
		},
	}
	commandPlugins := []plugin.CommandPlugin{}
	if context.CommandPlugin != nil {
		commandPlugins = append(commandPlugins, context.CommandPlugin)
	}

	definitionFiles := []string{}
	for _, data := range context.Definitions {
		definitionFiles = append(definitionFiles, data.Name+".yaml")
	}
	cli := commandline.Cli{
		StdIn:  context.StdIn,
		StdOut: stdout,
		StdErr: stderr,
		DefinitionProvider: commandline.DefinitionProvider{
			DefinitionStore: commandline.DefinitionStore{
				DefinitionFiles: definitionFiles,
				Definitions:     context.Definitions,
			},
			Parser:         parser.OpenApiParser{},
			CommandPlugins: commandPlugins,
		},
		ConfigProvider: config.ConfigProvider{
			ConfigStore: config.ConfigStore{
				Config:     []byte(context.Config),
				ConfigFile: context.ConfigFile,
			},
		},
		Executor: executor.HttpExecutor{
			Authenticators: authenticators,
		},
		PluginExecutor: executor.PluginExecutor{
			Authenticators: authenticators,
		},
	}
	args = append([]string{"uipath"}, args...)
	input := []byte{}
	if context.StdIn != nil {
		input = context.StdIn.Bytes()
	}
	err := cli.Run(args, input)

	return Result{
		Error:         err,
		StdOut:        stdout.String(),
		StdErr:        stderr.String(),
		RequestUrl:    requestUrl,
		RequestHeader: requestHeader,
		RequestBody:   requestBody,
	}
}

func createFile(t *testing.T) string {
	tempFile, err := os.CreateTemp("", "uipathcli-test")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Remove(tempFile.Name()) })
	return tempFile.Name()
}
