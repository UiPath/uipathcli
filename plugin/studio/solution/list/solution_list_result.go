package list

import "github.com/UiPath/uipathcli/utils/api"

type solutionListResult struct {
	Status    string             `json:"status"`
	Solutions []api.SolutionInfo `json:"solutions"`
}
