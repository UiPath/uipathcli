[![GitHub Release](https://img.shields.io/github/release/UiPath/uipathcli?style=flat-square)](https://github.com/UiPath/uipathcli/releases)
[![MIT License](https://img.shields.io/badge/license-MIT-blue?style=flat-square)](https://github.com/UiPath/uipathcli/blob/main/LICENSE)
[![Build](https://img.shields.io/github/actions/workflow/status/UiPath/uipathcli/ci.yaml?style=flat-square&color=light-green)](https://github.com/UiPath/uipathcli/actions)
[![Code Coverage](https://img.shields.io/coverallsCoverage/github/UiPath/uipathcli?style=flat-square&color=light-green)](https://coveralls.io/github/UiPath/uipathcli)
[![Documentation](https://img.shields.io/badge/Documentation-146e22?style=flat-square&logo=gitbook&logoColor=white)](https://uipath.github.io/uipathcli/)

# uipathcli

The uipathcli project is a command line interface to simplify, script and automate API calls for UiPath services. The CLI works on Windows, Linux and MacOS.

<img src="https://raw.githubusercontent.com/UiPath/uipathcli/main/documentation/images/getting_started_uipath.gif" />

*Try it out, enjoy, and let us know what you think. Remember, uipathcli is still in preview phase, a special time when your feedback, thoughts, or questions are more than appreciated, so please submit them [here](https://github.com/UiPath/uipathcli/issues).*

## Install

In order to get started quickly, you can run the install scripts for Windows, Linux and MacOS.

<details open>
  <summary>Install instructions for x86_64/amd64</summary>
  <p>

### Windows

```powershell
Invoke-WebRequest "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-windows-amd64.zip" -OutFile "uipathcli.zip" ; Expand-Archive -Force -Path "uipathcli.zip" -DestinationPath "."
```

### Linux

```bash
curl -sL "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-linux-amd64.tar.gz" | tar -xzv
```

### MacOS

```bash
curl -sL "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-darwin-amd64.tar.gz" | tar -xzv
```

  </p>
</details>

<details>
  <summary>Install instructions for arm64</summary>
  <p>

### Windows (ARM)

```powershell
Invoke-WebRequest "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-windows-arm64.zip" -OutFile "uipathcli.zip" ; Expand-Archive -Force -Path "uipathcli.zip" -DestinationPath "."
```

### Linux (ARM)

```bash
curl -sL "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-linux-arm64.tar.gz" | tar -xzv
```

### MacOS (ARM)

```bash
curl -sL "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-darwin-arm64.tar.gz" | tar -xzv
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

- **Client Credentials**: Generate secret and configure the CLI to use these long-term credentials. Client credentials should be used in case you want to use the CLI from a script in an automated way. 

- **OAuth Login**: Login to UiPath using your browser and SSO of choice. This is the preferred flow when you are using the CLI interactively. No need to manage any credentials.

- **Personal Access Token**: Generate a PAT and configure the CLI to use the access token.

### Client Credentials

In order to use client credentials, you need to set up an [External Application (Confidential)](https://docs.uipath.com/automation-cloud/docs/managing-external-applications) and generate an [application secret](https://docs.uipath.com/automation-suite/docs/managing-external-applications#generating-a-new-app-secret):

<img src="https://raw.githubusercontent.com/UiPath/uipathcli/main/documentation/images/getting_started_auth_credentials.gif" />


1. Go to [https://cloud.uipath.com/\<*your-org*\>/portal_/externalApps](https://cloud.uipath.com)

2. Click **+ Add Application**

3. Fill out the fields:
* **Application Name**: *\<your-app\>*
* **Application Type**: `Confidential application` 
* **+ Add Scopes**: Add the permissions you want to assign to your credentials, e.g. `OR.Users.Read`

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

After that the CLI should be ready and you can validate that it is working by invoking one of the services (requires `OR.Users.Read` scope):

```bash
uipath orchestrator users get
```

Response:
```json
{
  "@odata.context": "https://cloud.uipath.com/uipatcleitzc/DefaultTenant/orchestrator_/odata/$metadata#Users",
  "@odata.count": 1,
  "value": [
    {
      "CreationTime": "2021-10-19T10:49:18.907Z",
      "Name": "Administrators",
      "Type": "DirectoryGroup",
      "UserName": "administrators"
    }
  ]
}
```

### OAuth Login

In order to use oauth login, you need to set up an [External Application (Non-Confidential)](https://docs.uipath.com/automation-cloud/docs/managing-external-applications) with a redirect url which points to your local CLI:

<img src="https://raw.githubusercontent.com/UiPath/uipathcli/main/documentation/images/getting_started_auth_login.gif" />

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
uipath orchestrator users get
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
uipath config --auth pat
```

The CLI will ask you to enter the main config settings like
- `organization` and `tenant` used by UiPath services which are account-scoped or tenant-scoped
- `pat` your personal access token

```
Enter organization [not set]: uipatcleitzc
Enter tenant [not set]: DefaultTenant
Enter personal access token [*******26-1]: rt_B637A751...
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
uipath orchestrator users get
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

## Quickstart Guide

This section explains how to use the uipathcli and highlights some of the most common use-cases.

## Package, Deploy and Run Studio Projects

The CLI makes it very easy to package and upload studio projects. You can create a release and execute it using UiPath Orchestrator.

```bash
# Download example project
projectUrl="https://raw.githubusercontent.com/UiPath/uipathcli/refs/heads/main/plugin/studio/projects/crossplatform"
curl --remote-name "$projectUrl/Main.xaml" \
     --remote-name "$projectUrl/project.json"

# Build and package project
uipath studio package pack

# Upload package
uipath orchestrator processes upload-package --file "MyProcess.1.0.0.nupkg"

# Create release
folderId=$(uipath orchestrator folders get --query "value[0].Id")
releaseKey=$(uipath orchestrator releases post --folder-id $folderId \
                                               --name "MyProcess" \
                                               --process-key "MyProcess" \
                                               --process-version "1.0.0" \
                                               --query "Key" \
                                               --output text)

# Start process
jobId=$(uipath orchestrator jobs start-jobs --folder-id $folderId \
                                            --start-info "ReleaseKey=$releaseKey" \
                                            --query "value[0].Id")
uipath orchestrator jobs get-by-id --folder-id $folderId --key $jobId
```

## Manage Orchestrator Resources

There are various UiPath Orchestrator resources which you can manage through the CLI. This example shows how to create new assets and query for the existing ones:

```bash
folderId=$(uipath orchestrator folders get --query "value[0].Id")

# Create new Text asset
uipath orchestrator assets post --folder-id $folderId \
                                --name "MyAsset" \
                                --value-scope "Global" \
                                --value-type "Text" \
                                --string-value "my-value"

# List existing assets
uipath orchestrator assets get --folder-id $folderId
```

## Classify Documents using Document Understanding

You can use the CLI to upload a document and classify it using UiPath Document Understanding.

```bash
# Download example invoice
curl --remote-name "https://raw.githubusercontent.com/UiPath/uipathcli/refs/heads/main/documentation/examples/invoice.jpg"

# Start digitization and classification using the default project
uipath du digitization start --file invoice.jpg | uipath du classification classify --query "classificationResults[0].DocumentTypeId" --file -
```

```json
"invoices"
```

## Extract Data from Documents using Document Understanding

This snippet shows how to upload a document and extract data using UiPath Document Understanding. The CLI converts the result into a list of fields, the extracted values as well as their confidence levels:

```bash
# Download example invoice
curl --remote-name "https://raw.githubusercontent.com/UiPath/uipathcli/refs/heads/main/documentation/examples/invoice.jpg"

# Digitize the file and extract using the default project
uipath du digitization start --file "invoice.jpg" | uipath du extraction extract --query "extractionResult.ResultsDocument.Fields[?not_null(Values)].{field: FieldId, value: Values[0].Value, confidence: Values[0].Confidence}" --file - 
```

```json
[
  {
    "field": "name",
    "value": "Sit Amet Corp.",
    "confidence": 0.996
  },
  {
    "field": "total",
    "value": "17310.00",
    "confidence": 0.995
  }
  ...
]
```

## Extract Data from Documents using Generative Extractor

```bash
# Download example invoice
curl --remote-name "https://raw.githubusercontent.com/UiPath/uipathcli/refs/heads/main/documentation/examples/invoice.jpg"

# Digitize the file
documentId=$(uipath du digitization start --file "invoice.jpg" --query "documentId" --output text)

# Extract the total amount using the Generative Extractor
uipath du extraction extract --project-id 00000000-0000-0000-0000-000000000001 --extractor-id "generative_extractor" --document-id "$documentId" --query "extractionResult.ResultsDocument.Fields[0].Values[0].Value" --prompts "id=total; question=The total amount"
```

```json
"$17,310.00"
```

## Connect to Automation Suite

Log in to Automation Suite as an organization administrator and set up an external application to generate the client secrets for the uipathcli. After that, you can configure a new CLI profile for your automation suite server to specify the access url and client secrets:

```yaml
profiles:
  - name: automationsuite
    organization: test
    tenant: DefaultTenant
    auth:
      clientId: <your-client-id>
      clientSecret: <your-client-secret>
    uri: https://<your-automation-suite-cluster-url>
```

*Note: You can also disable HTTPS certificate validation by adding `insecure: true` to you profile but this imposes a security risk. Please make sure you understand the implications of this setting and just disable the certificate check when absolutely necessary.*

And you simply call the CLI with the `--profile automationsuite` parameter:

```bash
uipath orchestrator users get --profile automationsuite
```

## Commands and arguments

CLI commands consist of four main parts:

```bash
uipath <service-name> <resource-name> <operation-name> --<argument1> --<argument2>
```

- `<service-name>`: The UiPath product or service to call
- `<resource-name>`: The resource to access
- `<operation-name>`: The operation to perform on the resource
- `<argument>`: A list of arguments passed to the operation

Example:

```bash
uipath orchestrator folders get --orderby "DisplayName"
```

### Basic arguments

The CLI supports string, integer, floating point and boolean arguments. The arguments are automatically converted to the expected type:

```bash
uipath orchestrator folders get --orderby "DisplayName" --top 10
```

### Array arguments

Array arguments can be passed as comma-separated strings and are automatically converted to arrays in the JSON body. The CLI supports string, integer, floating point and boolean arrays.

```bash
uipath orchestrator jobs stop-jobs --folder-id "2000021" --strategy "SoftStop" --job-ids "451019658,451019773"
```

You can also provide arrays by specifing the same parameter multiple times:

```bash
uipath orchestrator jobs stop-jobs --folder-id "2000021" --strategy "SoftStop" --job-ids "451019658" --job-ids "451019773"
```

### Nested Object arguments

More complex nested objects can be passed as semi-colon separated list of property assigments:

```bash
uipath orchestrator jobs start-jobs --folder-id "2000021" --start-info "ReleaseKey=4bfcd6e6-44ae-46d2-b1a5-d8647bec8b66; RunAsMe=false; RuntimeType=Unattended"
```

The command creates the following JSON body in the HTTP request:

```json
{
  "startInfo": {
    "releaseKey": "4bfcd6e6-44ae-46d2-b1a5-d8647bec8b66",
    "runAsMe": false, 
    "runtimeType": "Unattended"
  }
}
```

You can also specify JSON directly as an argument, e.g.:
```bash
uipath orchestrator jobs start-jobs --folder-id "2000021" --start-info '{"releaseKey":"4bfcd6e6-44ae-46d2-b1a5-d8647bec8b66","runAsMe":false,"runtimeType":"Unattended"}'
```

### File Upload arguments

You can upload a file on disk using the `--file` argument. The following command reads the invoice from `documents/invoice.pdf` and uploads it to the digitize endpoint:

```bash
uipath du digitization digitize --project-id "c10e9750-7d33-46ba-8484-9e5cf6ea7374" --file documents/invoice.pdf
```

## Standard input (stdin) / Pipes

You can pipe JSON or any other input into the CLI as stdin and it will be used as the request body when the `--file -` argument was provided:

```bash
cat documents/invoice.pdf | uipath du digitization digitize --project-id "c10e9750-7d33-46ba-8484-9e5cf6ea7374" --content-type "application/pdf" --file -
```

## Output formats

The CLI supports multiple output formats:

- `json` (default): HTTP response is rendered as prettified json on standard output. The output can be used to pipe into `jq` or other command line utilities which support json.

- `text`: Fields are tab-separated and rows are outputted on separate lines. This output can be easily processed by standard unix tools like `cut`, `grep`, `sort`, etc...

In order to switch to text output, you can either set the environment variable `UIPATH_OUTPUT` to `text`, change the setting in your profile or pass it as an argument to the CLI:

```bash
uipath orchestrator users get --query "value[].Name" --output text

Administrator
Automation User
Automation Developer
```

```bash
uipath orchestrator users get --query "value[].[Name, CreationTime]" --output text | while IFS=$'\t' read -r name creation_time; do
    echo "User ${name} was created at ${creation_time}"
done

User Administrator was created at 2023-01-25T12:49:18.907Z
User Thomas was created at 2023-01-26T10:35:15.736Z
```

## Queries

The CLI supports [JMESPath queries](https://jmespath.org/tutorial.html) to filter and modify the service response on the client-side. This does not replace server-side filtering which is more efficient and works across paginated results but allows you to modify the CLI output without the need to install any external tools.

Examples:

```bash
# Select only the name of all returned users
uipath orchestrator users get --query "value[].Name"

[
  "Administrator",
  "Automation User",
  "Automation Developer"
]
```

```bash
# Select the first user with the name "Administrator"
uipath orchestrator users get --query "value[?Name == 'Administrator'] | [0]"

{
  "Id": 123456,
  "CreationTime": "2023-01-27T10:45:24.763Z",
  "Name": "Administrator"
}
```

```bash
# Sort the users by creation time and get the name of last created user
uipath orchestrator users get --query "sort_by(value, &CreationTime) | [-1].Name"

"Automation Developer"
```

## Debug

You can set the environment variable `UIPATH_DEBUG=true` or pass the parameter `--debug` in order to see detailed output of the request and response messages:

```bash
uipath du discovery projects --debug
```

```bash
GET https://cloud.uipath.com/uipatcleitzc/DefaultTenant/du_/api/framework/projects?api-version=1 HTTP/1.1
X-Request-Id: b033e39294147bcb1174c5b7ace6ac7c
Authorization: Bearer ...


HTTP/1.1 200 OK
Connection: keep-alive
Content-Type: application/json; charset=utf-8

{"projects":[{"id":"00000000-0000-0000-0000-000000000000","name":"Predefined"}]}


{
  "projects": [
    {
      "id": "00000000-0000-0000-0000-000000000000",
      "name": "Predefined"
    },
  ]
}
```

## Wait for conditions

You can specify JMESPath expressions on the response body to retry an operation until the provided condition evaluates to true. This allows you to write a sync call which waits for some backend operation to be carried out instead of polling manually.

The following command digitizes a file:

```bash
documentId=$(uipath du digitization start --file "invoice.jpg" --query "documentId" --output text)
uipath du digitization get --document-id $documentId
```

But the operation is asynchronous and can take some time to complete. Instead of calling the `get` operation in a loop, you can use the `--wait` flag with a condition to wait for:

```bash
uipath du digitization get --document-id $documentId --wait "status == 'Succeeded'"
```

The default timeout is 30s but can be adjusted by providing the `--wait-timeout` flag, e.g.

```bash
uipath du digitization get --document-id $documentId --wait "status == 'Succeeded'" --wait-timeout 300
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
uipath orchestrator users get --profile alpha
```

You can also change the profile using an environment variable (`UIPATH_PROFILE`):

```bash
UIPATH_PROFILE=alpha uipath orchestrator users get
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
| `--identity-uri` | `UIPATH_IDENTITY_URI` | `uri` | `https://cloud.uipath.com/identity_` | URL override for identity calls |
| | `UIPATH_CLIENT_ID` | `string` | | Client Id |
| | `UIPATH_CLIENT_SECRET` | `string` | | Client Secret |
| | `UIPATH_PAT` | `string` | | Personal Access Token |
| `--wait` | | `string` | | [JMESPath expression](https://jmespath.org/) to wait for |
| `--wait-timeout` | | `integer` | 30 | Time in seconds until giving up waiting for condition  |

## How to contribute?

Take a look at the [contribution guide](CONTRIBUTING.md) for details on how to contribute to this project.
