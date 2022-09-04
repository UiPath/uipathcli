Write-Host "Copying OpenAPI definitions"
New-Item -ItemType Directory -Force -Path build/definitions | out-null
Copy-Item definitions/* -Destination build/definitions/

Write-Host "Building linux executable"
$env:GOOS = "linux"; $env:GOARCH = "386"; go build -o build/uipathcli

Write-Host "Building windows executable"
$env:GOOS = "windows"; $env:GOARCH = "386"; go build -o build/uipathcli.exe

Write-Host "Building macos executable"
$env:GOOS = "darwin"; $env:GOARCH = "amd64"; go build -o build/uipathcli.osx

Write-Host "Successfully completed"