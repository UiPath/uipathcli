package api

import (
	"time"
)

type TestCaseExecution struct {
	Id                      int
	Status                  string
	TestCaseId              int
	VersionNumber           string
	EntryPointPath          string
	Info                    string
	StartTime               time.Time
	EndTime                 time.Time
	JobId                   int
	JobKey                  string
	InputArguments          string
	OutputArguments         string
	DataVariationIdentifier string
	Assertions              []TestCaseAssertion
	RobotLogs               []RobotLog
}

func (e *TestCaseExecution) SetRobotLogs(logs []RobotLog) {
	e.RobotLogs = logs
}

func NewTestCaseExecution(
	id int,
	status string,
	testCaseId int,
	versionNumber string,
	entryPointPath string,
	info string,
	startTime time.Time,
	endTime time.Time,
	jobId int,
	jobKey string,
	inputArguments string,
	outputArguments string,
	dataVariationIdentifier string,
	assertions []TestCaseAssertion,
	robotLogs []RobotLog) *TestCaseExecution {
	return &TestCaseExecution{
		id,
		status,
		testCaseId,
		versionNumber,
		entryPointPath,
		info,
		startTime,
		endTime,
		jobId,
		jobKey,
		inputArguments,
		outputArguments,
		dataVariationIdentifier,
		assertions,
		robotLogs,
	}
}

func (e TestCaseExecution) IsCompleted() bool {
	return e.Status == "Passed" || e.Status == "Failed" || e.Status == "Cancelled"
}
