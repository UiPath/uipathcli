package studio

import (
	"time"

	"github.com/UiPath/uipathcli/utils/api"
)

type testCaseRunResult struct {
	Status     string    `json:"status"`
	Id         int       `json:"id"`
	TestCaseId int       `json:"testCaseId"`
	Name       string    `json:"name"`
	Error      *string   `json:"error"`
	StartTime  time.Time `json:"startTime"`
	EndTime    time.Time `json:"endTime"`
}

func newTestCaseRunResult(execution api.TestCaseExecution) *testCaseRunResult {
	var err *string
	if execution.Status == "Failed" && execution.Info != "" {
		err = &execution.Info
	}
	return &testCaseRunResult{
		execution.Status,
		execution.Id,
		execution.TestCaseId,
		execution.EntryPointPath,
		err,
		execution.StartTime,
		execution.EndTime}
}
