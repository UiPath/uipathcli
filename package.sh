#! /bin/bash
set -e

mkdir -p build/packages
rm --force build/packages/*

pushd build/ > /dev/null

echo "Create Linux (amd64) Package"
tar --create --gzip --overwrite --file=packages/uipathcli-linux-amd64.tar.gz --transform='flags=r;s|uipathcli-linux-amd64|uipathcli|' uipathcli-linux-amd64 definitions

echo "Create Windows (amd64) Package"
zip -q packages/uipathcli-windows-amd64.zip uipathcli-windows-amd64.exe definitions/*
printf "@ uipathcli-windows-amd64.exe\n@=uipathcli.exe\n" | zipnote -w packages/uipathcli-windows-amd64.zip

echo "Create MacOS (amd64) Package"
tar --create --gzip --overwrite --file=packages/uipathcli-darwin-amd64.tar.gz --transform='flags=r;s|uipathcli-darwin-amd64|uipathcli|' uipathcli-darwin-amd64 definitions

echo "Create Linux (arm64) Package"
tar --create --gzip --overwrite --file=packages/uipathcli-linux-arm64.tar.gz --transform='flags=r;s|uipathcli-linux-arm64|uipathcli|' uipathcli-linux-arm64 definitions

echo "Create Windows (arm64) Package"
zip -q packages/uipathcli-windows-arm64.zip uipathcli-windows-arm64.exe definitions/*
printf "@ uipathcli-windows-arm64.exe\n@=uipathcli.exe\n" | zipnote -w packages/uipathcli-windows-arm64.zip

echo "Create MacOS (arm64) Package"
tar --create --gzip --overwrite --file=packages/uipathcli-darwin-arm64.tar.gz --transform='flags=r;s|uipathcli-darwin-arm64|uipathcli|' uipathcli-darwin-arm64 definitions


popd > /dev/null
echo "Successfully created packages"