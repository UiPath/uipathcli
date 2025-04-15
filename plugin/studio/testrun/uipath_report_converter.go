package testrun

import (
	"github.com/UiPath/uipathcli/utils/api"
)

type uipathReportConverter struct {
}

func (c uipathReportConverter) Convert(status []testRunStatus) uipathReport {
	executions := []uipathReportTestSetExecution{}
	for _, s := range status {
		executions = append(executions, c.convertTestSetExecution(s))
	}
	return uipathReport{executions}
}

func (c uipathReportConverter) convertTestSetExecution(status testRunStatus) uipathReportTestSetExecution {
	testCaseExecutions := []uipathReportTestCaseExecution{}
	for _, execution := range status.Execution.TestCaseExecutions {
		testCase := c.findTestCase(status.TestSet, execution.TestCaseId)
		testCaseExecutions = append(testCaseExecutions, c.convertTestCaseExecution(testCase, execution))
	}

	testPackages := []uipathReportTestPackage{}
	for _, p := range status.TestSet.Packages {
		testPackages = append(testPackages, uipathReportTestPackage{p.PackageIdentifier, p.VersionMask})
	}

	return uipathReportTestSetExecution{
		status.Execution.Id,
		status.Execution.TestSetId,
		status.Execution.Name,
		status.Execution.Status,
		status.Execution.StartTime,
		status.Execution.EndTime,
		status.Execution.TestCasesCount,
		status.Execution.PassedCount,
		status.Execution.FailuresCount,
		status.Execution.CanceledCount,
		testPackages,
		testCaseExecutions,
	}
}

func (c uipathReportConverter) convertTestCaseExecution(testCase api.TestCase, execution api.TestCaseExecution) uipathReportTestCaseExecution {
	var err *string
	if execution.Status == "Failed" && execution.Info != "" {
		err = &execution.Info
	}
	return uipathReportTestCaseExecution{
		execution.Id,
		execution.TestCaseId,
		testCase.Name,
		execution.Status,
		err,
		execution.StartTime,
		execution.EndTime,
		execution.JobId,
		execution.VersionNumber,
		testCase.PackageIdentifier,
		execution.EntryPointPath,
		execution.InputArguments,
		execution.OutputArguments,
		execution.DataVariationIdentifier,
		c.convertToTestCaseAssertions(execution.Assertions),
		c.convertToRobotLogs(execution.RobotLogs),
	}
}

func (c uipathReportConverter) convertToTestCaseAssertions(assertions []api.TestCaseAssertion) []uipathReportTestCaseAssertion {
	testCaseAssertions := []uipathReportTestCaseAssertion{}
	for _, a := range assertions {
		testCaseAssertions = append(testCaseAssertions, uipathReportTestCaseAssertion{a.Message, a.Succeeded})
	}
	return testCaseAssertions
}

func (c uipathReportConverter) convertToRobotLogs(robotLogs []api.RobotLog) []uipathRobotLog {
	if robotLogs == nil {
		return nil
	}
	logs := []uipathRobotLog{}
	for _, l := range robotLogs {
		logs = append(logs, uipathRobotLog{
			l.Id,
			l.Level,
			l.WindowsIdentity,
			l.ProcessName,
			l.TimeStamp,
			l.Message,
			l.RobotName,
			l.HostMachineName,
			l.MachineId,
			l.MachineKey,
			l.RuntimeType,
		})
	}
	return logs
}

func (c uipathReportConverter) findTestCase(testSet *api.TestSet, id int) api.TestCase {
	for _, testCase := range testSet.TestCases {
		if testCase.Id == id {
			return testCase
		}
	}
	return *api.NewTestCase(id, "", "")
}

func newUiPathReportConverter() *uipathReportConverter {
	return &uipathReportConverter{}
}
