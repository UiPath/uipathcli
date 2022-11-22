# UiPath OpenAPI Command-Line-Interface

The UiPath OpenAPI CLI project is a command line interface to simplify, script and automate API calls for UiPath services.

CLI operations and arguments are generated based OpenAPI 3 documents. OpenAPI documents for new services can be dropped in the `definitions` folder and will be automatically picked up and displayed in the CLI.

Executuables are available for Windows, Linux and MacOS.

## Usage

For more details about how to use the CLI, take a look at the [Getting Started](GETTING_STARTED.md) guide.

## Prerequisites

- [Go Compiler](https://go.dev/dl/)

## Build

You can build an excutable for your current platform using the standard go build command:

```
# Build the CLI on Windows
go build .
.\uipathcli.exe --help
```

```
# Build the CLI on Linux
go build .
./uipathcli --help
```

## Test

The following command runs the tests with detailed debug output:

```
go test -v ./...
```

## FAQ

### How to cross-compile the CLI?

You can also cross-compile the CLI using the PowerShell script (`build.ps1`) on Windows and the Bash script (`build.sh`) on Linux:

```
# Cross-compile the CLI on Windows for all supported platforms
.\build.ps1

# Generates
# - build/uipathcli        for Linux
# - build/uipathcli.exe    for Windows
# - build/uipathcli.osx    for MacOS

# Run the CLI (on windows)
.\build\uipathcli.exe --help
```

```
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

```
# Run tests and generate code coverage file
go test -v ./... -coverpkg ./... -coverprofile coverage.out

# Visualize coverage file
go tool cover --html=coverage.out
```

### How to run the CLI against service fabric?

You can set up a separate profile for service fabric which configures the URI and disables HTTPS certificate checks: 

```
profiles:
  - name: sf
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    path:
      organization: test
      tenant: DefaultTenant
    insecure: true
```

And you simply call the CLI with the `--profile sf` parameter:

```
./uipathcli metering ping --profile sf
```

### How to retrieve secrets from kubernetes?

The CLI has support for retrieving clientId and clientSecret from kubernetes using the [uipathcli-authenticator-k8s](https://github.com/UiPath/uipathcli-authenticator-k8s). You need to enable the authenticator plugin by creating the plugins configuration file `.uipathcli/plugins` in your home directory:

```
authenticators:
  - name: kubernetes
    path: ./uipathcli-authenticator-k8s
```

You can define the secret name, namespace and data keys so that the CLI fetches the clientId and clientSecret using the kube API and creates a bearer token based on these credentials:

```
profiles:
  - name: default
    auth:
      type: kubernetes
      kubeconfig: /home/tschmitt/.kube/config
      namespace: <my-namespace>
      secretName: <my-secret>
      clientId: ClientId          # data key in <my-secret>
      clientSecret: ClientSecret  # data key in <my-secret>
    path:
      organization: uipatcleitzc
      tenant: DefaultTenant
```
