package plugin

import "runtime"

const UipCliCrossPlatform = "uipcli"
const UipCliWindows = "uipcli-win"
const DotNet8 = "dotnet8-" + runtime.GOOS + "-" + runtime.GOARCH

var AvailableModules = map[string]ModuleDefinition{
	// uipcli - crossplatform
	UipCliCrossPlatform: *NewModuleDefinition(
		UipCliCrossPlatform,
		"https://github.com/UiPath/uipathcli/releases/download/plugins-v2.0.0/UiPath.CLI.24.12.9208.28468.nupkg",
		"UiPath.CLI.24.12.9208.28468.nupkg",
		ArchiveTypeZip,
		"tools/uipcli.dll",
	),
	// uipcli - windows
	UipCliWindows: *NewModuleDefinition(
		UipCliWindows,
		"https://github.com/UiPath/uipathcli/releases/download/plugins-v2.0.0/UiPath.CLI.Windows.24.12.9208.28468.nupkg",
		"UiPath.CLI.Windows.24.12.9208.28468.nupkg",
		ArchiveTypeZip,
		"tools/uipcli.exe",
	),

	// dotnet8 - x64
	"dotnet8-windows-amd64": *NewModuleDefinition(
		"dotnet8-windows-amd64",
		"https://aka.ms/dotnet/8.0/dotnet-runtime-win-x64.zip",
		"dotnet-runtime-win-x64.zip",
		ArchiveTypeZip,
		"dotnet.exe",
	),
	"dotnet8-darwin-amd64": *NewModuleDefinition(
		"dotnet8-darwin-amd64",
		"https://aka.ms/dotnet/8.0/dotnet-runtime-osx-x64.tar.gz",
		"dotnet-runtime-osx-x64.tar.gz",
		ArchiveTypeTarGz,
		"dotnet",
	),
	"dotnet8-linux-amd64": *NewModuleDefinition(
		"dotnet8-linux-amd64",
		"https://aka.ms/dotnet/8.0/dotnet-runtime-linux-x64.tar.gz",
		"dotnet-runtime-linux-x64.tar.gz",
		ArchiveTypeTarGz,
		"dotnet",
	),

	// dotnet8 - arm64
	"dotnet8-windows-arm64": *NewModuleDefinition(
		"dotnet8-windows-arm64",
		"https://aka.ms/dotnet/8.0/dotnet-runtime-win-arm64.zip",
		"dotnet-runtime-win-arm64.zip",
		ArchiveTypeZip,
		"dotnet.exe",
	),
	"dotnet8-darwin-arm64": *NewModuleDefinition(
		"dotnet8-darwin-arm64",
		"https://aka.ms/dotnet/8.0/dotnet-runtime-osx-arm64.tar.gz",
		"dotnet-runtime-osx-arm64.tar.gz",
		ArchiveTypeTarGz,
		"dotnet",
	),
	"dotnet8-linux-arm64": *NewModuleDefinition(
		"dotnet8-linux-arm64",
		"https://aka.ms/dotnet/8.0/dotnet-runtime-linux-arm64.tar.gz",
		"dotnet-runtime-linux-arm64.tar.gz",
		ArchiveTypeTarGz,
		"dotnet",
	),
}

type ModuleDefinition struct {
	Name        string
	Url         string
	ArchiveName string
	ArchiveType ArchiveType
	Executable  string
}

func NewModuleDefinition(name string, url string, archiveName string, archiveType ArchiveType, executable string) *ModuleDefinition {
	return &ModuleDefinition{name, url, archiveName, archiveType, executable}
}
