package publish

type packagePublishResult struct {
	Status      string  `json:"status"`
	Package     string  `json:"package"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     string  `json:"version"`
	ReleaseId   *int    `json:"releaseId"`
	Error       *string `json:"error"`
}

func newSucceededPackagePublishResult(packagePath string, name string, description string, version string, releaseId int) *packagePublishResult {
	return &packagePublishResult{"Succeeded", packagePath, name, description, version, &releaseId, nil}
}

func newFailedPackagePublishResult(err string, packagePath string, name string, description string, version string) *packagePublishResult {
	return &packagePublishResult{"Failed", packagePath, name, description, version, nil, &err}
}
