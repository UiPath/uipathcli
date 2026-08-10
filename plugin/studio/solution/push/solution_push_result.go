package push

type solutionPushResult struct {
	Status     string `json:"status"`
	Package    string `json:"package"`
	SolutionId string `json:"solutionId"`
	Error      string `json:"error,omitempty"`
}

func newSucceededSolutionPushResult(packagePath string, solutionId string) *solutionPushResult {
	return &solutionPushResult{"Succeeded", packagePath, solutionId, ""}
}
