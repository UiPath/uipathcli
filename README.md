# UiPath OpenAPI Command-Line-Interface

The UiPath OpenAPI CLI project is a command line interface to simplify, script and automate API calls for UiPath services. The CLI works on Windows, Linux and MacOS.

<img src="https://du-nst-cdn.azureedge.net/uipathcli/getting_started.gif" />

## Install

In order to get started quickly, you can run the install scripts for Windows, Linux and MacOS:

<details open>
  <summary>Install instructions for x86_64/amd64</summary>
  <p>

### Windows

```powershell
Invoke-WebRequest "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli-windows-amd64.zip" -OutFile "uipathcli.zip" ; Expand-Archive -Force -Path "uipathcli.zip" -DestinationPath "."
```

### Linux

```bash
curl --silent "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli-linux-amd64.tar.gz" | tar --extract --gzip --overwrite
```

### MacOS

```bash
curl --silent "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli-darwin-amd64.tar.gz" | tar --extract --gzip --overwrite
```

  </p>
</details>

<details>
  <summary>Install instructions for arm64</summary>
  <p>

### Windows (ARM)

```powershell
Invoke-WebRequest "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli-windows-arm64.zip" -OutFile "uipathcli.zip" ; Expand-Archive -Force -Path "uipathcli.zip" -DestinationPath "."
```

### Linux (ARM)

```bash
curl --silent "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli-linux-arm64.tar.gz" | tar --extract --gzip --overwrite
```

### MacOS (ARM)

```bash
curl --silent "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli-darwin-arm64.tar.gz" | tar --extract --gzip --overwrite
```

  </p>
</details>

<details>
  <summary>Enable command completion</summary>
  <p>

For autocompletion to work, the `uipathcli` executable needs to be in your PATH. Make sure the following commands output the path to the `uipathcli` executable:

### PowerShell

```powershell
(Get-Command uipathcli).Path
```

### Bash

```bash
which uipathcli
```

You can enable autocompletion by running the following commands depending on which shell you are using:

### PowerShell

```powershell
uipathcli autocomplete enable --shell "powershell"
```

### Bash

```bash
uipathcli autocomplete enable --shell "bash"
```

  </p>
</details>

<br />

After installing the `uipathcli` executable, you can run the interactive config command to finish setting up your CLI:

```
uipathcli config
```

More details about how to configure the CLI can be found in the following sections.

## Configuration

The CLI supports multiple ways to authorize with the UiPath services:
- **Client Credentials**: Generate secret and configure the CLI to use these long-term credentials.

- **OAuth Login**: Login to UiPath using your browser and SSO of choice: Microsoft Login, Google Login, LinkedIn, Custom SSO or simple username/password. No need to manage any credentials.

- **Personal Access Token**: Generate a PAT and configure the CLI to use the access token.

### Client Credentials

In order to use client credentials, you need to set up an [External Application (Confidential)](https://docs.uipath.com/automation-cloud/docs/managing-external-applications) and generate an [application secret](https://docs.uipath.com/automation-suite/docs/managing-external-applications#generating-a-new-app-secret):

<img src="https://du-nst-cdn.azureedge.net/uipathcli/auth_credentials.gif" />


1. Go to [https://cloud.uipath.com/\<*your-org*\>/portal_/externalApps](https://cloud.uipath.com)

2. Click **+ Add Application**

3. Fill out the fields:
* **Application Name**: *\<your-app\>*
* **Application Type**: `Confidential application` 
* **+ Add Scopes**: Add the permissions you want to assign to your credentials

4. Click **Add** and the app id (`clientId`) and app secret (`clientSecret`) should be displayed.

5. Run the interactive CLI configuration:

```bash
uipathcli config
```

The CLI will ask you to enter the main config settings like
- `clientId` and `clientSecret` to retrieve the JWT bearer token for authentication
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped

```
Enter client id [*******9026]: <your-client-id>
Enter client secret [*******pcnN]: <your-client-secret>
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Successfully configured uipathcli
```

After that the CLI should be ready and you can validate that it is working by invoking one of the services:

```bash
uipathcli metering ping
```

Response:
```json
{
  "location": "westeurope",
  "serverRegion": "westeurope",
  "clusterId": "du-prod-du-we-g-dns",
  "version": "22.11-20-main.v0b5ce6",
  "timestamp": "2022-11-24T09:46:57.3190592Z"
}
```

### OAuth Login

In order to use oauth login, you need to set up an [External Application (Non-Confidential)](https://docs.uipath.com/automation-cloud/docs/managing-external-applications) with a redirect url which points to your local CLI:

<img src="https://du-nst-cdn.azureedge.net/uipathcli/auth_login.gif" />

1. Go to [https://cloud.uipath.com/\<*your-org*\>/portal_/externalApps](https://cloud.uipath.com)

2. Click **+ Add Application**

3. Fill out the fields:
* **Application Name**: *\<your-app\>*
* **Application Type**: `Non-Confidential application` 
* **+ Add Scopes**: Add the permissions you want to grant the CLI
* **Redirect URL**: `http://localhost:12700`

4. Click **Add** and note the app id (`clientId`).

5. Run the interactive CLI configuration:

```bash
uipathcli config --auth login
```

The CLI will ask you to enter the main config settings like
- `clientId`, `redirectUri` and `scopes` which are needed to initiate the OAuth flow
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped

```
Enter client id [*******9026]: <your-external-application-id>
Enter redirect uri [not set]: http://localhost:12700
Enter scopes [not set]: OR.Users
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Successfully configured uipathcli
```

5. After that the CLI should be ready and you can validate that it is working by invoking one of the services:

```bash
uipathcli orchestrator Users_Get
```

### Personal Access Token

You need to generate a personal access token (PAT) and configure the CLI to use it:

1. Go to [https://cloud.uipath.com/\<*your-org*\>/portal_/personalAccessToken](https://cloud.uipath.com)

2. Click **+ Generate new token**

3. Fill out the fields:
* **Name**: *\<token-name\>*
* **Expiration Date**: Set an expiry date for the token
* **+ Add Scopes**: Add the permissions you want to grant the PAT

5. Click **Save** and make sure you copy the generated token.

4. Run the interactive CLI configuration:

```bash
uipathcli config -- auth pat
```

The CLI will ask you to enter the main config settings like
- `pat` your personal access token
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped

```
Enter personal access token [*******9026]: <your-pat>
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Successfully configured uipathcli
```

After that the CLI should be ready and you can validate that it is working by invoking one of the services.

### Configuration File

You can also manually create or edit the configuration file `.uipathcli/config` in your home directory. The following config file sets up the default profile with clientId, clientSecret so that the CLI can generate a bearer token before calling any of the services. It also sets the organization and tenant in the route for services which require it.

```bash
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

```bash
./uipathcli metering ping
```

## Commands and arguments

CLI commands consist of three main parts:

```bash
./uipathcli <service-name> <operation-name> <arguments>
```

- `<service-name>`: The CLI discovers the existing OpenAPI specifications and shows each of them as a separate service
- `<operation-name>`: The operation typically represents the route to call
- `<arguments>`: A list of arguments which are used as request parameters (in the path, header, querystring or body)

Example:

```bash
./uipathcli metering validate --product-name "DU" --model-name "my-model"
```

### Basic arguments

The CLI supports string, integer, floating point and boolean arguments. The arguments are automatically converted to the type defined in the OpenAPI specification:

```bash
./uipathcli product create --name "new-product" --stock "5" --price "1.4" --deleted "false"
```

### Array arguments

Array arguments can be passed as comma-separated strings and are automatically converted to arrays in the JSON body. The CLI supports string, integer, floating point, boolean and object arrays.

```bash
./uipathcli product list --name-filter "my-product,new-product"
```

### Nested Object arguments

More complex nested objects can be passed as semi-colon separated list of property assigments:

```bash
./uipathcli product create --product "name=my-product;price.value=340;price.sale.discount=10;price.sale.value=306"
```

The command creates the following JSON body in the HTTP request:

```json
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

```bash
./uipathcli digitizer digitize --file "hello-world"
```

CLI arguments can also refer to files on disk. This command reads the invoice from `/documents/invoice.pdf` and uploads it to the digitize endpoint:

```bash
./uipathcli digitizer digitize ---file file:///documents/invoice.pdf
```

## Standard input (stdin) / Pipes

You can pipe JSON into the CLI as stdin and it will be used as the request body instead of CLI parameters. The following example reads an orchestrator setting, modifies the value using `jq` and pipes the output back into to the CLI to update it:

```bash
./uipathcli orchestrator settings-get | jq '.value[] | select(.Id == "Alerts.Email.Enabled") | .Value = "FALSE"' | ./uipathcli orchestrator settings-putbyid
```

## Debug

You can set the environment variable `UIPATH_DEBUG=true` or pass the parameter `--debug` in order to see detailed output of the request and response messages:

```bash
./uipathcli metering ping --debug
```

```bash
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

```yaml
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

```bash
./uipathcli metering ping --profile alpha
```

You can also change the profile using an environment variable (`UIPATH_PROFILE`):

```bash
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

## FAQ

### How to run the CLI against Automation Suite?

You can set up a separate profile for automation suite which configures the URI and disables HTTPS certificate checks (if necessary).

*Note: Disabling HTTPS certificate validation imposes a security risk. Please make sure you understand the implications of this setting and just disable the certificate check when absolutely necessary.*

```yaml
profiles:
  - name: automationsuite
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    path:
      organization: test
      tenant: DefaultTenant
    insecure: true
```

And you simply call the CLI with the `--profile automationsuite` parameter:

```bash
./uipathcli metering ping --profile automationsuite
```

### How to contribute?

Take a look at the [contribution guide](CONTRIBUTING.md) for details on how to contribute to this project.