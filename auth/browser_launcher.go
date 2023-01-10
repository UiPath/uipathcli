package auth

type BrowserLauncher interface {
	OpenBrowser(url string) error
}
