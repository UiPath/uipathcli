package commandline

import (
	"fmt"
	"io"
	"runtime"
)

// This Version variable is overridden during build time
// by providing the linker flag:
// -ldflags="-X github.com/UiPath/uipathcli/commandline.Version=1.2.3"
var Version = "main"

// The VersionCommandHandler outputs the build information
//
// Example:
// uipath --version
//
// uipathcli v1.0.0 (windows, amd64)
type versionCommandHandler struct {
	StdOut io.Writer
}

func (h versionCommandHandler) Execute() {
	fmt.Fprintf(h.StdOut, "uipathcli %s (%s, %s)\n", Version, runtime.GOOS, runtime.GOARCH)
}

func newVersionCommandHandler(stdOut io.Writer) *versionCommandHandler {
	return &versionCommandHandler{stdOut}
}
