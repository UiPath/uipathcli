package studio

import "time"

type TestCaseExecution struct {
	Id             int
	Status         string
	TestCaseId     int
	EntryPointPath string
	Info           string
	StartTime      time.Time
	EndTime        time.Time
}

func NewTestCaseExecution(id int, status string, testCaseId int, entryPointPath string, info string, startTime time.Time, endTime time.Time) *TestCaseExecution {
	return &TestCaseExecution{id, status, testCaseId, entryPointPath, info, startTime, endTime}
}

func (e TestCaseExecution) IsCompleted() bool {
	return e.Status == "Passed" || e.Status == "Failed" || e.Status == "Cancelled"
}
