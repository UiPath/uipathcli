package utils

import (
	"io"
	"os/exec"
)

type ExecDefaultCmd struct {
	Cmd *exec.Cmd
}

func (c ExecDefaultCmd) StdoutPipe() (io.ReadCloser, error) {
	return c.Cmd.StdoutPipe()
}

func (c ExecDefaultCmd) StderrPipe() (io.ReadCloser, error) {
	return c.Cmd.StderrPipe()
}

func (c ExecDefaultCmd) Start() error {
	return c.Cmd.Start()
}

func (c ExecDefaultCmd) Wait() error {
	return c.Cmd.Wait()
}

func (c ExecDefaultCmd) ExitCode() int {
	return c.Cmd.ProcessState.ExitCode()
}
