//go:build windows

package auth

import (
	"github.com/UiPath/uipathcli/utils"
)

func (l BrowserLauncher) openBrowser(url string) utils.ExecCmd {
	return l.Exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
}
