#! /bin/bash
set -e

echo "Copying README.md"
mkdir -p build
cp README.md build/README.md

echo "Building Linux (amd64) executable"
GOOS=linux GOARCH=amd64 go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$UIPATHCLI_VERSION" -o build/uipath-linux-amd64

echo "Building Windows (amd64) executable"
GOOS=windows GOARCH=amd64 go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$UIPATHCLI_VERSION" -o build/uipath-windows-amd64.exe

echo "Building MacOS (amd64) executable"
GOOS=darwin GOARCH=amd64 go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$UIPATHCLI_VERSION" -o build/uipath-darwin-amd64

echo "Building Linux (arm64) executable"
GOOS=linux GOARCH=arm64 go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$UIPATHCLI_VERSION" -o build/uipath-linux-arm64

echo "Building Windows (arm64) executable"
GOOS=windows GOARCH=arm64 go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$UIPATHCLI_VERSION" -o build/uipath-windows-arm64.exe

echo "Building MacOS (arm64) executable"
GOOS=darwin GOARCH=arm64 go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$UIPATHCLI_VERSION" -o build/uipath-darwin-arm64

echo "Successfully completed"