#! /bin/bash
############################################################
# Generates the next version to publish
#
# DESCRIPTION:
# This scripts retrieves the latest tag, increments the
# version and creates a new release on GitHub.
############################################################

set -o pipefail
set -e

############################################################
# Retrieves the latest version by sorting the git tags
#
# Returns:
#   The latest git tag
############################################################
function get_latest_version()
{
    local version_filter="$1"

    git fetch --all --tags --quiet
    git tag | sort -V | grep "^$version_filter.*" | tail -1 || echo "$version_filter"
}

############################################################
# Increment patch version on the provided semver string
#
# Arguments:
#   - The version (semver format, e.g. 1.0.0)
#
# Returns:
#   Incremented patch version (e.g. 1.0.1)
############################################################
function increment_patch_version()
{
    local version="$1"

    local array
    local IFS='.'; read -r -a array <<< "$version"
    if [ -z "${array[2]}" ]; then
        array[2]="0"
    else
        array[2]=$((array[2]+1))
    fi
    echo "$(local IFS='.'; echo "${array[*]}")"
}

############################################################
# Create new tag
#
# Arguments:
#   - The tag name
############################################################
function create_tag()
{
    local tag_name="$1"

    git tag "$tag_name"
}

############################################################
# Main
############################################################
function main()
{
    local base_version="$1"
    local latest_version
    local new_version

    latest_version=$(get_latest_version "$base_version")
    new_version=$(increment_patch_version "$latest_version")
    echo "$new_version"
}
main "$1"
