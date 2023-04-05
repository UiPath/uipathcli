# Contributing

We are happy that you would like to contribute to this project. All types of contributions are encouraged and valued. Please make sure to read the relevant section before making your contribution. It will make it a lot easier for us maintainers and smooth out the experience for all involved. We are looking forward to your contributions!

## Questions

If you want to ask a question, we assume that you have read the available [documentation](README.md).

Before you ask a question, it is best to search for existing [issues](https://github.com/UiPath/uipathcli/issues) that might help you. In case you have found a suitable issue and still need clarification, you can write your question in this issue. It is also advisable to search the internet for answers first.

If you then still feel the need to ask a question and need clarification, we recommend the following:

- Open an [issue](https://github.com/UiPath/uipathcli/issues/new).
- Provide as much context as you can about what you're running into.
- Provide project and platform versions, depending on what seems relevant.

We will then take care of the issue as soon as possible.

## Reporting Bugs

A good bug report should not leave others needing to chase you up for more information. Therefore, we ask you to investigate carefully, collect information and describe the issue in detail in your report. Please complete the following steps in advance to help us fix any potential bug as fast as possible.

- Make sure that you are using the latest version.
- To see if other users have experienced (and potentially already solved) the same issue you are having, check if there is not already a bug report in the [issues](https://github.com/UiPath/uipathcli/issues).
- Collect information about the bug:
  - Error message and stack trace
  - OS, Platform and Version (Windows, Linux, macOS, x86, ARM)
  - Version of the interpreter, compiler, SDK, runtime environment, package manager, depending on what seems relevant.
  - Possibly your input and the output

## Code Contributions

### License

When contributing to this project, you must agree that you have authored 100% of the content, that you have the necessary rights to the content and that the content you contribute may be [released](https://help.github.com/articles/github-terms-of-service/#6-contributions-under-repository-license) under the [project license](LICENSE).

### Build

You need the [Go Compiler](https://go.dev/dl/) toolchain to build this project.  You can build an excutable for your current platform using the standard go build command:

### Windows
```powershell
go build -o uipath.exe

.\uipath.exe --help
```

### Linux
```bash
go build -o uipath

./uipath --help
```

### Test

The following command runs the tests with detailed debug output:

```bash
go test -v ./...
```

### Submit a pull request

1. [Fork](https://github.com/UiPath/uipathcli/fork) the repository
2. Commit your change
3. Push your change to your fork and [submit a pull request](https://github.com/UiPath/uipathcli/pulls)
4. Wait for your pull request to be reviewed and merged by the [CODEOWNERS](CODEOWNERS)

Here are a few things you can do to make code reviews go as smooth as possible:

- Keep your change focused. If you want to make multiple changes, submit them as separate pull requests.
- Write a good PR description and commit message to explain the intent of the change
- For bigger features, start a design discussion upfront to avoid large changes or refactorings during the code review phase

## FAQ

### How to integrate my service with the CLI?

If you just want to try out a service, you can copy the OpenAPI specfication in the `definition/` folder next to the `uipath` executable. The CLI automatically picks up all the OpenAPI definitions and shows them as top-level commands. The file name of the definition will be the command name. You can see all installed definitions by running:

```bash
uipath --help
```

If you would like to publish your service definition bundled together with the CLI, you can simply copy it in the [definitions](definitions) folder of this repository and create a PR with your change.

### How to cross-compile the CLI?

You can also cross-compile the CLI using the PowerShell script (`build.ps1`) on Windows and the Bash script (`build.sh`) on Linux:

```powershell
# Cross-compile the CLI on Windows for all supported platforms
.\build.ps1

# Generates
# - build/uipath-linux-amd64          for Linux   (x86-64)
# - build/uipath-windows-amd64.exe    for Windows (x86-64)
# - build/uipath-darwin-amd64         for MacOS   (x86-64)
# - build/uipath-linux-arm64          for Linux   (ARM)
# - build/uipath-windows-arm64.exe    for Windows (ARM)
# - build/uipath-darwin-arm64         for MacOS   (ARM)

# Run the CLI (on windows)
.\build\uipath-windows-amd64.exe --help
```

```bash
# Cross-compile the CLI on Linux for all supported platforms
./build.sh

# Run the CLI (on linux)
./build/uipath-linux-amd64 --help
```

### How to generate code coverage?

```bash
# Run tests and generate code coverage file
go test -v ./... -coverpkg ./... -coverprofile coverage.out

# Visualize coverage file
go tool cover --html=coverage.out
```

### How to create complex commands or how to adapt existing commands?

The CLI supports a pluggable infrastructure which allows you to implement complex custom commands. The [`DigitizeCommand`](plugin/digitizer/digitize_command.go) is an example for a custom command which abstracts away the complexity of the async digitization API. The digitizer API requires you to upload a file, followed by polling the status API until the digitization finished in order to retrieve the digitization result. The `DigitizeCommand` allows the user to just invoke one command for uploading the file and retrieving the result:

```bash
uipath du digitizer digitize --file documents/invoice.pdf
```

returns
```json
{
  "status": "Succeeded",
  "result": {
    "documentId": "c684a468-41fe-4b4f-a99b-7bbd3242137e",
    "contentType": "application/pdf",
    "length": 1839,
    "pages": [
      {
        ...
      }
    ]
  }
}
```

**Implement new command**

These are the steps to implement new command or override existing one:

1. Create a struct which implements the [`CommandPlugin`](plugin/command_plugin.go) interface

```go
type CreateCommand struct{}

func (c CreateCommand) Command() plugin.Command {
  // Provides the definition of the command like name, description and parameters
}

func (c CreateCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
  // Invoked when the CLI command is executed
}
```

2. Implement the `Command()` function and return the definition of your command and all the parameters. In this example we define the `create-product` command with three parameters `id`, `name` and `description`. The `id` is an integer and is required. `name` is a string and required, too. `description` is an optional string.

```go
func (c CreateCommand) Command() plugin.Command {
  return *plugin.NewCommand("myservice").
      WithOperation("create-product", "Creates a product").
      WithParameter("id", plugin.ParameterTypeInteger, "The product id", true).
      WithParameter("name", plugin.ParameterTypeString, "The product name", true).
      WithParameter("description", plugin.ParameterTypeString, "The product description", false)
}
```

3. Implement `Execute(...)` function which performs the operation of the command. The `context` parameters gives you all the input provided by the user and `writer` can be used to write information on standard output and `logger` for debug output.

4. Register the command with the `uipath` CLI in [`main.go`](main.go)

```go
cli := commandline.Cli{
  CommandPlugins: []plugin.CommandPlugin{
    plugin_myservice.CreateCommand{},
  },
}
```

4. You can call your new command:

```bash
uipath myservice create-product --id "1" --name "tv" --description "40 inch Smart TV"

uipath myservice create-product --id "2" --name "table"
```

**Hide existing command**

You can also hide an existing command by setting the hidden flag:

1. Create a struct which implements the [`CommandPlugin`](plugin/command_plugin.go) interface

```go
type StatusCommand struct{}

func (c StatusCommand) Command() plugin.Command {
  return *plugin.NewCommand("myservice").
      WithOperation("status", "").
      IsHidden()
}

func (c StatusCommand) Execute(context plugin.ExecutionContext, writer output.OutputWriter, logger log.Logger) error {
  return fmt.Errorf("Status command not supported")
}
```

2. Register the command with the `uipath` CLI in [`main.go`](main.go)

```go
cli := commandline.Cli{
  CommandPlugins: []plugin.CommandPlugin{
    plugin_myservice.StatusCommand{},
  },
}
```