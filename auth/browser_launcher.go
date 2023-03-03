package auth

// BrowserLauncher interface for opening browser windows.
type BrowserLauncher interface {
	Open(url string) error
}
