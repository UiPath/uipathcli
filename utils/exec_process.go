package utils

type ExecProcess interface {
	Command(name string, args ...string) ExecCmd
}
