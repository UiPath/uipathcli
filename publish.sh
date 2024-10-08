#! /bin/bash
############################################################
# Publishes a new version of the uipathcli on GitHub
#
# DESCRIPTION:
# This scripts publishes a tag with the latest version and
# creates a new release on GitHub.
############################################################

set -o pipefail
set -e

declare -r GITHUB_API_VERSION="2022-11-28"
declare -r OWNER="UiPath"
declare -r REPO="uipathcli"
declare -a RELEASE_FILES=(
    "build/packages/uipathcli-windows-amd64.zip"
    "build/packages/uipathcli-linux-amd64.tar.gz"
    "build/packages/uipathcli-darwin-amd64.tar.gz"
    "build/packages/uipathcli-windows-arm64.zip"
    "build/packages/uipathcli-linux-arm64.tar.gz"
    "build/packages/uipathcli-darwin-arm64.tar.gz"
)

############################################################
# Create new tag and push it to remote
#
# Arguments:
#   - The tag name
############################################################
function create_tag()
{
    local tag_name="$1"

    git tag "$tag_name" || true
    git push --tags --quiet
}

############################################################
# Create new release on GitHub
#
# Arguments:
#   - The new version (semver format, e.g. 1.0.1)
#
# Returns:
#   The release id from GitHub
############################################################
function create_release()
{
    local new_version="$1"

    response=$(curl --silent \
                    -H "Accept: application/vnd.github+json" \
                    -H "Authorization: Bearer $GITHUB_TOKEN"\
                    -H "X-GitHub-Api-Version: $GITHUB_API_VERSION" \
                    "https://api.github.com/repos/$OWNER/$REPO/releases" \
                    -d '{"tag_name":"'"$new_version"'","name":"uipathcli '"$new_version"'","generate_release_notes":true}')
    jq -r '.id' <<< "$response"
}

############################################################
# Upload a file for the given release
#
# Arguments:
#   - The GitHub release id
#   - The path to the file to upload
############################################################
function upload_release_file()
{
    local release_id="$1"
    local file_path="$2"

    local file_name
    local mime_type

    file_name=${file_path##*/}
    mime_type=$(file -b --mime-type "$file_path")

    curl --silent --output /dev/null \
         -H "Accept: application/vnd.github+json" \
         -H "Authorization: Bearer $GITHUB_TOKEN"\
         -H "X-GitHub-Api-Version: $GITHUB_API_VERSION" \
         -H "Content-Type: $mime_type" \
         --data-binary "@$file_path" \
         "https://uploads.github.com/repos/$OWNER/$REPO/releases/$release_id/assets?name=$file_name"
}

############################################################
# Main
############################################################
function main()
{
    local new_version="$1"

    if [ -z "${new_version}" ]; then
        echo "Please provide the new version to publish!"
        exit 1
    fi

    echo "==================================="
    echo "Releasing new version of uipathcli:"
    echo "$new_version"
    echo "==================================="

    local release_id
    create_tag "$new_version"
    release_id=$(create_release "$new_version")
    echo "Created new release '$new_version' with id '$release_id'"

    local release_file
    for release_file in "${RELEASE_FILES[@]}"
    do
        upload_release_file "$release_id" "$release_file"
        echo "Uploaded file '$release_file' for release '$new_version'"
    done

    echo "Successfully created release '$new_version'"
}
main "$1"
