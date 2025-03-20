package studio

import (
	"bufio"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
	"github.com/UiPath/uipathcli/utils/process"
)

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
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		c.Logger.Log(scanner.Text())
	}
}

func (c uipcli) getUipcliPath(targetFramework TargetFramework) (string, error) {
	manager := plugin.NewExternalPluginManager(c.Logger)
	if targetFramework.IsWindowsOnly() {
		return manager.Get(plugin.UipCliWindows)
	}
	return manager.Get(plugin.UipCliCrossPlatform)
}

func (c uipcli) getDotnetPath() (string, error) {
	manager := plugin.NewExternalPluginManager(c.Logger)
	return manager.Get(plugin.DotNet8)
}

func newUipcli(exec process.ExecProcess, logger log.Logger) *uipcli {
	return &uipcli{exec, logger, "", []string{}}
}
