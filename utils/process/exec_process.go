package process

type ExecProcess interface {
	Command(name string, args ...string) ExecCmd
}
