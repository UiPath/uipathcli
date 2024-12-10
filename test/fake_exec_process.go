package test

import (
	"io"
	"strings"

	"github.com/UiPath/uipathcli/utils"
)

type FakeExecCmd struct {
	StdOut io.ReadCloser
	StdErr io.ReadCloser
	Exit   int
}

func (c FakeExecCmd) StdoutPipe() (io.ReadCloser, error) {
	return c.StdOut, nil
}

func (c FakeExecCmd) StderrPipe() (io.ReadCloser, error) {
	return c.StdErr, nil
}

func (c FakeExecCmd) Start() error {
	return nil
}

func (c FakeExecCmd) Wait() error {
	return nil
}

func (c FakeExecCmd) ExitCode() int {
	return c.Exit
}

type FakeExecProcess struct {
	Cmd utils.ExecCmd
}

func (e FakeExecProcess) Command(name string, args ...string) utils.ExecCmd {
	return e.Cmd
}

func NewFakeExecProcess(exitCode int, stdout string, stderr string) *FakeExecProcess {
	return &FakeExecProcess{
		Cmd: FakeExecCmd{
			StdOut: io.NopCloser(strings.NewReader(stdout)),
			StdErr: io.NopCloser(strings.NewReader(stderr)),
			Exit:   exitCode,
		},
	}
}
