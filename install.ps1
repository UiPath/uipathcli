<#
    .SYNOPSIS
        Install script for uipathcli
    .DESCRIPTION
        Clones the uipathcli source code, builds it using the go compiler and installs it.
        Additionally, it downloads service definitions which are used to generate the CLI commands.
#>

if ($null -eq (Get-Command "go" -ErrorAction SilentlyContinue))
{ 
   Write-Host "Unable to find go executable in your PATH. Please install golang from https://go.dev/dl/"
   exit 1
}

<#
    .SYNOPSIS
        Installs uipathcli using the go install command
#>
function Install-uipathcli() {
    go env -w GOPRIVATE=github.com/UiPath/uipathcli
    go install github.com/UiPath/uipathcli@latest
}

<#
    .SYNOPSIS
        Installs uipathcli kubernetes authenticator using the go install command
#>
function Install-uipathcli-authenticator-k8s() {
    go env -w GOPRIVATE=github.com/UiPath/uipathcli-authenticator-k8s
    go install github.com/UiPath/uipathcli-authenticator-k8s@latest
}

<#
    .SYNOPSIS
        Finds the definitions folder and creates it, if needed
    .OUTPUTS
        System.String. The path to the definitions folder
#>
function New-DefinitionsFolder() {
    $definitionsFolder = Join-Path -Path "$(go env GOPATH)" -ChildPath "/bin/definitions"
    New-Item -ItemType Directory -Force -Path "$definitionsFolder" | Out-Null
    return $definitionsFolder
}

<#
    .SYNOPSIS
        Downloads the OpenAPI document
    .PARAMETER Path
        System.String. Path to the definitions folder which needs to be next
        to the uipathcli executable, typically $GOPATH/bin/definitions
    .PARAMETER Name
        System.String. The service name
    .PARAMETER Url
        System.String. Url of the Open API definition
#>
function Get-OpenApiDoc() {
    param (
        [Parameter(Mandatory = $true)]
        [String]$Path,
        [Parameter(Mandatory = $true)]
        [String]$Name,
        [Parameter(Mandatory = $true)]
        [String]$Url
    )

    $ProgressPreference = "SilentlyContinue"
    $fileName = Join-Path -Path "$path" -ChildPath "$name.yaml"
    Invoke-WebRequest -Uri "$url" -OutFile "$fileName"
    $ProgressPreference = "Continue"
}

<#
    .SYNOPSIS
        Finds the config file and returns the path to it
    .OUTPUTS
        System.String. The path to the config file
#>
function Get-DefaultConfigFile() {
    $configFile = Join-Path -Path "$env:userprofile" -ChildPath "/.uipathcli/config"
    return $configFile
}

<#
    .SYNOPSIS
        Creates a configuration file with default content
    .PARAMETER ConfigFile
        System.String. Path to the configuration file,
        typically $HOME/.uipathcli/config
    .OUTPUTS
        System.String. The new default content of the configuration file
#>
function New-DefaultConfigFile() {
    param (
        [Parameter(Mandatory = $true)]
        [String]$ConfigFile
    )

    New-Item -Force -Path $ConfigFile | Out-Null
    $content = @"
profiles:
  - name: default
    uri: https://cloud.uipath.com
    auth:
      clientId: <enter your client id here>
      clientSecret: <enter your client secret here>
    path:
      organization: <enter your organization name here>
      tenant: <enter your tenant name here>
"@
    Add-Content "$ConfigFile" "$content"
    return $content
}

<#
    .SYNOPSIS
        Finds the plugins file and returns the path to it
    .OUTPUTS
        System.String. The path to the plugins file
#>
function Get-DefaultPluginsFile() {
    $pluginsFile = Join-Path -Path "$env:userprofile" -ChildPath "/.uipathcli/plugins"
    return $pluginsFile
}

<#
    .SYNOPSIS
        Creates a plugins file with the kubernetes authenticator enabled
        (uipathcli-authenticator-k8s)
    .PARAMETER PluginsFile
        System.String. Path to the plugins file,
        typically $HOME/.uipathcli/plugins
    .OUTPUTS
        System.String. The content of the plugins file
#>
function New-DefaultPluginsFile() {
    param (
        [Parameter(Mandatory = $true)]
        [String]$PluginsFile
    )

    New-Item -Force -Path $PluginsFile | Out-Null
    $content = @"
authenticators:
  - name: kubernetes
    path: ./uipathcli-authenticator-k8s
"@
    Add-Content "$PluginsFile" "$content"
    return $content
}

Write-Host "Downloading and installing uipathcli..."

Install-uipathcli

Write-Host "Downloading and installing uipathcli-authenticator-k8s..."

Install-uipathcli-authenticator-k8s

Write-Host "Downloading service definitions..."

$definitionsFolder = New-DefinitionsFolder
Get-OpenApiDoc "$definitionsFolder" "metering" "https://cloud.uipath.com/testdwfdcxqn/DefaultTenant/du_/api/metering/swagger/v1/swagger.yaml"
Get-OpenApiDoc "$definitionsFolder" "events" "https://cloud.uipath.com/testdwfdcxqn/DefaultTenant/du_/api/eventservice/swagger/v1/swagger.yaml"

$pluginsFile = Get-DefaultPluginsFile
if (!(Test-Path -Path "$pluginsFile")) {
    Write-Host "Setting up default plugins file..."
    $pluginsContent = New-DefaultPluginsFile "$pluginsFile"
    Write-Host "`n$pluginsContent`n"
}

$configFile = Get-DefaultConfigFile
if (!(Test-Path -Path "$configFile")) {
    Write-Host "Setting up default config file..."
    $configContent = New-DefaultConfigFile "$configFile"
    Write-Host "`n$configContent`n"
}

Write-Host "`nYou can configure the CLI using the config file: '$configFile'`n"
Write-Host -ForegroundColor DarkGreen "Successfully installed uipathcli"
