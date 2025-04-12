package studio

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/process"
)

const uipcliCrossPlatformUrl = "https://github.com/UiPath/uipathcli/releases/download/plugins-v2.0.0/UiPath.CLI.24.12.9208.28468.nupkg"
const uipcliWindowsUrl = "https://github.com/UiPath/uipathcli/releases/download/plugins-v2.0.0/UiPath.CLI.Windows.24.12.9208.28468.nupkg"

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

func (c uipcli) ExecuteAndWait(args ...string) (int, string, error) {
	cmd, err := c.Execute(args...)
	if err != nil {
		return 1, "", err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return 1, "", fmt.Errorf("Could not run command: %v", err)
	}
	defer stdout.Close()
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return 1, "", fmt.Errorf("Could not run command: %v", err)
	}
	defer stderr.Close()
	err = cmd.Start()
	if err != nil {
		return 1, "", fmt.Errorf("Could not run command: %v", err)
	}

	stderrOutputBuilder := new(strings.Builder)
	stderrReader := io.TeeReader(stderr, stderrOutputBuilder)

	var wg sync.WaitGroup
	wg.Add(3)
	go c.readOutput(stdout, &wg)
	go c.readOutput(stderrReader, &wg)
	go c.wait(cmd, &wg)
	wg.Wait()

	return cmd.ExitCode(), stderrOutputBuilder.String(), nil
}

func (c uipcli) wait(cmd process.ExecCmd, wg *sync.WaitGroup) {
	defer wg.Done()
	_ = cmd.Wait()
}

func (c uipcli) readOutput(output io.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	scanner := bufio.NewScanner(output)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		c.Logger.Log(scanner.Text() + "\n")
	}
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
