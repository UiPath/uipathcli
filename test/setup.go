package commandline

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/cache"
	"github.com/UiPath/uipathcli/commandline"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/executor"
	"github.com/UiPath/uipathcli/parser"
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

func (b *ContextBuilder) WithResponse(statusCode int, body string) *ContextBuilder {
	b.context.ResponseStatus = statusCode
	b.context.ResponseBody = body
	return b
}

func (b *ContextBuilder) WithResponseHeader(header map[string]string) *ContextBuilder {
	b.context.ResponseHeader = header
	return b
}

func (b *ContextBuilder) Build() Context {
	return b.context
}

type Context struct {
	Config         string
	DefinitionName string
	DefinitionData string
	ResponseStatus int
	ResponseHeader map[string]string
	ResponseBody   string
}

type Result struct {
	Error         error
	StdOut        string
	StdErr        string
	RequestUrl    string
	RequestHeader map[string]string
	RequestBody   string
}

func runCli(args []string, context Context) Result {
	requestUrl := ""
	requestHeader := map[string]string{}
	requestBody := ""

	if context.ResponseStatus != 0 {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cli := commandline.Cli{
		StdOut:         stdout,
		StdErr:         stderr,
		Parser:         parser.OpenApiParser{},
		ConfigProvider: config.ConfigProvider{},
		Executor: executor.HttpExecutor{
			Authenticators: []auth.Authenticator{
				auth.TokenAuthenticator{
					Cache: cache.FileCache{},
				},
			},
		},
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
