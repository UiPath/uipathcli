#! /bin/bash

############################################################
# Install script for uipathcli
#
# DESCRIPTION:
# This scripts makes it easy to get started and set up the
# UiPath CLI locally.
# Downloads the uipathcli binaries, service definitions
# and sets up a default configuration.
############################################################

set -e

############################################################
# Installs uipathcli using the go install command
############################################################
function install_uipathcli()
{
    local tmp_file=`mktemp`
    wget "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli.zip" --output-document=$tmp_file --quiet
    unzip -qq -o -d "." $tmp_file
    rm $tmp_file
    if [[ "$OSTYPE" == "darwin"* ]]; then
        mv uipathcli uipathcli.linux
        mv uipathcli.osx uipathcli
    fi
    chmod +x uipathcli
}

############################################################
# Finds the config file and returns the path to it
#
# Returns:
#   The path to the config file
############################################################
function get_defaultconfigfile()
{
    local config_file="$HOME/.uipathcli/config"
    echo "$config_file"
}

############################################################
# Creates a configuration file with default content
#
# Arguments:
#   - Path to the configuration file
#     typically $HOME/.uipathcli/config
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
    auth:
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

config_file=$(get_defaultconfigfile)
if [ ! -f "$config_file" ]
then
    echo -e "Setting up default config file..."
    config_content=$(create_defaultconfigfile "$config_file")
    echo -e "\n$config_content\n"
fi

echo -e "\nConfiguration file: '$config_file'\n"

echo -e "You can configure the CLI running the command:"
echo -e "    ./uipathcli config\n"

GREEN='\033[0;32m'
NO_COLOR='\033[0m'
echo -e "${GREEN}Successfully installed uipathcli${NO_COLOR}"
