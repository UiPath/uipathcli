package api

import "time"

type TestExecution struct {
	Id                 int
	Status             string
	TestSetId          int
	Name               string
	StartTime          time.Time
	EndTime            time.Time
	TestCaseExecutions []TestCaseExecution
}

func NewTestExecution(id int, status string, testSetId int, name string, startTime time.Time, endTime time.Time, testCaseExecutions []TestCaseExecution) *TestExecution {
	return &TestExecution{id, status, testSetId, name, startTime, endTime, testCaseExecutions}
}

func (e TestExecution) IsCompleted() bool {
	return e.Status == "Passed" || e.Status == "Failed" || e.Status == "Cancelled"
}
