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

organization="uipatricjvjx"
tenant="dudev"

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
# Sets all tags to the given value
#
# Arguments:
#   - The new tag value
############################################################
function update_all_tags() 
{
  local tag_name="$1"
  bin/yq '(.paths[] | select(.get.tags != null)).get.tags = ["'"$tag_name"'"]' \
  | bin/yq '(.paths[] | select(.post.tags != null)).post.tags = ["'"$tag_name"'"]' \
  | bin/yq '(.paths[] | select(.put.tags != null)).put.tags = ["'"$tag_name"'"]' \
  | bin/yq '(.paths[] | select(.patch.tags != null)).patch.tags = ["'"$tag_name"'"]' \
  | bin/yq '(.paths[] | select(.delete.tags != null)).delete.tags = ["'"$tag_name"'"]' \
  | bin/yq '(.paths[] | select(.head.tags != null)).head.tags = ["'"$tag_name"'"]'
}

############################################################
# Updates all tags with the given value to a new value
#
# Arguments:
#   - The current tag value
#   - The new tag value
############################################################
function update_tags() 
{
  local previous_tag_name="$1"
  local new_tag_name="$2"
  bin/yq '(.paths[] | select(.get.tags[0] == "'"$previous_tag_name"'")).get.tags = ["'"$new_tag_name"'"]' \
  | bin/yq '(.paths[] | select(.post.tags[0] == "'"$previous_tag_name"'")).post.tags = ["'"$new_tag_name"'"]' \
  | bin/yq '(.paths[] | select(.put.tags[0] == "'"$previous_tag_name"'")).put.tags = ["'"$new_tag_name"'"]' \
  | bin/yq '(.paths[] | select(.patch.tags[0] == "'"$previous_tag_name"'")).patch.tags = ["'"$new_tag_name"'"]' \
  | bin/yq '(.paths[] | select(.delete.tags[0] == "'"$previous_tag_name"'")).delete.tags = ["'"$new_tag_name"'"]' \
  | bin/yq '(.paths[] | select(.head.tags[0] == "'"$previous_tag_name"'")).head.tags = ["'"$new_tag_name"'"]'
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

############################################################
# Shortens aicenter tags and groups operations together
############################################################
function update_aicenter_tags() 
{
  update_tags "action-task-controller" "ActionTask" \
  | update_tags "AI UNITS APIs" "AiUnits" \
  | update_tags "app-controller" "App" \
  | update_tags "AuditLog related APIs" "AuditLog" \
  | update_tags "Dataset Controller related APIs" "Dataset" \
  | update_tags "Dataset Download Controller related APIs" "Dataset" \
  | update_tags "dataset-controller" "Dataset" \
  | update_tags "Directory related APIs" "Directory" \
  | update_tags "Eager Provisioning APIs" "Provisioning" \
  | update_tags "events-controller" "Events" \
  | update_tags "Helper Project Creation Api" "Project" \
  | update_tags "Helper Tenant Provision Api" "Provisioning" \
  | update_tags "Hit Count APIs" "HitCount" \
  | update_tags "import-controller" "Import" \
  | update_tags "labelling-file-controller" "Labelling" \
  | update_tags "labelling-oob-template-controller" "Labelling" \
  | update_tags "labelling-template-controller" "Labelling" \
  | update_tags "ML package image related APIs" "MlPackage" \
  | update_tags "ML package Language related APIs" "MlPackage" \
  | update_tags "ML package related APIs" "MlPackage" \
  | update_tags "ML Skill related APIs" "MlSkill" \
  | update_tags "ML Skill related V2 APIs" "MlSkill" \
  | update_tags "ML Skill replication APIs" "MlSkill" \
  | update_tags "ML Skill Version related APIs" "MlSkill" \
  | update_tags "ML skills cleanup related APIs" "MlSkill" \
  | update_tags "ML skills migration related APIs" "MlSkill" \
  | update_tags "MLPackage versions related APIs" "MlPackage" \
  | update_tags "Permission related APIs" "Permission" \
  | update_tags "permission-controller" "Permission" \
  | update_tags "Project Access Api" "Project" \
  | update_tags "Project Controller related APIs" "Project" \
  | update_tags "RabbitMQ monitoring related APIs" "Rabbitmq" \
  | update_tags "Role creation APIs" "Role" \
  | update_tags "Service entry related APIs" "ServiceEntry" \
  | update_tags "SignedURL Generation Related APIs" "SignedUrl" \
  | update_tags "swagger-ui-config-controller" "Swagger" \
  | update_tags "Tenant APIs" "Tenant" \
  | update_tags "Tenant Controller related APIs" "Tenant" \
  | update_tags "Tenant License Stats APIs" "Tenant" \
  | update_tags "tenant-controller" "Tenant" \
  | update_tags "tenant-rbac-migration-controller" "Tenant" \
  | update_tags "Training pipeline related APIs" "Training" \
  | update_tags "Training run related APIs" "Training"  
}

############################################################
# Bug in license-accountant definition pulls in CLR type
############################################################
function remove_clr_type()
{
  bin/yq 'del(.. | select(has("clrType")).clrType)'
}

if [ ! -f "bin/yq" ]; then
  echo "Installing yq..."
  install_yq
fi

echo "Updating oms definition..."
download_definition "https://alpha.uipath.com/$organization/portal_/organization/swagger/v1.0/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/portal_/organization" \
| save_definition "oms"

echo "Updating identity definition..."
download_definition "https://alpha.uipath.com/$organization/identity_/swagger/internal/swagger.json" \
| save_definition "identity"

echo "Updating du.storage definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/du_/api/storage/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/du_/api/storage" \
| update_all_tags "Storage" \
| bin/yq '.paths."/store/{objectKey}".put.requestBody =
  {
    "content": {
      "application/octet-stream": {
        "schema": {
          "type": "string",
          "format": "binary",
          "description": "The file to upload"
        }
      }
    }
  }' \
| save_definition "du.storage"

echo "Updating du.events definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/du_/api/eventservice/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/du_/api/eventservice" \
| update_all_tags "Events" \
| save_definition "du.events"

echo "Updating du.metering definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/aimetering_/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/aimetering_" \
| update_all_tags "Metering" \
| save_definition "du.metering"

echo "Updating du.digitizer definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/du_/api/digitizer/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/du_/api/digitizer" \
| save_definition "du.digitizer"

echo "Updating du.framework definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/du_/api/framework/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/du_/api/framework" \
| save_definition "du.framework"

echo "Updating orchestrator definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/orchestrator_/swagger/v16.0/swagger.json" "v2" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/orchestrator_" \
| set_parameter_property "X-UIPATH-OrganizationUnitId" "x-name" "\"folder-id\"" \
| set_parameter_property "X-UIPATH-OrganizationUnitId" "required" "true" \
| save_definition "orchestrator"

echo "Updating aicenter.helper definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/aifabric_/ai-helper/v3/api-docs/aiCenterApi" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/aifabric_/ai-helper" \
| update_aicenter_tags \
| save_definition "aicenter.helper"

echo "Updating aicenter.pkgmanager definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/aifabric_/ai-pkgmanager/v3/api-docs/aiCenterApi" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/aifabric_/ai-pkgmanager" \
| update_aicenter_tags \
| save_definition "aicenter.pkgmanager"

echo "Updating aicenter.deployer definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/aifabric_/ai-deployer/v3/api-docs/aiCenterApi" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/aifabric_/ai-deployer" \
| update_aicenter_tags \
| save_definition "aicenter.deployer"

echo "Updating aicenter.trainer definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/aifabric_/ai-trainer/v3/api-docs/aiCenterApi" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/aifabric_/ai-trainer" \
| update_aicenter_tags \
| save_definition "aicenter.trainer"

echo "Updating aicenter.appmanager definition..."
download_definition "https://alpha.uipath.com/$organization/$tenant/aifabric_/ai-appmanager/v3/api-docs/aiCenterApi" \
| update_server_url "https://cloud.uipath.com/{organization}/{tenant}/aifabric_/ai-appmanager" \
| update_aicenter_tags \
| save_definition "aicenter.appmanager"

echo "Updating license.lrm definition..."
download_definition "https://alpha.uipath.com/$organization/license_/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/license_" \
| save_definition "license.lrm"

echo "Updating license.la definition..."
download_definition "https://alpha.uipath.com/$organization/lease_/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/lease_" \
| remove_clr_type \
| save_definition "license.la"

echo "Updating studio definition..."
download_definition "https://alpha.uipath.com/$organization/studio_/backend/swagger/v1/swagger.json" \
| update_server_url "https://cloud.uipath.com/{organization}/studio_/backend" \
| save_definition "studio"
