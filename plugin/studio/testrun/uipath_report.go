package testrun

import "time"

type uipathReport struct {
	TestSetExecutions []uipathReportTestSetExecution `json:"testSetExecutions"`
}

type uipathReportTestSetExecution struct {
	Id                 int                             `json:"id"`
	TestSetId          int                             `json:"testSetId"`
	Name               string                          `json:"name"`
	Status             string                          `json:"status"`
	StartTime          time.Time                       `json:"startTime"`
	EndTime            time.Time                       `json:"endTime"`
	TestCasesCount     int                             `json:"testCasesCount"`
	PassedCount        int                             `json:"passedCount"`
	FailuresCount      int                             `json:"failuresCount"`
	CanceledCount      int                             `json:"canceledCount"`
	Packages           []uipathReportTestPackage       `json:"packages"`
	TestCaseExecutions []uipathReportTestCaseExecution `json:"testCaseExecutions"`
}

type uipathReportTestPackage struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type uipathReportTestCaseExecution struct {
	Id                      int                             `json:"id"`
	TestCaseId              int                             `json:"testCaseId"`
	Name                    string                          `json:"name"`
	Status                  string                          `json:"status"`
	Error                   *string                         `json:"error"`
	StartTime               time.Time                       `json:"startTime"`
	EndTime                 time.Time                       `json:"endTime"`
	JobId                   int                             `json:"jobId"`
	VersionNumber           string                          `json:"versionNumber"`
	PackageIdentifier       string                          `json:"packageIdentifier"`
	EntryPointPath          string                          `json:"entryPointPath"`
	InputArguments          string                          `json:"inputArguments"`
	OutputArguments         string                          `json:"outputArguments"`
	DataVariationIdentifier string                          `json:"dataVariationIdentifier"`
	Assertions              []uipathReportTestCaseAssertion `json:"assertions"`
	RobotLogs               []uipathRobotLog                `json:"robotLogs,omitempty"`
}

type uipathReportTestCaseAssertion struct {
	Message   string `json:"message"`
	Succeeded bool   `json:"succeeded"`
}

type uipathRobotLog struct {
	Id              int       `json:"id"`
	Level           string    `json:"level"`
	WindowsIdentity string    `json:"windowsIdentity"`
	ProcessName     string    `json:"processName"`
	TimeStamp       time.Time `json:"timeStamp"`
	Message         string    `json:"message"`
	RobotName       string    `json:"robotName"`
	HostMachineName string    `json:"hostMachineName"`
	MachineId       int       `json:"machineId"`
	MachineKey      string    `json:"machineKey"`
	RuntimeType     string    `json:"runtimeType"`
}
