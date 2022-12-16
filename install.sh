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

set -o pipefail
set -e

############################################################
# Installs uipathcli using the go install command
############################################################
function install_uipathcli()
{
    local tmp_file
    tmp_file=$(mktemp)
    wget "https://du-nst-cdn.azureedge.net/uipathcli/uipathcli.zip" --output-document="$tmp_file" --quiet
    unzip -qq -o -d "." $tmp_file
    rm "$tmp_file"
    if [[ "$OSTYPE" == "darwin"* ]]; then
        mv uipathcli uipathcli.linux
        mv uipathcli.osx uipathcli
    fi
    chmod +x uipathcli
}

############################################################
# Autocomplete snipped for uipathcli
#
# Returns:
#   String containing the autocomplete snipped
############################################################
function get_autocomplete_script()
{
    local content=""
    IFS='' read -r -d '' content <<"EOF"

function _uipathcli_bash_complete()
{
  local executable="${COMP_WORDS[0]}"
  local cur="${COMP_WORDS[COMP_CWORD]}" IFS=$'\n'
  local candidates
  read -d '' -ra candidates < <($executable complete --command "${COMP_LINE}" 2>/dev/null)
  read -d '' -ra COMPREPLY < <(compgen -W "${candidates[*]:-}" -- "$cur")
}
complete -f -F _uipathcli_bash_complete uipathcli
EOF
    echo "$content"
}

############################################################
# Registers autocomplete for uipathcli in .bashrc
############################################################
function enable_autocomplete()
{
    local profile_file="$HOME/.bashrc"

    mkdir -p "${profile_file%/*}/"

    if grep -q "_uipathcli_bash_complete" "$profile_file"
    then
        return
    fi
    local profile_content=""
    profile_content=$(get_autocomplete_script)
    echo "$profile_content" >> "$profile_file"
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
# Empty default config for uipathcli
#
# Returns:
#   String containing the default config
############################################################
function get_defaultconfigcontent()
{
    local content=""
    IFS='' read -r -d '' content <<"EOF"
profiles:
  - name: default
    uri: https://cloud.uipath.com
    auth:
      clientId: 
      clientSecret: 
    path:
      organization: 
      tenant: 
EOF
    echo "$content"
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
    config_content=$(get_defaultconfigcontent)
    echo "$config_content" > "$config_file"
    echo "$config_content"
}

echo -e "Downloading and installing uipathcli..."

install_uipathcli
enable_autocomplete

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
