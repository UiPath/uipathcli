#! /bin/bash

############################################################
# Install script for uipathcli
#
# DESCRIPTION:
#   Clones the uipathcli source code, builds it using the 
#   go compiler and installs it.
#   Additionally, it downloads service definitions which 
#   are used to generate the CLI commands.
############################################################

set -e

if ! command -v "go" &> /dev/null
then
    echo "Unable to find go executable in your PATH. Please install golang from https://go.dev/dl/"
    exit 1
fi

############################################################
# Installs uipathcli using the go install command
############################################################
function install_uipathcli()
{
    go env -w GOPRIVATE=github.com/UiPath/uipathcli
    go install github.com/UiPath/uipathcli@latest
}

############################################################
# Finds the definitions folder and creates it, if needed
#
# Returns:
#   The path to the definitions folder
############################################################
function create_definitionsfolder()
{
    local definitions_folder
    definitions_folder="$(go env GOPATH)/bin/definitions"
    mkdir -p "$definitions_folder"
    echo "$definitions_folder"
}

############################################################
# Downloads the OpenAPI document
#
# Arguments:
#   - Path to the definitions folder which needs to be next 
#     to the uipathcli executable
#     typically $GOPATH/bin/definitions
#   - The service name
#   - Url of the Open API definition
############################################################
function download_openapidoc()
{
    local path="$1"
    local name="$2"
    local url="$3"

    wget --quiet --output-document="$path/$name.yaml" "$url"
}

############################################################
# Finds the config files and returns the path to it
#
# Returns:
#   The path to the config file
############################################################
function get_defaultconfigfile()
{
    local config_file="$HOME/uipathcli.profiles.yaml"
    echo "$config_file"
}

############################################################
# Creates a configuration file with default content
#
# Arguments:
#   - Path to the configuration file
#     typically $HOME/uipathcli.profiles.yaml
#
# Returns:
#   The new default content of the configuration file
############################################################
function create_defaultconfigfile()
{
    local config_file="$1"

    mkdir -p "${config_file%/*}/"

    local config_content=""
    IFS='' read -r -d '' config_content <<"EOF"
profiles:
  - name: default
    uri: https://cloud.uipath.com
    clientId: <enter your client id here>
    clientSecret: <enter your client secret here>
    path:
      organization: <enter your organization name here>
      tenant: <enter your tenant name here>
EOF
    echo "$config_content" > "$config_file"
    echo "$config_content"
}

echo -e "Downloading and installing uipathcli..."

install_uipathcli

echo -e "Downloading service definitions..."

definitions_folder=$(create_definitionsfolder)
download_openapidoc "$definitions_folder" "metering" "https://cloud.uipath.com/testdwfdcxqn/DefaultTenant/du_/api/metering/swagger/v1/swagger.yaml"
download_openapidoc "$definitions_folder" "events" "https://cloud.uipath.com/testdwfdcxqn/DefaultTenant/du_/api/eventservice/swagger/v1/swagger.yaml"
download_openapidoc "$definitions_folder" "digitizer" "https://cloud.uipath.com/testdwfdcxqn/DefaultTenant/du_/api/digitizer/swagger/v1/swagger.yaml"

config_file=$(get_defaultconfigfile)
if [ ! -f "$config_file" ]
then
    echo -e "Setting up default config file..."
    config_content=$(create_defaultconfigfile "$config_file")
    echo -e "\n$config_content\n"
fi
echo -e "\nYou can configure the CLI using the config file: '$config_file'\n"

GREEN='\033[0;32m'
NO_COLOR='\033[0m'
echo -e "${GREEN}Successfully installed uipathcli${NO_COLOR}"
