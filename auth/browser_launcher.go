package auth

import (
	"fmt"
	"time"

	"github.com/UiPath/uipathcli/utils"
)

// BrowserLauncher tries to open the default browser on the local system.
type BrowserLauncher struct {
	Exec utils.ExecProcess
}

func (l BrowserLauncher) Open(url string) error {
	cmd := l.openBrowser(url)
	err := cmd.Start()
	if err != nil {
		return err
	}
	done := make(chan error)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Second):
		return fmt.Errorf("Timed out waiting for browser to start")
	}
}

func NewBrowserLauncher() *BrowserLauncher {
	return &BrowserLauncher{utils.NewExecProcess()}
}
