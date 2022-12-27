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
go build .

.\uipathcli.exe --help
```

### Linux
```bash
go build .

./uipathcli --help
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

### How to cross-compile the CLI?

You can also cross-compile the CLI using the PowerShell script (`build.ps1`) on Windows and the Bash script (`build.sh`) on Linux:

```powershell
# Cross-compile the CLI on Windows for all supported platforms
.\build.ps1

# Generates
# - build/uipathcli        for Linux
# - build/uipathcli.exe    for Windows
# - build/uipathcli.osx    for MacOS

# Run the CLI (on windows)
.\build\uipathcli.exe --help
```

```bash
# Cross-compile the CLI on Linux for all supported platforms
./build.sh

# Generates
# - build/uipathcli        for Linux
# - build/uipathcli.exe    for Windows
# - build/uipathcli.osx    for MacOS

# Run the CLI (on linux)
./build/uipathcli --help
```

### How to generate code coverage?

```bash
# Run tests and generate code coverage file
go test -v ./... -coverpkg ./... -coverprofile coverage.out

# Visualize coverage file
go tool cover --html=coverage.out
```
