package testrun

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
	"github.com/UiPath/uipathcli/utils/process"
)

func TestRunNonExistentProjectShowsProjectJsonNotFound(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewTestRunCommand()).
		Build()

	result := test.RunCli([]string{"studio", "test", "run", "--source", "non-existent", "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if !strings.Contains(result.StdErr, "project.json not found") {
		t.Errorf("Expected stderr to show that project.json was not found, but got: %v", result.StdErr)
	}
}

func TestRunPassed(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewTestRunCommand()).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyProcess_Tests'", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", 201, `{"name":"MyProcess_Tests","processKey":"MyProcess_Tests","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "25819").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=25819&triggerType=ExternalTool", 200, "349562").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(25819)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":169537,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[{
                 "PackageIdentifier":"MyTestPackage",
                 "VersionMask":"1.2.3"
               }]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(349562)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyProcess_Tests - 1.0.195912597",
               "TestSetId":25819,
               "StartTime":"2025-03-17T12:10:09.053Z",
               "EndTime":"2025-03-17T12:10:18.183Z",
               "Status":"Passed",
               "Id":349562,
               "TestCaseExecutions":[{
                 "Id":704170,
                 "TestCaseId":169537,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed",
                 "JobId":12345,
                 "JobKey":"b6dd3f45-03c6-46c2-98ad-95a05f59905d",
                 "DataVariationIdentifier":"1",
                 "VersionNumber":"2.2.2",
                 "InputArguments":"{}",
                 "OutputArguments":"{}"
               }]
             }`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"testSetExecutions": []interface{}{
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.183Z",
				"failuresCount": 0.0,
				"id":            349562.0,
				"name":          "Automated - MyProcess_Tests - 1.0.195912597",
				"packages": []interface{}{
					map[string]interface{}{
						"name":    "MyTestPackage",
						"version": "1.2.3",
					},
				},
				"passedCount": 1.0,
				"startTime":   "2025-03-17T12:10:09.053Z",
				"status":      "Passed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "1",
						"endTime":                 "2025-03-17T12:10:18.083Z",
						"entryPointPath":          "TestCase.xaml",
						"error":                   nil,
						"id":                      704170.0,
						"inputArguments":          "{}",
						"jobId":                   12345.0,
						"name":                    "MyTestCase",
						"outputArguments":         "{}",
						"packageIdentifier":       "1.1.1",
						"startTime":               "2025-03-17T12:10:09.087Z",
						"status":                  "Passed",
						"testCaseId":              169537.0,
						"versionNumber":           "2.2.2",
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      25819.0,
			},
		},
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}

func TestRunFailed(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", 201, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "29991").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=29991&triggerType=ExternalTool", 200, "349001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(29991)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":169537,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               },
               {
                 "Id":169538,
                 "Definition":{
                   "Name":"MyTestCase2",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(349001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyLibrary - 1.0.195912597",
               "TestSetId":29991,
               "StartTime":"2025-03-17T12:10:09.087Z",
               "EndTime":"2025-03-17T12:10:25.058Z",
               "Status":"Failed",
               "Id":349001,
               "TestCaseExecutions":[{
                 "Id":704123,
                 "TestCaseId":169537,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed"
               },{
                 "Id":704124,
                 "TestCaseId":169538,
                 "EntryPointPath":"TestCase2.xaml",
                 "StartTime":"2025-03-17T12:10:21.015Z",
                 "EndTime":"2025-03-17T12:10:25.058Z",
                 "Status":"Failed",
				 "Info": "There was an error"
               }]
             }`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"testSetExecutions": []interface{}{
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:25.058Z",
				"failuresCount": 1.0,
				"id":            349001.0,
				"name":          "Automated - MyLibrary - 1.0.195912597",
				"packages":      []interface{}{},
				"passedCount":   1.0,
				"startTime":     "2025-03-17T12:10:09.087Z",
				"status":        "Failed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "",
						"endTime":                 "2025-03-17T12:10:18.083Z",
						"entryPointPath":          "TestCase.xaml",
						"error":                   nil,
						"id":                      704123.0,
						"inputArguments":          "",
						"jobId":                   0.0,
						"name":                    "MyTestCase",
						"outputArguments":         "",
						"packageIdentifier":       "1.1.1",
						"startTime":               "2025-03-17T12:10:09.087Z",
						"status":                  "Passed",
						"testCaseId":              169537.0,
						"versionNumber":           "",
					},
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "",
						"endTime":                 "2025-03-17T12:10:25.058Z",
						"entryPointPath":          "TestCase2.xaml",
						"error":                   "There was an error",
						"id":                      704124.0,
						"inputArguments":          "",
						"jobId":                   0.0,
						"name":                    "MyTestCase2",
						"outputArguments":         "",
						"packageIdentifier":       "1.1.1",
						"startTime":               "2025-03-17T12:10:21.015Z",
						"status":                  "Failed",
						"testCaseId":              169538.0,
						"versionNumber":           "",
					},
				},
				"testCasesCount": 2.0,
				"testSetId":      29991.0,
			},
		},
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}

func TestRunGeneratesJUnitReport(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[{"id":12345,"name":"MyLibrary"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(12345)", 200, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "29991").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=29991&triggerType=ExternalTool", 200, "349001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(29991)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":169537,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[{
                 "PackageIdentifier":"MyTestPackage",
                 "VersionMask":"1.2.3"
               }]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(349001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyLibrary - 1.0.195912597",
               "TestSetId":29991,
               "StartTime":"2025-03-17T12:10:09.087Z",
               "EndTime":"2025-03-17T12:10:18.083Z",
               "Status":"Passed",
               "Id":349001,
               "TestCaseExecutions":[{
                 "Id":704123,
                 "TestCaseId":169537,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed"
               }]
             }`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--results-output", "junit", "--organization", "my-org", "--tenant", "my-tenant"}, context)

	expected := `<testsuites>
  <testsuite id="349001" name="Automated - MyLibrary - 1.0.195912597" time="8.996" package="MyTestPackage-1.2.3" tests="1" failures="0" errors="0" cancelled="0">
    <system-out>Test set execution Automated - MyLibrary - 1.0.195912597 took 8996ms.&#xA;Test set execution url: ` + result.BaseUrl + `/my-org/my-tenant/orchestrator_/test/executions/349001?fid=938064&#xA;</system-out>
    <testcase name="MyTestCase" time="8.996" status="Passed" classname="1.1.1">
      <system-out>Test case MyTestCase (v) executed as job 0 and took 8996ms.&#xA;Test case logs url: ` + result.BaseUrl + `/my-org/my-tenant/orchestrator_/test/executions/349001/logs/0?fid=938064&#xA;Test case execution url: ` + result.BaseUrl + `/my-org/my-tenant/orchestrator_/test/executions/349001?search=MyTestCase&amp;fid=938064&#xA;Input arguments: &#xA;Output arguments: &#xA;</system-out>
    </testcase>
  </testsuite>
</testsuites>`
	if result.StdOut != expected {
		t.Errorf("Expected output '%v', but got: '%v'", expected, result.StdOut)
	}
}

func TestRunAttachesRobotLogs(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[{"id":12345,"name":"MyLibrary"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(12345)", 200, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "29991").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=29991&triggerType=ExternalTool", 200, "349001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(29991)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":169537,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[{
                 "PackageIdentifier":"MyTestPackage",
                 "VersionMask":"1.2.3"
               }]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/RobotLogs?$filter=JobKey%20eq%20b6dd3f45-03c6-46c2-98ad-95a05f59905d", 200,
			`{
                "value":[{
                  "Level":"Info",
                  "WindowsIdentity":"7b0afa98-bfe6-2500-1a91-0b481a664e06\\robotuser",
                  "ProcessName":"MyLibrary_Tests",
                  "TimeStamp":"2025-04-14T12:12:34.6730769Z",
                  "Message":"Verification passed.",
                  "RobotName":"myrobot-unattended",
                  "HostMachineName":"7b0afa98-bfe6-2500-1a91-0b481a664e06",
                  "MachineId":617182,
                  "MachineKey":"958977df-9af4-4b5d-b6b0-4bfbd5b2f514",
                  "RuntimeType":null,
                  "Id":0
                }]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(349001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyLibrary - 1.0.195912597",
               "TestSetId":29991,
               "StartTime":"2025-03-17T12:10:09.087Z",
               "EndTime":"2025-03-17T12:10:18.083Z",
               "Status":"Passed",
               "Id":349001,
               "TestCaseExecutions":[{
                 "Id":704123,
                 "TestCaseId":169537,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed",
                 "JobId":111111,
                 "JobKey":"b6dd3f45-03c6-46c2-98ad-95a05f59905d"
               }]
             }`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--attach-robot-logs", "true", "--organization", "my-org", "--tenant", "my-tenant"}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"testSetExecutions": []interface{}{
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.083Z",
				"failuresCount": 0.0,
				"id":            349001.0,
				"name":          "Automated - MyLibrary - 1.0.195912597",
				"packages": []interface{}{
					map[string]interface{}{
						"name":    "MyTestPackage",
						"version": "1.2.3",
					},
				},
				"passedCount": 1.0,
				"startTime":   "2025-03-17T12:10:09.087Z",
				"status":      "Passed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "",
						"endTime":                 "2025-03-17T12:10:18.083Z",
						"entryPointPath":          "TestCase.xaml",
						"error":                   nil,
						"id":                      704123.0,
						"inputArguments":          "",
						"jobId":                   111111.0,
						"name":                    "MyTestCase",
						"outputArguments":         "",
						"packageIdentifier":       "1.1.1",
						"robotLogs": []interface{}{
							map[string]interface{}{
								"hostMachineName": "7b0afa98-bfe6-2500-1a91-0b481a664e06",
								"id":              0.0,
								"level":           "Info",
								"machineId":       617182.0,
								"machineKey":      "958977df-9af4-4b5d-b6b0-4bfbd5b2f514",
								"message":         "Verification passed.",
								"processName":     "MyLibrary_Tests",
								"robotName":       "myrobot-unattended",
								"runtimeType":     "",
								"timeStamp":       "2025-04-14T12:12:34.6730769Z",
								"windowsIdentity": "7b0afa98-bfe6-2500-1a91-0b481a664e06\\robotuser",
							},
						},
						"startTime":     "2025-03-17T12:10:09.087Z",
						"status":        "Passed",
						"testCaseId":    169537.0,
						"versionNumber": "",
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      29991.0,
			},
		},
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}

func TestRunUpdatesExistingRelease(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[{"id":12345,"name":"MyLibrary"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(12345)", 200, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "29991").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=29991&triggerType=ExternalTool", 200, "349001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(29991)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":169537,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(349001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyLibrary - 1.0.195912597",
               "TestSetId":29991,
               "StartTime":"2025-03-17T12:10:09.087Z",
               "EndTime":"2025-03-17T12:10:18.083Z",
               "Status":"Passed",
               "Id":349001,
               "TestCaseExecutions":[{
                 "Id":704123,
                 "TestCaseId":169537,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed"
               }]
             }`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"testSetExecutions": []interface{}{
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.083Z",
				"failuresCount": 0.0,
				"id":            349001.0,
				"name":          "Automated - MyLibrary - 1.0.195912597",
				"packages":      []interface{}{},
				"passedCount":   1.0,
				"startTime":     "2025-03-17T12:10:09.087Z",
				"status":        "Passed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "",
						"endTime":                 "2025-03-17T12:10:18.083Z",
						"entryPointPath":          "TestCase.xaml",
						"error":                   nil,
						"id":                      704123.0,
						"inputArguments":          "",
						"jobId":                   0.0,
						"name":                    "MyTestCase",
						"outputArguments":         "",
						"packageIdentifier":       "1.1.1",
						"startTime":               "2025-03-17T12:10:09.087Z",
						"status":                  "Passed",
						"testCaseId":              169537.0,
						"versionNumber":           "",
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      29991.0,
			},
		},
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}

func TestRunTimesOutWaitingForTestExecutionToFinish(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", 201, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "29991").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=29991&triggerType=ExternalTool", 200, "349001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(29991)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":169537,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(349001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyLibrary - 1.0.195912597",
               "TestSetId":29991,
               "StartTime":"2025-03-17T12:10:09.087Z",
               "EndTime":null,
               "Status":"InProgress",
               "Id":349001,
               "TestCaseExecutions":[{
                 "Id":704123,
                 "TestCaseId":169537,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":null,
                 "Status":"InProgress"
               }]
             }`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--timeout", "3", "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Timeout waiting for test execution '349001' to finish." {
		t.Errorf("Expected test execution timeout error, but got: %v", result.Error)
	}
}

func TestRunFailsWithMissingFolder(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--timeout", "3", "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Could not find 'Shared' orchestrator folder." {
		t.Errorf("Expected missing folder error, but got: %v", result.Error)
	}
}

func TestRunFailsWithServerErrorOnStartExecution(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", 201, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 201, "29991").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=29991&triggerType=ExternalTool", 500, "{}").
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Service returned status code '500' and body '{}'" {
		t.Errorf("Expected server error, but got: %v", result.Error)
	}
}

func TestRunFailsWithServerErrorOnCreateTestSet(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", 201, `{"name":"MyLibrary","processKey":"MyLibrary","processVersion":"1.0.195912597"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion", 500, "{}").
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Service returned status code '500' and body '{}'" {
		t.Errorf("Expected server error, but got: %v", result.Error)
	}
}

func TestRunFailsWithInvalidJsonOnCreateRelease(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 200, `{"value":[]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases", 201, "invalid { json }").
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Orchestrator returned invalid response body 'invalid { json }'" {
		t.Errorf("Expected invalid response error, but got: %v", result.Error)
	}
}

func TestRunFailsWithBadRequestOnGetReleases(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyLibrary'", 400, `{"value":[]}`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != `Orchestrator returned status code '400' and body '{"value":[]}'` {
		t.Errorf("Expected client error, but got: %v", result.Error)
	}
}

func TestRunFailsWithBadRequestOnUploadPackage(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 400, `Bad Request`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Orchestrator returned status code '400' and body 'Bad Request'" {
		t.Errorf("Expected server error, but got: %v", result.Error)
	}
}

func TestRunFailsWithUnauthorizedOnGetFolders(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "Done", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 401, `{}`).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	if result.Error == nil || result.Error.Error() != "Orchestrator returned status code '401' and body '{}'" {
		t.Errorf("Expected server error, but got: %v", result.Error)
	}
}

func TestRunWithLibraryAuthentication(t *testing.T) {
	config := `
profiles:
  - name: default
    organization: my-org
    tenant: my-tenant
    auth:
      clientId: success-client-id
      clientSecret: success-client-secret
`
	commandArgs := []string{}
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		commandArgs = args
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithConfig(config).
		WithResponse(200, "").
		WithIdentityResponse(200, `{"access_token": "my-jwt-access-token", "expires_in": 3600, "token_type": "Bearer", "scope": "OR.Ping"}`).
		WithCommandPlugin(TestRunCommand{exec}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source}, context)

	identityUrl := test.GetArgumentValue(commandArgs, "--libraryIdentityUrl")
	if identityUrl != result.BaseUrl+"/identity_" {
		t.Errorf("Expected identity url as argument, but got: %v", commandArgs)
	}

	orchestratorUrl := test.GetArgumentValue(commandArgs, "--libraryOrchestratorUrl")
	if orchestratorUrl != result.BaseUrl {
		t.Errorf("Expected orchestrator url as argument, but got: %v", commandArgs)
	}

	authToken := test.GetArgumentValue(commandArgs, "--libraryOrchestratorAuthToken")
	if authToken != "my-jwt-access-token" {
		t.Errorf("Expected jwt bearer token as argument, but got: %v", commandArgs)
	}

	organization := test.GetArgumentValue(commandArgs, "--libraryOrchestratorAccountName")
	if organization != "my-org" {
		t.Errorf("Expected organization as argument, but got: %v", commandArgs)
	}

	tenant := test.GetArgumentValue(commandArgs, "--libraryOrchestratorTenant")
	if tenant != "my-tenant" {
		t.Errorf("Expected tenant as argument, but got: %v", commandArgs)
	}
}

func TestRunParallelPassed(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewTestRunCommand()).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Folders", 200, `{"value":[{"Id":938064,"FullyQualifiedName":"Shared"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyFirstProcess_Tests'", 200, `{"value":[{"id":10000,"name":"MyFirstProcess_Tests"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MySecondProcess_Tests'", 200, `{"value":[{"id":20000,"name":"MySecondProcess_Tests"}]}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(10000)", 200, `{"name":"MyFirstProcess_Tests","processKey":"MyFirstProcess_Tests","processVersion":"1.0.0"}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases(20000)", 200, `{"name":"MySecondProcess_Tests","processKey":"MySecondProcess_Tests","processVersion":"2.0.0"}`).
		WithResponseHandler(func(request test.RequestData) test.ResponseData {
			if request.URL.Path == "/my-org/my-tenant/orchestrator_/api/TestAutomation/CreateTestSetForReleaseVersion" {
				body := map[string]interface{}{}
				err := json.Unmarshal(request.Body, &body)
				if err != nil {
					return test.ResponseData{Status: 500, Body: err.Error()}
				}
				if body["releaseId"] == 10000.0 {
					return test.ResponseData{Status: 201, Body: "100002"}
				}
				return test.ResponseData{Status: 201, Body: "200002"}
			}
			return test.ResponseData{Status: 500, Body: fmt.Sprintf("Unhandled HTTP request %s", request.URL.Path)}
		}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=100002&triggerType=ExternalTool", 200, "100001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/api/TestAutomation/StartTestSetExecution?testSetId=200002&triggerType=ExternalTool", 200, "200001").
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(100002)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":100004,
                 "Definition":{
                   "Name":"MyTestCase",
                   "PackageIdentifier":"1.1.1"
                 }
               }],
               "Packages":[]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSets(200002)?$expand=TestCases($expand=Definition;$select=Id,Definition,DefinitionId,ReleaseId,VersionNumber),Packages&$select=TestCases,Name", 200,
			`{
               "TestCases":[{
                 "Id":200004,
                 "Definition":{
                   "Name":"MySecondTestCase",
                   "PackageIdentifier":"2.2.2"
                 }
               }],
               "Packages":[]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(100001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MyFirstProcess_Tests - 1.0.0",
               "TestSetId":100002,
               "StartTime":"2025-03-17T12:10:09.053Z",
               "EndTime":"2025-03-17T12:10:18.183Z",
               "Status":"Passed",
               "Id":100001,
               "TestCaseExecutions":[{
                 "Id":100003,
                 "TestCaseId":100004,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Passed"
               }]
             }`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/TestSetExecutions(200001)?$expand=TestCaseExecutions($expand=TestCaseAssertions)", 200,
			`{
               "Name":"Automated - MySecondProcess_Tests - 2.0.0",
               "TestSetId":200002,
               "StartTime":"2025-03-17T12:10:09.053Z",
               "EndTime":"2025-03-17T12:10:18.183Z",
               "Status":"Failed",
               "Id":200001,
               "TestCaseExecutions":[{
                 "Id":200003,
                 "TestCaseId":200004,
                 "EntryPointPath":"TestCase.xaml",
                 "StartTime":"2025-03-17T12:10:09.087Z",
                 "EndTime":"2025-03-17T12:10:18.083Z",
                 "Status":"Failed"
               }]
             }`).
		Build()

	source1 := test.NewCrossPlatformProject(t).
		WithProjectName("MyFirstProcess").
		Build()
	source2 := test.NewCrossPlatformProject(t).
		WithProjectName("MySecondProcess").
		Build()
	result := test.RunCli([]string{"studio", "test", "run", "--source", source1 + "," + source2, "--organization", "my-org", "--tenant", "my-tenant"}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"testSetExecutions": []interface{}{
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.183Z",
				"failuresCount": 0.0,
				"id":            100001.0,
				"name":          "Automated - MyFirstProcess_Tests - 1.0.0",
				"packages":      []interface{}{},
				"passedCount":   1.0,
				"startTime":     "2025-03-17T12:10:09.053Z",
				"status":        "Passed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "",
						"endTime":                 "2025-03-17T12:10:18.083Z",
						"entryPointPath":          "TestCase.xaml",
						"error":                   nil,
						"id":                      100003.0,
						"inputArguments":          "",
						"jobId":                   0.0,
						"name":                    "MyTestCase",
						"outputArguments":         "",
						"packageIdentifier":       "1.1.1",
						"startTime":               "2025-03-17T12:10:09.087Z",
						"status":                  "Passed",
						"testCaseId":              100004.0,
						"versionNumber":           "",
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      100002.0,
			},
			map[string]interface{}{
				"canceledCount": 0.0,
				"endTime":       "2025-03-17T12:10:18.183Z",
				"failuresCount": 1.0,
				"id":            200001.0,
				"name":          "Automated - MySecondProcess_Tests - 2.0.0",
				"packages":      []interface{}{},
				"passedCount":   0.0,
				"startTime":     "2025-03-17T12:10:09.053Z",
				"status":        "Failed",
				"testCaseExecutions": []interface{}{
					map[string]interface{}{
						"assertions":              []interface{}{},
						"dataVariationIdentifier": "",
						"endTime":                 "2025-03-17T12:10:18.083Z",
						"entryPointPath":          "TestCase.xaml",
						"error":                   nil,
						"id":                      200003.0,
						"inputArguments":          "",
						"jobId":                   0.0,
						"name":                    "MySecondTestCase",
						"outputArguments":         "",
						"packageIdentifier":       "2.2.2",
						"startTime":               "2025-03-17T12:10:09.087Z",
						"status":                  "Failed",
						"testCaseId":              200004.0,
						"versionNumber":           "",
					},
				},
				"testCasesCount": 1.0,
				"testSetId":      200002.0,
			},
		},
	}

	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}

func TestRunUsesProvidedFolderId(t *testing.T) {
	exec := process.NewExecCustomProcess(0, "", "", func(name string, args []string) {
		outputDirectory := test.GetArgumentValue(args, "--output")
		writeNupkgArchive(t, filepath.Join(outputDirectory, "MyLibrary_Tests.nupkg"))
	})
	header := map[string]string{}
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(TestRunCommand{exec}).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Processes/UiPath.Server.Configuration.OData.UploadPackage", 200, `{}`).
		WithUrlResponse("/my-org/my-tenant/orchestrator_/odata/Releases?$filter=ProcessKey%20eq%20'MyProcess_Tests'", 200, `{"value":[]}`).
		WithResponseHandler(func(request test.RequestData) test.ResponseData {
			header = request.Header
			return test.ResponseData{Status: 200, Body: ""}
		}).
		Build()

	source := test.NewCrossPlatformProject(t).
		Build()
	test.RunCli([]string{"studio", "test", "run", "--source", source, "--organization", "my-org", "--tenant", "my-tenant", "--folder-id", "12345"}, context)

	folderId := header["x-uipath-organizationunitid"]
	if folderId != "12345" {
		t.Errorf("Expected x-uipath-organizationunitid header from argument, but got: '%s'", folderId)
	}
}

func writeNupkgArchive(t *testing.T, fileName string) {
	err := studio.NewNupkgWriter(fileName).
		WithNuspec(*studio.NewNuspec("MyLibrary", "My Library", "1.0.0")).
		Write()
	if err != nil {
		t.Fatal(err)
	}
}
