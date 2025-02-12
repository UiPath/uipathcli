package studio

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/process"
)

const uipcliCrossPlatformVersion = "24.12.9111.31003"
const uipcliCrossPlatformUrl = "https://uipath.pkgs.visualstudio.com/Public.Feeds/_apis/packaging/feeds/1c781268-d43d-45ab-9dfc-0151a1c740b7/nuget/packages/UiPath.CLI/versions/" + uipcliCrossPlatformVersion + "/content"

const uipcliWindowsVersion = "24.12.9111.31003"
const uipcliWindowsUrl = "https://uipath.pkgs.visualstudio.com/Public.Feeds/_apis/packaging/feeds/1c781268-d43d-45ab-9dfc-0151a1c740b7/nuget/packages/UiPath.CLI.Windows/versions/" + uipcliWindowsVersion + "/content"

const dotnetLinuxX64Url = "https://aka.ms/dotnet/8.0/dotnet-runtime-linux-x64.tar.gz"
const dotnetMacOsX64Url = "https://aka.ms/dotnet/8.0/dotnet-runtime-osx-x64.tar.gz"
const dotnetWindowsX64Url = "https://aka.ms/dotnet/8.0/dotnet-runtime-win-x64.zip"

const dotnetLinuxArm64Url = "https://aka.ms/dotnet/8.0/dotnet-runtime-linux-arm64.tar.gz"
const dotnetMacOsArm64Url = "https://aka.ms/dotnet/8.0/dotnet-runtime-osx-arm64.tar.gz"
const dotnetWindowsArm64Url = "https://aka.ms/dotnet/8.0/dotnet-runtime-win-arm64.zip"

type uipcli struct {
	Exec   process.ExecProcess
	Logger log.Logger
	path   string
	args   []string
}

func (c *uipcli) Initialize(targetFramework TargetFramework) error {
	uipcliPath, err := c.getUipcliPath(targetFramework)
	if err != nil {
		return err
	}
	c.path = uipcliPath

	if filepath.Ext(c.path) == ".dll" {
		dotnetPath, err := c.getDotnetPath()
		if err != nil {
			return err
		}
		c.path = dotnetPath
		c.args = []string{uipcliPath}
	}
	return nil
}

func (c uipcli) Execute(args ...string) (process.ExecCmd, error) {
	args = append(c.args, args...)
	cmd := c.Exec.Command(c.path, args...)
	return cmd, nil
}

func (c uipcli) getUipcliPath(targetFramework TargetFramework) (string, error) {
	externalPlugin := plugin.NewExternalPlugin(c.Logger)
	name := "uipcli"
	url := uipcliCrossPlatformUrl
	executable := "tools/uipcli.dll"
	if targetFramework.IsWindowsOnly() {
		name = "uipcli-win"
		url = uipcliWindowsUrl
		executable = "tools/uipcli.exe"
	}
	return externalPlugin.GetTool(name, url, plugin.ArchiveTypeZip, executable)
}

func (c uipcli) getDotnetPath() (string, error) {
	externalPlugin := plugin.NewExternalPlugin(c.Logger)
	name := fmt.Sprintf("dotnet8-%s-%s", runtime.GOOS, runtime.GOARCH)
	url, archiveType, executable := c.dotnetUrl()
	return externalPlugin.GetTool(name, url, archiveType, executable)
}

func (c uipcli) dotnetUrl() (string, plugin.ArchiveType, string) {
	if c.isArm() {
		switch runtime.GOOS {
		case "windows":
			return dotnetWindowsArm64Url, plugin.ArchiveTypeZip, "dotnet.exe"
		case "darwin":
			return dotnetMacOsArm64Url, plugin.ArchiveTypeTarGz, "dotnet"
		default:
			return dotnetLinuxArm64Url, plugin.ArchiveTypeTarGz, "dotnet"
		}
	}
	switch runtime.GOOS {
	case "windows":
		return dotnetWindowsX64Url, plugin.ArchiveTypeZip, "dotnet.exe"
	case "darwin":
		return dotnetMacOsX64Url, plugin.ArchiveTypeTarGz, "dotnet"
	default:
		return dotnetLinuxX64Url, plugin.ArchiveTypeTarGz, "dotnet"
	}
}

func (c uipcli) isArm() bool {
	return strings.HasPrefix(strings.ToLower(runtime.GOARCH), "arm")
}

func newUipcli(exec process.ExecProcess, logger log.Logger) *uipcli {
	return &uipcli{exec, logger, "", []string{}}
}
