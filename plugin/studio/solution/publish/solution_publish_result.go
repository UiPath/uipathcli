package publish

type solutionPublishResult struct {
	Status    string `json:"status"`
	RequestId string `json:"requestId"`
	Error     string `json:"error,omitempty"`
}
