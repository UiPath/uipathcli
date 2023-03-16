# UiPath OpenAPI Command-Line-Interface (preview)

The UiPath OpenAPI CLI project is a command line interface to simplify, script and automate API calls for UiPath services. The CLI works on Windows, Linux and MacOS.

<img src="https://du-nst-cdn.azureedge.net/uipathcli/getting_started_uipath.gif" />

## Install

*Try it out, enjoy, and let us know what you think. Remember, uipathcli is still in preview phase, a special time when your feedback, thoughts, or questions are more than appreciated, so please submit them [here](https://github.com/UiPath/uipathcli/issues).*

In order to get started quickly, you can run the install scripts for Windows, Linux and MacOS.

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

For autocompletion to work, the `uipath` executable needs to be in your PATH. Make sure the following commands output the path to the `uipath` executable:

### PowerShell

```powershell
(Get-Command uipath).Path
```

### Bash

```bash
which uipath
```

You can enable autocompletion by running the following commands depending on which shell you are using:

### PowerShell

```powershell
uipath autocomplete enable --shell "powershell"
```

### Bash

```bash
uipath autocomplete enable --shell "bash"
```

  </p>
</details>

<br />

After installing the `uipath` executable, you can run the interactive config command to finish setting up your CLI:

```
uipath config
```

More details about how to configure the CLI can be found in the following sections.

## Configuration

The CLI supports multiple ways to authorize with the UiPath services:
- **Client Credentials**: Generate secret and configure the CLI to use these long-term credentials.

- **OAuth Login**: Login to UiPath using your browser and SSO of choice: Microsoft Login, Google Login, LinkedIn, Custom SSO or simple username/password. No need to manage any credentials.

- **Personal Access Token**: Generate a PAT and configure the CLI to use the access token.

### Client Credentials

In order to use client credentials, you need to set up an [External Application (Confidential)](https://docs.uipath.com/automation-cloud/docs/managing-external-applications) and generate an [application secret](https://docs.uipath.com/automation-suite/docs/managing-external-applications#generating-a-new-app-secret):

<img src="https://du-nst-cdn.azureedge.net/uipathcli/getting_started_auth_credentials.gif" />


1. Go to [https://cloud.uipath.com/\<*your-org*\>/portal_/externalApps](https://cloud.uipath.com)

2. Click **+ Add Application**

3. Fill out the fields:
* **Application Name**: *\<your-app\>*
* **Application Type**: `Confidential application` 
* **+ Add Scopes**: Add the permissions you want to assign to your credentials

4. Click **Add** and the app id (`clientId`) and app secret (`clientSecret`) should be displayed.

5. Run the interactive CLI configuration:

```bash
uipath config --auth credentials
```

The CLI will ask you to enter the main config settings like
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped
- `clientId` and `clientSecret` to retrieve the JWT bearer token for authentication

```
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Enter client id [*******9026]: <your-client-id>
Enter client secret [*******pcnN]: <your-client-secret>
Successfully configured uipath CLI
```

After that the CLI should be ready and you can validate that it is working by invoking one of the services:

```bash
uipath du metering ping
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

<img src="https://du-nst-cdn.azureedge.net/uipathcli/getting_started_auth_login.gif" />

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
uipath config --auth login
```

The CLI will ask you to enter the main config settings like
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped
- `clientId`, `redirectUri` and `scopes` which are needed to initiate the OAuth flow

```
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Enter client id [*******9026]: <your-external-application-id>
Enter redirect uri [not set]: http://localhost:12700
Enter scopes [not set]: OR.Users
Successfully configured uipath CLI
```

5. After that the CLI should be ready and you can validate that it is working by invoking one of the services:

```bash
uipath orchestrator Users_Get
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
uipath config -- auth pat
```

The CLI will ask you to enter the main config settings like
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped
- `pat` your personal access token

```
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Enter personal access token [*******9026]: <your-pat>
Successfully configured uipath CLI
```

After that the CLI should be ready and you can validate that it is working by invoking one of the services.

### Configuration File

You can also manually create or edit the configuration file `.uipath/config` in your home directory. The following config file sets up the default profile with clientId, clientSecret so that the CLI can generate a bearer token before calling any of the services. It also sets the organization and tenant for services which require it.

```bash
cat <<EOT > $HOME/.uipath/config
---
profiles:
  - name: default
    organization: <organization-name>
    tenant: <tenant-name>
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
EOT
```

Once you have created the configuration file with the proper secrets, org and tenant information, you should be able to successfully call the services, e.g.

```bash
uipath du metering ping
```

### Set Config File Values

The CLI provides the command `uipath config set` to update values in the configuration file.

Example: Set the organization in the default profile

```bash
uipath config set --key "organization" --value "myorg"
```

Example: Create profile which uses the staging environment

```bash
uipath config set --key "uri" --value "https://staging.uipath.com" --profile staging
```

## Commands and arguments

CLI commands consist of three main parts:

```bash
uipath <service-name> <operation-name> <arguments>
```

- `<service-name>`: The CLI discovers the existing OpenAPI specifications and shows each of them as a separate service
- `<operation-name>`: The operation typically represents the route to call
- `<arguments>`: A list of arguments which are used as request parameters (in the path, header, querystring or body)

Example:

```bash
uipath du metering validate --product-name "DU" --model-name "my-model"
```

### Basic arguments

The CLI supports string, integer, floating point and boolean arguments. The arguments are automatically converted to the type defined in the OpenAPI specification:

```bash
uipath product create --name "new-product" --stock "5" --price "1.4" --deleted "false"
```

### Array arguments

Array arguments can be passed as comma-separated strings and are automatically converted to arrays in the JSON body. The CLI supports string, integer, floating point, boolean and object arrays.

```bash
uipath product list --name-filter "my-product,new-product"
```

### Nested Object arguments

More complex nested objects can be passed as semi-colon separated list of property assigments:

```bash
uipath product create --product "name=my-product;price.value=340;price.sale.discount=10;price.sale.value=306"
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

CLI arguments can also refer to files on disk. This command reads the invoice from `/documents/invoice.pdf` and uploads it to the digitize endpoint:

```bash
uipath digitizer digitize ---file documents/invoice.pdf
```

## Standard input (stdin) / Pipes

You can pipe JSON into the CLI as stdin and it will be used as the request body instead of CLI parameters. The following example reads an orchestrator setting, modifies the value using `jq` and pipes the output back into to the CLI to update it:

```bash
uipath orchestrator settings-get | jq '.value[] | select(.Id == "Alerts.Email.Enabled") | .Value = "FALSE"' | uipath orchestrator settings-putbyid
```

## Output formats

The CLI supports multiple output formats:

- `json` (default): HTTP response is rendered as prettified json on standard output. The output can be used to pipe into `jq` or other command line utilities which support json.

- `text`: Fields are tab-separated and rows are outputted on separate lines. This output can be easily processed by standard unix tools like `cut`, `grep`, `sort`, etc...

In order to switch to text output, you can either set the environment variable `UIPATH_OUTPUT` to `text`, change the setting in your profile or pass it as an argument to the CLI:

```bash
uipath du metering ping --output text

du-prod-du-we-g-dns westeurope  westeurope  2023-01-27T10:56:59.8477522Z  23.1-105-main.v3105ad
```

```bash
uipath orchestrator users-get --query "value[].[Name, CreationTime]" --output text | while IFS=$'\t' read -r name creation_time; do
    echo "User ${name} was created at ${creation_time}"
done

User Administrator was created at 2023-01-25T12:49:18.907Z
User Thomas was created at 2023-01-26T10:35:15.736Z
```

## Queries

The CLI supports [JMESPath queries](https://jmespath.org/tutorial.html) to filter and modify the service response on the client-side. This does not replace server-side filtering which is more efficient and works across paginated results. JMESPath queries simply allow you to modify the CLI output only without the need to install any external tools.

Examples:

```bash
# Select only the name of all returned users
uipath orchestrator users-get --query "value[].Name"

[
  "Administrator",
  "Automation User",
  "Automation Developer"
]
```

```bash
# Select the first user with the name "Administrator"
uipath orchestrator users-get --query "value[?Name == 'Administrator'] | [0]"

{
  "Id": 123456,
  "CreationTime": "2023-01-27T10:45:24.763Z",
  "Name": "Administrator",
  ...
}
```

```bash
# Sort the users by creation time and get the name of last created user
uipath orchestrator users-get --query "sort_by(value, &CreationTime) | [-1].Name"

"Automation Developer"
```



## Debug

You can set the environment variable `UIPATH_DEBUG=true` or pass the parameter `--debug` in order to see detailed output of the request and response messages:

```bash
uipath du metering ping --debug
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
    organization: uipatcleitzc
    tenant: DefaultTenant
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
  - name: apikey
    uri: https://du.uipath.com/metering/
    header:
      X-UIPATH-License: <your-api-key>
      X-UIPATH-MLService: MLSERVICE_TIEMODEL
    output: text
  - name: alpha
    uri: https://alpha.uipath.com
    organization: UiPatricjvjx
    tenant: DefaultTenant
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
```

If you do not provide the `--profile` parameter, the `default` profile is automatically selected. Otherwise it will use the settings from the provided profile. The following command will send a request to the alpha.uipath.com environment:

```bash
uipath du metering ping --profile alpha
```

You can also change the profile using an environment variable (`UIPATH_PROFILE`):

```bash
UIPATH_PROFILE=alpha uipath du metering ping
```

## Global Arguments

You can either pass global arguments as CLI parameters, set an env variable or set them using the configuration file. Here is a list of the supported global arguments which can be applied to all CLI operations:

| Name | Env-Variable | Type | Default Value | Description |
| ----------- | ----------- | ----------- | ----------- | ----------- |
| `--debug` | `UIPATH_DEBUG` | `boolean` | `false` | Show debug output |
| `--insecure` | `UIPATH_INSECURE` | `boolean` | `false` |*Warning: Disables HTTPS certificate checks* |
| `--output` | `UIPATH_OUTPUT` | `string` | `json` | Response output format, supported values: json and text |
| `--profile` | `UIPATH_PROFILE` | `string` | `default` | Use profile from configuration file |
| `--query` | | `string` | | [JMESPath queries](https://jmespath.org/) for transforming the output |
| `--uri` | `UIPATH_URI` | `uri` | `https://cloud.uipath.com` | URL override |
| `--organization` | `UIPATH_ORGANIZATION` | `string` | | Organization name |
| `--tenant` | `UIPATH_TENANT` | `string` | | Tenant name |
| | `UIPATH_CLIENT_ID` | `string` | | Client Id |
| | `UIPATH_CLIENT_SECRET` | `string` | | Client Secret |
| | `UIPATH_PAT` | `string` | | Personal Access Token |


## FAQ

### How to run the CLI against Automation Suite?

You can set up a separate profile for automation suite which configures the URI and disables HTTPS certificate checks (if necessary).

*Note: Disabling HTTPS certificate validation imposes a security risk. Please make sure you understand the implications of this setting and just disable the certificate check when absolutely necessary.*

```yaml
profiles:
  - name: automationsuite
    organization: test
    tenant: DefaultTenant
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    uri: https://sfdev1234567-cluster.infra-sf-ea.infra.uipath-dev.com
    insecure: true
```

And you simply call the CLI with the `--profile automationsuite` parameter:

```bash
uipath du metering ping --profile automationsuite
```

### How to bootstrap Automation Suite?

You can use the CLI to create a new org on Automation Suite and license the server. As a prerequisite, you need to create a client secret on the server which allows grant type `password`. Once set up, you can configure the CLI to retrieve bearer tokens for the `Host` admin user:

```yaml
profiles:
  - name: default
    organization: Host
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
      grantType: password
      properties:
        username: admin
        password: <your-admin-password>
        acr_values: tenant:Host
    uri: https://sfdev1234567-cluster.infra-sf-ea.infra.uipath-dev.com
    insecure: true
```

After that you can create a new organization and license it:

```bash
org_name="testorg"
password="<new-password>"
license_code="<license-code>"

# Create a new organization
org_id=$(uipath oms on-prem-organization create-organization-on-prem
  --organization-name "$org_name" \
  --admin-email "admin@uipath.com" \
  --admin-user-name "admin" \
  --admin-first-name "Admin" \
  --admin-last-name "User" \
  --admin-password "$password" \
  --language "en" \
  --query "id" \
  --output "text")

# Use the new organization and activate it
uipath config set --key "organization" --value "$org_name"
uipath config set --key "auth.properties.acr_values" --value "tenant:$org_id"
uipath config set --key "auth.properties.password" --value "$password"
uipath oms license activate --license "$license_code"
```

### How to contribute?

Take a look at the [contribution guide](CONTRIBUTING.md) for details on how to contribute to this project.
