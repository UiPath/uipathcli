#! /bin/bash
set -e

echo "Copying OpenAPI definitions"
mkdir -p build/definitions
cp definitions/* build/definitions/
cp README.md build/README.md

echo "Building linux executable"
GOOS=linux GOARCH=386 go build -o build/uipathcli

echo "Building windows executable"
GOOS=windows GOARCH=386 go build -o build/uipathcli.exe

echo "Building macos executable"
GOOS=darwin GOARCH=amd64 go build -o build/uipathcli.osx

echo "Successfully completed"