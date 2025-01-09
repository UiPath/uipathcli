package studio

type packageAnalyzeResult struct {
	Status     string                    `json:"status"`
	Violations []packageAnalyzeViolation `json:"violations"`
	Error      *string                   `json:"error"`
}

type packageAnalyzeViolation struct {
	ErrorCode           string                    `json:"errorCode"`
	Description         string                    `json:"description"`
	RuleName            string                    `json:"ruleName"`
	FilePath            string                    `json:"filePath"`
	ActivityId          *packageAnalyzeActivityId `json:"activityId"`
	ActivityDisplayName string                    `json:"activityDisplayName"`
	WorkflowDisplayName string                    `json:"workflowDisplayName"`
	Item                *packageAnalyzeItem       `json:"item"`
	ErrorSeverity       int                       `json:"errorSeverity"`
	Recommendation      string                    `json:"recommendation"`
	DocumentationLink   string                    `json:"documentationLink"`
}
type packageAnalyzeActivityId struct {
	Id    string `json:"id"`
	IdRef string `json:"idRef"`
}

type packageAnalyzeItem struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

func newSucceededPackageAnalyzeResult(violations []packageAnalyzeViolation) *packageAnalyzeResult {
	return &packageAnalyzeResult{"Succeeded", violations, nil}
}

func newFailedPackageAnalyzeResult(violations []packageAnalyzeViolation, err string) *packageAnalyzeResult {
	return &packageAnalyzeResult{"Failed", violations, &err}
}
