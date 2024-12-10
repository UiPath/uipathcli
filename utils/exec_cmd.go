package utils

import "io"

type ExecCmd interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
	ExitCode() int
}
