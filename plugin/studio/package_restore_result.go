package studio

type packageRestoreResult struct {
	Status      string  `json:"status"`
	Name        *string `json:"name"`
	Description *string `json:"description"`
	ProjectId   *string `json:"projectId"`
	Output      *string `json:"output"`
	Error       *string `json:"error"`
}

func newSucceededPackageRestoreResult(output string, name string, description string, projectId string) *packageRestoreResult {
	return &packageRestoreResult{"Succeeded", &name, &description, &projectId, &output, nil}
}

func newFailedPackageRestoreResult(err string, name *string, description *string, projectId *string) *packageRestoreResult {
	return &packageRestoreResult{"Failed", name, description, projectId, nil, &err}
}
