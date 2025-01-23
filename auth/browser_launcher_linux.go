//go:build linux

package auth

import "github.com/UiPath/uipathcli/utils/process"

func (l BrowserLauncher) openBrowser(url string) process.ExecCmd {
	return l.Exec.Command("xdg-open", url)
}
