package studio

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/process"
)

const uipcliVersion = "24.12.9111.31003"
const uipcliUrl = "https://uipath.pkgs.visualstudio.com/Public.Feeds/_apis/packaging/feeds/1c781268-d43d-45ab-9dfc-0151a1c740b7/nuget/packages/UiPath.CLI/versions/" + uipcliVersion + "/content"

const uipcliWindowsVersion = "24.12.9111.31003"
const uipcliWindowsUrl = "https://uipath.pkgs.visualstudio.com/Public.Feeds/_apis/packaging/feeds/1c781268-d43d-45ab-9dfc-0151a1c740b7/nuget/packages/UiPath.CLI.Windows/versions/" + uipcliWindowsVersion + "/content"

type uipcli struct {
	Exec   process.ExecProcess
	Logger log.Logger
}

func (c uipcli) Execute(targetFramework TargetFramework, args ...string) (process.ExecCmd, error) {
	if targetFramework == TargetFrameworkWindows {
		return c.execute("uipcli-win", uipcliWindowsUrl, args)
	}
	return c.execute("uipcli", uipcliUrl, args)
}

func (c uipcli) execute(name string, url string, args []string) (process.ExecCmd, error) {
	uipcliPath, err := c.getPath(name, url)
	if err != nil {
		return nil, err
	}

	path := uipcliPath
	if filepath.Ext(uipcliPath) == ".dll" {
		path, err = exec.LookPath("dotnet")
		if err != nil {
			return nil, fmt.Errorf("Could not find dotnet runtime to run command: %v", err)
		}
		args = append([]string{uipcliPath}, args...)
	}

	cmd := c.Exec.Command(path, args...)
	return cmd, nil
}

func (c uipcli) getPath(name string, url string) (string, error) {
	externalPlugin := plugin.NewExternalPlugin(c.Logger)
	executable := "tools/uipcli.dll"
	if c.isWindows() {
		executable = "tools/uipcli.exe"
	}
	return externalPlugin.GetTool(name, url, executable)
}

func (c uipcli) isWindows() bool {
	return runtime.GOOS == "windows"
}

func newUipcli(exec process.ExecProcess, logger log.Logger) *uipcli {
	return &uipcli{exec, logger}
}
