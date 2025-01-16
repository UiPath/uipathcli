package commandline

import (
	"fmt"
	"io"
	"runtime"

	"github.com/UiPath/uipathcli/utils"
)

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
	fmt.Fprintf(h.StdOut, "uipathcli %s (%s, %s)\n", utils.Version, runtime.GOOS, runtime.GOARCH)
}

func newVersionCommandHandler(stdOut io.Writer) *versionCommandHandler {
	return &versionCommandHandler{stdOut}
}
