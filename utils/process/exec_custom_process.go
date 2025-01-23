package process

import (
	"io"
	"strings"
)

type ExecCustomProcess struct {
	ExitCode int
	Stdout   string
	Stderr   string
	OnStart  func(name string, args []string)
}

func (e ExecCustomProcess) Command(name string, args ...string) ExecCmd {
	return ExecCustomCmd{
		Name:    name,
		Args:    args,
		StdOut:  io.NopCloser(strings.NewReader(e.Stdout)),
		StdErr:  io.NopCloser(strings.NewReader(e.Stderr)),
		Exit:    e.ExitCode,
		OnStart: e.OnStart,
	}
}

func NewExecCustomProcess(exitCode int, stdout string, stderr string, onStart func(name string, args []string)) *ExecCustomProcess {
	return &ExecCustomProcess{
		ExitCode: exitCode,
		Stdout:   stdout,
		Stderr:   stderr,
		OnStart:  onStart,
	}
}
