// Package process provides abstractions for executiong processes
// which makes it easier to call executables as well as faking
// them during testing.
package process

import "io"

type ExecCmd interface {
	StdoutPipe() (io.ReadCloser, error)
	StderrPipe() (io.ReadCloser, error)
	Start() error
	Wait() error
	ExitCode() int
}
