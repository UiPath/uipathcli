package pack

type solutionPackResult struct {
	Status     string `json:"status"`
	Package    string `json:"package"`
	SolutionId string `json:"solutionId"`
	Name       string `json:"name"`
	Size       int64  `json:"size"`
	Error      string `json:"error,omitempty"`
}

func newSucceededSolutionPackResult(packagePath string, solutionId string, name string, size int64) *solutionPackResult {
	return &solutionPackResult{"Succeeded", packagePath, solutionId, name, size, ""}
}
