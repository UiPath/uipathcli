# UiPath OpenAPI Command-Line-Interface

The UiPath OpenAPI CLI project is a simple command line interface to simplify, script and automate API calls for UiPath services.

CLI operations and arguments are generated based OpenAPI 3 documents. OpenAPI documents for new services can be dropped in the `definitions` folder and will be automatically picked up and displayed in the CLI.

Executuables are available for Windows, Linux and MacOS.

## Prerequisites

- [Go Compiler](https://go.dev/dl/)

## Install

In order to quickly get started, you can run the install scripts for windows and linux:

```
# Install uipathcli.exe on Windows
.\install.ps1
.\uipathcli.exe --help
```

```
# Install uipathcli on Linux
./install.sh
./uipathcli --help
```

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

## Configuration

Create a configuration file `.uipathcli/config` in your home directory. The following config file sets up the default profile with clientId, clientSecret so that the CLI can generate a bearer token before calling any of the services. It also sets the organization and tenant in the route for services which require it.

```
cat <<EOT > $HOME/.uipathcli/config
---
profiles:
  - name: default
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    path:
      organization: <organization-name>
      tenant: <tenant-name>
EOT
```

Once you have created the configuration file with the proper secrets, org and tenant information, you should be able to successfully call the services, e.g.

```
./uipathcli metering ping
```

## Commands and arguments

CLI commands consist of three main parts:

```
./uipathcli <service-name> <operation-name> <arguments>
```

- `<service-name>`: The CLI discovers the existing OpenAPI specifications and shows each of them as a separate service
- `<operation-name>`: The operation typically represents the route to call
- `<arguments>`: A list of arguments which are used as request parameters (in the path, header, querystring or body)

Example:

```
./uipathcli metering validate --product-name "DU" --model-name "my-model"
```

### Basic arguments

The CLI supports string, integer, floating point and boolean arguments. The arguments are automatically converted to the type defined in the OpenAPI specification:

```
./uipathcli product create --name "new-product" --stock "5" --price "1.4" --deleted "false"
```

### Array arguments

Array arguments can be passed as comma-separated strings and are automatically converted to arrays in the JSON body. The CLI supports string, integer, floating point, boolean and object arrays.

```
./uipathcli product list --name-filter "my-product,new-product"
```

### Nested Object arguments

More complex nested objects can be passed as semi-colon separated list of property assigments:

```
./uipathcli product create --product "name=my-product;price.value=340;price.sale.discount=10;price.sale.value=306"
```

The command creates the following JSON body in the HTTP request:

```
{
  "product": {
      "name": "my-product",
      "price": {
          "value": 340,
          "sale": {
              "discount": 10,
              "value": 306
          }
      }
  }
}
```
### File Upload arguments

File content can be uploaded directly from a command line argument. The following command will upload a file with the content `hello-world`:

```
./uipathcli digitizer digitize --api-version 1 --file "hello-world"
```

CLI arguments can also refer to files on disk. This command reads the invoice `C:\documents\Invoice.pdf` and uploads it to the digitize endpoint:

```
./uipathcli digitizer digitize --api-version 1 --file file://C:\documents\Invoice.pdf
```

## Debug

You can set the environment variable `UIPATH_DEBUG=true` or pass the parameter `--debug` in order to see detailed output of the request and response messages:

```
./uipathcli metering ping --debug
```

```
GET https://cloud.uipath.com/uipatcleitzc/DefaultTenant/du_/api/metering/ping HTTP/1.1
X-Request-Id: b033e39294147bcb1174c5b7ace6ac7c
Authorization: Bearer ...


HTTP/1.1 200 OK
Connection: keep-alive
Content-Type: application/json; charset=utf-8

{
  "location": "westeurope",
  "serverRegion": "westeurope",
  "clusterId": "du-prod-du-we-g-dns",
  "version": "22.8-63-main.v29c916",
  "timestamp": "2022-08-23T12:23:19.0121688Z"
}
```

## Service Fabric

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

### Global Arguments

| Name | Env-Variable |Type | Default Value | Description |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| `--debug` | `UIPATH_DEBUG` | `boolean` | `false` | Show debug output |
| `--profile` | `UIPATH_PROFILE` | `string` | `default` | Use profile from configuration file |
| `--uri` | `UIPATH_URI` | `uri` | `https://cloud.uipath.com` | URL override |
| `--insecure` | `UIPATH_INSECURE` | `boolean` | `false` |*Warning: Disables HTTPS certificate checks* |

## FAQ

### How to run the tests?

The following command runs the tests with detailed debug output:

```
go test -v ./...
```

### How to generate code coverage?

```
# Run tests and generate code coverage file
go test -v ./... -coverpkg ./... -coverprofile coverage.out

# Visualize coverage file
go tool cover --html=coverage.out
```

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

### How to use multiple profiles?

You can also define multiple configuration profiles to target different environments (like alpha, staging or prod), configure separate auth credentials, or manage multiple organizations/tenants:

```
profiles:
  - name: default
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    path:
      organization: uipatcleitzc
      tenant: DefaultTenant
  - name: apikey
    uri: https://du.uipath.com/metering/
    header:
      X-UIPATH-License: <your-api-key>
      X-UIPATH-MLService: MLSERVICE_TIEMODEL
  - name: alpha
    uri: https://alpha.uipath.com
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    path:
      organization: UiPatricjvjx
      tenant: DefaultTenant
```

If you do not provide the `--profile` parameter, the `default` profile is automatically selected. Otherwise it will use the settings from the provided profile. The following command will send a request to the alpha.uipath.com environment:

```
./uipathcli metering ping --profile alpha
```

You can also change the profile using an environment variable (`UIPATH_PROFILE`):

```
UIPATH_PROFILE=alpha ./uipathcli metering ping
```
