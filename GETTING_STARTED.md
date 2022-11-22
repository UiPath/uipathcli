# UiPath OpenAPI Command-Line-Interface

The UiPath OpenAPI CLI project is a command line interface to simplify, script and automate API calls for UiPath services.

The CLI works on Windows, Linux and MacOS.

## Install

In order to get started quickly, you can run the install scripts for windows and linux:

```
# Install uipathcli.exe on Windows

. { iwr https://du-nst-cdn.azureedge.net/uipathcli/install.ps1 } | iex

.\uipathcli.exe --help
```

```
# Install uipathcli on Linux and MacOS

curl -sL https://du-nst-cdn.azureedge.net/uipathcli/install.sh | bash

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

## Multiple Profiles

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

## Global Arguments

You can either pass global arguments as CLI parameters, set an env variable or set them using the configuration file. Here is a list of the supported global arguments which can be applied to all CLI operations:

| Name | Env-Variable | Type | Default Value | Description |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| `--debug` | `UIPATH_DEBUG` | `boolean` | `false` | Show debug output |
| `--profile` | `UIPATH_PROFILE` | `string` | `default` | Use profile from configuration file |
| `--uri` | `UIPATH_URI` | `uri` | `https://cloud.uipath.com` | URL override |
| `--insecure` | `UIPATH_INSECURE` | `boolean` | `false` |*Warning: Disables HTTPS certificate checks* |
