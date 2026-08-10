package pull

type solutionPullResult struct {
	Status     string `json:"status"`
	File       string `json:"file"`
	SolutionId string `json:"solutionId"`
	Size       int64  `json:"size"`
	Error      string `json:"error,omitempty"`
}

func newSucceededSolutionPullResult(filePath string, solutionId string, size int64) *solutionPullResult {
	return &solutionPullResult{"Succeeded", filePath, solutionId, size, ""}
}
