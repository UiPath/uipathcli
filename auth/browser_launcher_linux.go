//go:build linux

package auth

import (
	"github.com/UiPath/uipathcli/utils"
)

func (l BrowserLauncher) openBrowser(url string) utils.ExecCmd {
	return l.Exec.Command("xdg-open", url)
}
