// Package test provides shared test utilities for writing integration tests.
package test

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/stream"
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
	definitionData := commandline.NewDefinitionData(name, "", []byte(data))
	b.context.Definitions = append(b.context.Definitions, *definitionData)
	return b
}

func (b *ContextBuilder) WithDefinitionVersion(name string, serviceVersion string, data string) *ContextBuilder {
	definitionData := commandline.NewDefinitionData(name, serviceVersion, []byte(data))
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

func (b *ContextBuilder) WithResponseHandler(handler func(RequestData) ResponseData) *ContextBuilder {
	b.context.ResponseHandler = handler
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

type RequestData struct {
	URL    url.URL
	Header map[string]string
	Body   []byte
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
	ResponseHandler  func(RequestData) ResponseData
	IdentityResponse ResponseData
	CommandPlugin    plugin.CommandPlugin
}

type Result struct {
	Error         error
	StdOut        string
	StdErr        string
	BaseUrl       string
	RequestUrl    string
	RequestHeader map[string]string
	RequestBody   string
}

func handleIdentityTokenRequest(context Context, request *http.Request, response http.ResponseWriter) {
	body, _ := io.ReadAll(request.Body)
	requestBody := string(body)
	data, _ := url.ParseQuery(requestBody)
	if len(data["client_id"]) != 1 || data["client_id"][0] == "" {
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte("client_id is missing"))
		return
	}
	if len(data["client_secret"]) != 1 || data["client_secret"][0] == "" {
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte("client_secret is missing"))
		return
	}
	if len(data["grant_type"]) != 1 || data["grant_type"][0] != "client_credentials" {
		response.WriteHeader(http.StatusBadRequest)
		_, _ = response.Write([]byte("Invalid grant_type"))
		return
	}
	response.WriteHeader(context.IdentityResponse.Status)
	_, _ = response.Write([]byte(context.IdentityResponse.Body))
}

func RunCli(args []string, context Context) Result {
	baseUrl := ""
	requestUrl := ""
	requestHeader := map[string]string{}
	requestBody := ""

	if len(context.Responses) > 0 || context.ResponseHandler != nil {
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

			decodedRequestUrl := r.URL.Path
			query, _ := url.QueryUnescape(r.URL.RawQuery)
			if query != "" {
				decodedRequestUrl += "?" + query
			}
			response, found := context.Responses[decodedRequestUrl]
			if !found {
				response, found = context.Responses["*"]
			}
			if found {
				w.WriteHeader(response.Status)
				_, _ = w.Write([]byte(response.Body))
				return
			}

			if context.ResponseHandler != nil {
				response = context.ResponseHandler(RequestData{
					URL:    *r.URL,
					Header: requestHeader,
					Body:   body,
				})
				w.WriteHeader(response.Status)
				_, _ = w.Write([]byte(response.Body))
				return
			}
			panic(fmt.Sprintf("Request Url has not been handled '%s'", requestUrl))
		}))
		defer srv.Close()
		args = append(args, "--uri", srv.URL)
		baseUrl = srv.URL
	}

	if context.ConfigFile != "" && context.Config != "" {
		err := os.WriteFile(context.ConfigFile, []byte(context.Config), 0600)
		if err != nil {
			panic(fmt.Errorf("Error writing config file '%s': %w", context.ConfigFile, err))
		}
	}

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	authenticators := []auth.Authenticator{
		auth.NewPatAuthenticator(),
		auth.NewOAuthAuthenticator(cache.NewFileCache(), *auth.NewBrowserLauncher()),
		auth.NewBearerAuthenticator(cache.NewFileCache()),
	}
	commandPlugins := []plugin.CommandPlugin{}
	if context.CommandPlugin != nil {
		commandPlugins = append(commandPlugins, context.CommandPlugin)
	}

	cli := commandline.NewCli(
		context.StdIn,
		stdout,
		stderr,
		false,
		*commandline.NewDefinitionProvider(
			commandline.NewDefinitionFileStoreWithData(context.Definitions),
			parser.NewOpenApiParser(),
			commandPlugins,
		),
		*config.NewConfigProvider(
			config.NewConfigFileStoreWithData(context.ConfigFile, []byte(context.Config)),
		),
		executor.NewHttpExecutor(authenticators),
		executor.NewPluginExecutor(authenticators),
	)
	args = append([]string{"uipath"}, args...)
	var input stream.Stream
	if context.StdIn != nil {
		input = stream.NewMemoryStream(parser.RawBodyParameterName, context.StdIn.Bytes())
	}
	err := cli.Run(args, input)

	return Result{
		Error:         err,
		StdOut:        stdout.String(),
		StdErr:        stderr.String(),
		BaseUrl:       baseUrl,
		RequestUrl:    requestUrl,
		RequestHeader: requestHeader,
		RequestBody:   requestBody,
	}
}

func TempFile(t *testing.T) string {
	directory := t.TempDir()
	return filepath.Join(directory, RandomString(50))
}

func CreateTempFile(t *testing.T, data string) string {
	return CreateTempFileBinary(t, []byte(data))
}

func CreateTempFileBinary(t *testing.T, data []byte) string {
	path := TempFile(t)
	err := os.WriteFile(path, data, 0600)
	if err != nil {
		t.Fatalf("Error writing file '%s': %v", path, err)
	}
	return path
}

func ParseOutput(t *testing.T, output string) map[string]interface{} {
	stdout := map[string]interface{}{}
	err := json.Unmarshal([]byte(output), &stdout)
	if err != nil {
		t.Errorf("Failed to deserialize command output: %v", err)
	}
	return stdout
}

func GetArgumentValue(args []string, name string) string {
	index := slices.Index(args, name)
	if index == -1 {
		return ""
	}
	return args[index+1]
}

func RandomString(length int) string {
	randBytes := make([]byte, length)
	_, err := rand.Read(randBytes)
	if err != nil {
		panic(fmt.Errorf("Error generating random string: %w", err))
	}
	return hex.EncodeToString(randBytes)[:length]
}
