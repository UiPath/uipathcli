Write-Host "Copying README.md"
New-Item -ItemType Directory -Force -Path build | out-null
Copy-Item README.md -Destination build/README.md

Write-Host "Building Linux (amd64) executable"
pwsh -Command { $env:GOOS = "linux"; $env:GOARCH = "amd64"; go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$env:UIPATHCLI_VERSION" -o build/uipath-linux-amd64 }

Write-Host "Building Windows (amd64) executable"
pwsh -Command { $env:GOOS = "windows"; $env:GOARCH = "amd64"; go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$env:UIPATHCLI_VERSION" -o build/uipath-windows-amd64.exe }

Write-Host "Building MacOS (amd64) executable"
pwsh -Command { $env:GOOS = "darwin"; $env:GOARCH = "amd64"; go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$env:UIPATHCLI_VERSION" -o build/uipath-darwin-amd64 }

Write-Host "Building Linux (arm64) executable"
pwsh -Command { $env:GOOS = "linux"; $env:GOARCH = "arm64"; go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$env:UIPATHCLI_VERSION" -o build/uipath-linux-arm64 }

Write-Host "Building Windows (arm64) executable"
pwsh -Command { $env:GOOS = "windows"; $env:GOARCH = "arm64"; go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$env:UIPATHCLI_VERSION" -o build/uipath-windows-arm64.exe }

Write-Host "Building MacOS (arm64) executable"
pwsh -Command { $env:GOOS = "darwin"; $env:GOARCH = "arm64"; go build -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=$env:UIPATHCLI_VERSION" -o build//uipath-darwin-arm64 }

Write-Host "Successfully completed"