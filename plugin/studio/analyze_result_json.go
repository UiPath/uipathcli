package studio

type analyzeResultJson []struct {
	ErrorCode           string                   `json:"ErrorCode"`
	Description         string                   `json:"Description"`
	RuleName            string                   `json:"RuleName"`
	FilePath            string                   `json:"FilePath"`
	ActivityId          *analyzeResultActivityId `json:"ActivityId"`
	ActivityDisplayName string                   `json:"ActivityDisplayName"`
	WorkflowDisplayName string                   `json:"WorkflowDisplayName"`
	Item                *analyzeResultItem       `json:"Item"`
	ErrorSeverity       int                      `json:"ErrorSeverity"`
	Recommendation      string                   `json:"Recommendation"`
	DocumentationLink   string                   `json:"DocumentationLink"`
}

type analyzeResultActivityId struct {
	Id    string `json:"Id"`
	IdRef string `json:"IdRef"`
}

type analyzeResultItem struct {
	Name string `json:"Name"`
	Type int    `json:"Type"`
}
