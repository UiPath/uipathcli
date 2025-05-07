package testrun

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/UiPath/uipathcli/utils/api"
)

type jUnitReportConverter struct {
	client *api.OrchestratorClient
}

func (c jUnitReportConverter) Convert(status []testRunStatus) junitReport {
	testSuites := []junitReportTestSuite{}
	for _, s := range status {
		testSuite := c.convertTestSetExecution(s)
		testSuites = append(testSuites, testSuite)
	}
	return junitReport{
		TestSuites: testSuites,
	}
}

func (c jUnitReportConverter) convertTestSetExecution(status testRunStatus) junitReportTestSuite {
	junitTestCases := []junitReportTestCase{}
	for _, execution := range status.Execution.TestCaseExecutions {
		testCase := c.findTestCase(status.TestSet, execution.TestCaseId)
		junitTestCase := c.convertTestCaseExecution(status.FolderId, status.Execution.Id, testCase, execution)
		junitTestCases = append(junitTestCases, junitTestCase)
	}

	testSetExecutionUrl := c.client.GetTestSetExecutionUrl(status.FolderId, status.Execution.Id)
	durationMs := status.Execution.EndTime.Sub(status.Execution.StartTime).Milliseconds()
	systemOut := fmt.Sprintf(`Test set execution %s took %dms.
Test set execution url: %s
`, status.Execution.Name, durationMs, testSetExecutionUrl)

	packages := ""
	for _, p := range status.TestSet.Packages {
		if packages != "" {
			packages += "; "
		}
		packages += p.PackageIdentifier + "-" + p.VersionMask
	}

	return junitReportTestSuite{
		Id:        strconv.Itoa(status.Execution.Id),
		Name:      status.Execution.Name,
		Time:      status.Execution.EndTime.Sub(status.Execution.StartTime).Seconds(),
		Package:   packages,
		Tests:     status.Execution.TestCasesCount,
		Failures:  status.Execution.FailuresCount,
		Errors:    0,
		Cancelled: status.Execution.CanceledCount,
		SystemOut: systemOut,
		TestCases: junitTestCases,
	}
}

func (c jUnitReportConverter) convertTestCaseExecution(folderId int, testSetExecutionId int, testCase api.TestCase, execution api.TestCaseExecution) junitReportTestCase {
	name := testCase.Name
	if execution.DataVariationIdentifier != "" {
		name = testCase.Name + "_" + execution.DataVariationIdentifier
	}
	testCaseExecutionUrl := c.client.GetTestCaseExecutionUrl(folderId, testSetExecutionId, testCase.Name)
	testCaseExecutionLogsUrl := c.client.GetTestCaseExecutionLogsUrl(folderId, testSetExecutionId, execution.JobId)
	durationMs := execution.EndTime.Sub(execution.StartTime).Milliseconds()
	systemOut := fmt.Sprintf(`Test case %s (v%s) executed as job %d and took %dms.
Test case logs url: %s
Test case execution url: %s
Input arguments: %s
Output arguments: %s
`, name, execution.VersionNumber, execution.JobId, durationMs, testCaseExecutionLogsUrl, testCaseExecutionUrl, execution.InputArguments, execution.OutputArguments)

	if len(execution.Assertions) > 0 {
		systemOut += "Assertions:\n"
		for _, assertion := range execution.Assertions {
			systemOut += "  " + assertion.Message + "\n"
		}
	}

	if len(execution.RobotLogs) > 0 {
		json, _ := json.MarshalIndent(execution.RobotLogs, "  ", "  ")
		systemOut += "Robot logs:\n" + string(json) + "\n"
	}

	return junitReportTestCase{
		Name:      name,
		Status:    execution.Status,
		Time:      execution.EndTime.Sub(execution.StartTime).Seconds(),
		Classname: testCase.PackageIdentifier,
		SystemOut: systemOut,
	}
}

func (c jUnitReportConverter) findTestCase(testSet *api.TestSet, id int) api.TestCase {
	for _, testCase := range testSet.TestCases {
		if testCase.Id == id {
			return testCase
		}
	}
	return *api.NewTestCase(id, "", "")
}

func newJUnitReportConverter(client *api.OrchestratorClient) *jUnitReportConverter {
	return &jUnitReportConverter{client}
}
