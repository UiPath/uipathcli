package commandline

import (
	"fmt"
	"io"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
)

// The offlineCommandHandler downloads all external resources to a local folder
// so that the uipathcli can be used without downloading any additional external
// dependencies. This can be useful in airgapped scenarios.
type offlineCommandHandler struct {
	StdOut io.Writer
	Logger log.Logger
}

func (h offlineCommandHandler) Execute() error {
	moduleManager := plugin.NewModuleManager(h.Logger)
	status, err := moduleManager.Offline()
	fmt.Fprint(h.StdOut, status+"\n\n")
	if err == nil {
		fmt.Fprint(h.StdOut, "Successfully downloaded all modules for offline mode!")
		return nil
	}
	return fmt.Errorf("Failed to download modules required for offline mode: %w", err)
}

func newOfflineCommandHandler(stdOut io.Writer, logger log.Logger) *offlineCommandHandler {
	return &offlineCommandHandler{stdOut, logger}
}
