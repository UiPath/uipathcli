#! /bin/bash

set -o pipefail
set -e

############################################################
# Updates the OpenAPI definitions
#
# DESCRIPTION:
# This scripts retrieves the latest definitions from
# cloud.uipath.com and stores them locally to package 
# with the CLI
############################################################

organization="uipatcleitzc"
tenant="defaulttenant"

############################################################
# Downloads yq and installs it in bin/
# https://github.com/mikefarah/yq
############################################################
function install_yq()
{
  local version="v4.30.8"
  mkdir -p bin/
  wget --quiet "https://github.com/mikefarah/yq/releases/download/$version/yq_linux_amd64.tar.gz" -O - \
  | tar --extract --gunzip --directory=bin/ ./yq_linux_amd64 \
  && mv bin/yq_linux_amd64 bin/yq
}

############################################################
# Download an OpenAPI specification from the given URL
#
# Arguments:
#   - The url to download from
#   - The OpenAPI version, in case of "v2" the specification
#     will be automatically converted to version 3
#
# Returns:
#   The OpenAPI specification content
############################################################
function download_definition()
{
  local url="$1"
  local version="$2"
  if [[ "$version" == "v2" ]]; then
    wget --quiet "https://converter.swagger.io/api/convert?url=$url" -O - -o /dev/null
  else
    wget --quiet "$url" -O - -o /dev/null
  fi
}

############################################################
# Saves the given definition as yaml in the 
# definitions/ folder
#
# Arguments:
#   - The definition name
############################################################
function save_definition()
{
  local name="$1"
  bin/yq eval -P > "definitions/$name.yaml"
}

############################################################
# Updates the server url
#
# Arguments:
#   - The new url
############################################################
function update_server_url() 
{
  local url="$1"
  local tenant_variable=""

  if [[ "$url" == *"{tenant}"* ]]; then
    tenant_variable=',
          "tenant": {
            "description": "The tenant name (or id)",
            "default": "my-tenant"
          }'
  fi

  bin/yq '.servers = 
  [
    {
      "url": "'"$url"'",
      "description": "The production url",
      "variables": {
        "organization": {
          "description": "The organization name (or id)",
          "default": "my-org"
        }'"$tenant_variable"'
      }
    }
  ]' \
  | bin/yq 'del(.schemes)' \
  | bin/yq 'del(.host)' \
  | bin/yq 'del(.basePath)'
}

############################################################
# Sets a property for the given parameter
#
# Arguments:
#   - The parameter name
#   - The property to set
#   - The new property value
############################################################
function set_parameter_property()
{
  local name="$1"
  local property_name="$2"
  local property_value="$3"
  bin/yq '.paths[] |= with(select(.get.parameters != null); (.get.parameters[] | select(.name == "'"$name"'"))."'"$property_name"'" = '"$property_value"')' \
  | bin/yq '.paths[] |= with(select(.post.parameters != null); (.post.parameters[] | select(.name == "'"$name"'"))."'"$property_name"'" = '"$property_value"')' \
  | bin/yq '.paths[] |= with(select(.put.parameters != null); (.put.parameters[] | select(.name == "'"$name"'"))."'"$property_name"'" = '"$property_value"')' \
  | bin/yq '.paths[] |= with(select(.patch.parameters != null); (.patch.parameters[] | select(.name == "'"$name"'"))."'"$property_name"'" = '"$property_value"')' \
  | bin/yq '.paths[] |= with(select(.delete.parameters != null); (.delete.parameters[] | select(.name == "'"$name"'"))."'"$property_name"'" = '"$property_value"')' \
  | bin/yq '.paths[] |= with(select(.head.parameters != null); (.head.parameters[] | select(.name == "'"$name"'"))."'"$property_name"'" = '"$property_value"')'
}

if [ ! -f "bin/yq" ]; then
  echo "Installing yq..."
  install_yq
fi

mkdir -p definitions/

echo "Updating identity definition..."
download_definition "https://cloud.uipath.com/$organization/identity_/swagger/external/swagger.json" \
| save_definition "identity"

echo "Updating du.framework definition..."
download_definition "https://cloud.uipath.com/$organization/$tenant/du_/api/framework/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/du_/api/framework" \
| save_definition "du.framework"

echo "Updating orchestrator definition..."
download_definition "https://cloud.uipath.com/$organization/$tenant/orchestrator_/swagger/v20.0/swagger.json" "v2" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/orchestrator_" \
| set_parameter_property "X-UIPATH-OrganizationUnitId" "x-uipathcli-name" "\"folder-id\"" \
| set_parameter_property "X-UIPATH-OrganizationUnitId" "required" "true" \
| save_definition "orchestrator"
