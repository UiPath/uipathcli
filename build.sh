#! /bin/bash
set -e

echo "Copying OpenAPI definitions"
mkdir -p build/definitions
cp definitions/* build/definitions/
cp README.md build/README.md

echo "Building Linux (amd64) executable"
GOOS=linux GOARCH=amd64 go build -o build/uipathcli-linux-amd64

echo "Building Windows (amd64) executable"
GOOS=windows GOARCH=amd64 go build -o build/uipathcli-windows-amd64.exe

echo "Building MacOS (amd64) executable"
GOOS=darwin GOARCH=amd64 go build -o build/uipathcli-darwin-amd64

echo "Building Linux (arm64) executable"
GOOS=linux GOARCH=arm64 go build -o build/uipathcli-linux-arm64

echo "Building Windows (arm64) executable"
GOOS=windows GOARCH=arm64 go build -o build/uipathcli-windows-arm64.exe

echo "Building MacOS (arm64) executable"
GOOS=darwin GOARCH=arm64 go build -o build/uipathcli-darwin-arm64

echo "Successfully completed"