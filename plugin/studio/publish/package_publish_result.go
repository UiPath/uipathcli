package publish

type packagePublishResult struct {
	Status  string  `json:"status"`
	Package string  `json:"package"`
	Name    string  `json:"name"`
	Version string  `json:"version"`
	Error   *string `json:"error"`
}

func newSucceededPackagePublishResult(packagePath string, name string, version string) *packagePublishResult {
	return &packagePublishResult{"Succeeded", packagePath, name, version, nil}
}

func newFailedPackagePublishResult(err string, packagePath string, processKey string, processVersion string) *packagePublishResult {
	return &packagePublishResult{"Failed", packagePath, processKey, processVersion, &err}
}
