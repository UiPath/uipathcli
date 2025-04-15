package pack

type packagePackResult struct {
	Status      string  `json:"status"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ProjectId   *string `json:"projectId"`
	Version     *string `json:"version"`
	Output      *string `json:"output"`
	Error       *string `json:"error"`
}

func newSucceededPackagePackResult(output string, name string, description string, projectId string, version string) *packagePackResult {
	return &packagePackResult{"Succeeded", &name, &description, &projectId, &version, &output, nil}
}

func newFailedPackagePackResult(err string, name *string, description *string, projectId *string) *packagePackResult {
	return &packagePackResult{"Failed", name, description, projectId, nil, nil, &err}
}
