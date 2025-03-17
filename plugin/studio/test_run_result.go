package studio

import "time"

type testRunResult struct {
	Status             string              `json:"status"`
	Id                 int                 `json:"id"`
	TestSetId          int                 `json:"testSetId"`
	Name               string              `json:"name"`
	StartTime          time.Time           `json:"startTime"`
	EndTime            time.Time           `json:"endTime"`
	TestCasesCount     int                 `json:"testCasesCount"`
	PassedCount        int                 `json:"passedCount"`
	FailuresCount      int                 `json:"failuresCount"`
	CanceledCount      int                 `json:"canceledCount"`
	TestCaseExecutions []testCaseRunResult `json:"testCaseExecutions"`
}

func newTestRunResult(execution TestExecution) *testRunResult {
	testCasesCount := 0
	passedCount := 0
	failuresCount := 0
	canceledCount := 0

	testCaseRunResults := []testCaseRunResult{}
	for _, execution := range execution.TestCaseExecutions {
		testCasesCount++
		if execution.Status == "Passed" {
			passedCount++
		} else if execution.Status == "Failed" {
			failuresCount++
		} else if execution.Status == "Cancelling" || execution.Status == "Cancelled" {
			canceledCount++
		}
		testCaseRunResults = append(testCaseRunResults, *newTestCaseRunResult(execution))
	}
	return &testRunResult{
		execution.Status,
		execution.Id,
		execution.TestSetId,
		execution.Name,
		execution.StartTime,
		execution.EndTime,
		testCasesCount,
		passedCount,
		failuresCount,
		canceledCount,
		testCaseRunResults,
	}
}
