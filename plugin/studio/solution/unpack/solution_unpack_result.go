package unpack

type solutionUnpackResult struct {
	Status       string `json:"status"`
	Directory    string `json:"directory"`
	SolutionId   string `json:"solutionId"`
	ProjectCount int    `json:"projectCount"`
	Error        string `json:"error,omitempty"`
}

func newSucceededSolutionUnpackResult(directory string, solutionId string, projectCount int) *solutionUnpackResult {
	return &solutionUnpackResult{"Succeeded", directory, solutionId, projectCount, ""}
}
