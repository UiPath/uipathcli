package utils

import (
	"io"
)

type ExecCustomCmd struct {
	Name    string
	Args    []string
	Exit    int
	StdOut  io.ReadCloser
	StdErr  io.ReadCloser
	OnStart func(name string, args []string)
}

func (c ExecCustomCmd) StdoutPipe() (io.ReadCloser, error) {
	return c.StdOut, nil
}

func (c ExecCustomCmd) StderrPipe() (io.ReadCloser, error) {
	return c.StdErr, nil
}

func (c ExecCustomCmd) Start() error {
	c.OnStart(c.Name, c.Args)
	return nil
}

func (c ExecCustomCmd) Wait() error {
	return nil
}

func (c ExecCustomCmd) ExitCode() int {
	return c.Exit
}
