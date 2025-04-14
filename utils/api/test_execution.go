package api

import "time"

type TestExecution struct {
	Id                 int
	Status             string
	TestSetId          int
	Name               string
	StartTime          time.Time
	EndTime            time.Time
	TestCasesCount     int
	PassedCount        int
	FailuresCount      int
	CanceledCount      int
	TestCaseExecutions []TestCaseExecution
}

func NewTestExecution(
	id int,
	status string,
	testSetId int,
	name string,
	startTime time.Time,
	endTime time.Time,
	testCaseExecutions []TestCaseExecution) *TestExecution {
	testCasesCount := 0
	passedCount := 0
	failuresCount := 0
	canceledCount := 0

	for _, execution := range testCaseExecutions {
		testCasesCount++
		if execution.Status == "Passed" {
			passedCount++
		} else if execution.Status == "Failed" {
			failuresCount++
		} else if execution.Status == "Cancelling" || execution.Status == "Cancelled" {
			canceledCount++
		}
	}

	return &TestExecution{
		id,
		status,
		testSetId,
		name,
		startTime,
		endTime,
		testCasesCount,
		passedCount,
		failuresCount,
		canceledCount,
		testCaseExecutions,
	}
}

func (e TestExecution) IsCompleted() bool {
	return e.Status == "Passed" || e.Status == "Failed" || e.Status == "Cancelled"
}
