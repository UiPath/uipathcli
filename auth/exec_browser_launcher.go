package auth

import (
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

type ExecBrowserLauncher struct{}

func (l ExecBrowserLauncher) OpenBrowser(url string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		return fmt.Errorf("Platform not supported: %s", runtime.GOOS)
	}

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
