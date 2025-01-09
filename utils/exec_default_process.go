package utils

import "os/exec"

type ExecDefaultProcess struct {
}

func (e ExecDefaultProcess) Command(name string, args ...string) ExecCmd {
	return ExecDefaultCmd{exec.Command(name, args...)}
}

func NewExecProcess() ExecProcess {
	return &ExecDefaultProcess{}
}
