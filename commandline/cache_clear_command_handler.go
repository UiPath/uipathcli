package commandline

import (
	"fmt"
	"io"
	"os"

	"github.com/UiPath/uipathcli/utils/directories"
)

// cacheClearCommandHandler implements command to clear auth token cache
type cacheClearCommandHandler struct {
	StdOut io.Writer
}

func (h cacheClearCommandHandler) Clear() error {
	cacheDirectory, err := directories.Cache()
	if err != nil {
		return err
	}
	err = os.RemoveAll(cacheDirectory)
	if err != nil {
		return fmt.Errorf("Could not clear cache: %v", err)
	}
	fmt.Fprintln(h.StdOut, "Cache has been successfully cleared")
	return nil
}

func newCacheClearCommandHandler(stdOut io.Writer) *cacheClearCommandHandler {
	return &cacheClearCommandHandler{
		StdOut: stdOut,
	}
}
