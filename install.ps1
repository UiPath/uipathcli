<#
    .SYNOPSIS
        Install script for uipathcli
    .DESCRIPTION
        This scripts makes it easy to get started and set up the UiPath CLI locally.
        Downloads the uipathcli binaries, service definitions and sets up a default configuration.
#>

<#
    .SYNOPSIS
        Downloads the uipathcli and extracts it to the current directory
#>
function Install-uipathcli() {
    $tmp = New-TemporaryFile | Rename-Item -NewName { $_ -replace "tmp$", "zip" } -PassThru
    Invoke-WebRequest https://du-nst-cdn.azureedge.net/uipathcli/uipathcli.zip -OutFile $tmp | Out-Null
    $tmp | Expand-Archive -DestinationPath "." -Force
    $tmp | Remove-Item
}

<#
    .SYNOPSIS
        Registers autocomplete for uipathcli in the current user profile of powershell
#>
function Enable-AutoComplete() {
    $profileFile = $PROFILE.CurrentUserAllHosts
    New-Item -Force -Path $profileFile | Out-Null
    if (Select-String -Path $profileFile -Pattern "uipathcliAutocomplete")
    {
        return
    }
    $content = @'
$uipathcliAutocomplete = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $command, $params = $commandAst.ToString() -split " ", 2
    & $command complete --command "$commandAst" | foreach-object {
        [system.management.automation.completionresult]::new($_, $_, 'parametervalue', $_)
    }
}
Register-ArgumentCompleter -Native -CommandName uipathcli -ScriptBlock $uipathcliAutocomplete
'@
    Add-Content "$profileFile" "$content"
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

Write-Host "Downloading and installing uipathcli..."

Install-uipathcli
Enable-AutoComplete

$configFile = Get-DefaultConfigFile
if (!(Test-Path -Path "$configFile")) {
    Write-Host "Setting up default config file..."
    $configContent = New-DefaultConfigFile "$configFile"
    Write-Host "`n$configContent`n"
}

Write-Host "`nConfiguration file: '$configFile'`n"

Write-Host "You can configure the CLI running the command:"
Write-Host "    ./uipathcli.exe config`n"

Write-Host -ForegroundColor DarkGreen "Successfully installed uipathcli"
