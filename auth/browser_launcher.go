package auth

import (
	"errors"
	"time"

	"github.com/UiPath/uipathcli/utils/process"
)

// BrowserLauncher tries to open the default browser on the local system.
type BrowserLauncher struct {
	Exec process.ExecProcess
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
		return errors.New("Timed out waiting for browser to start")
	}
}

func NewBrowserLauncher() *BrowserLauncher {
	return &BrowserLauncher{process.NewExecProcess()}
}
