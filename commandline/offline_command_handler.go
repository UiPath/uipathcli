package commandline

import (
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin"
)

// The offlineCommandHandler downloads all external resources to a local folder
// so that the uipathcli can be used without downloading any additional external
// dependencies. This can be useful in airgapped scenarios.
type offlineCommandHandler struct {
	Logger log.Logger
}

func (h offlineCommandHandler) Execute() error {
	manager := plugin.NewExternalPluginManager(h.Logger)
	return manager.Offline()
}

func newOfflineCommandHandler(logger log.Logger) *offlineCommandHandler {
	return &offlineCommandHandler{logger}
}
