package testrun

import "github.com/UiPath/uipathcli/utils/api"

const (
	TestRunStatusPackaging = "packaging"
	TestRunStatusUploading = "uploading"
	TestRunStatusRunning   = "running"
	TestRunStatusDone      = "done"
	TestRunStatusError     = "error"
)

type testRunStatus struct {
	ExecutionId    int
	State          string
	FolderId       int
	TotalTests     int
	CompletedTests int
	TestSet        *api.TestSet
	Execution      *api.TestExecution
	Err            error
}

func newTestRunStatusUploading(executionId int) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusUploading, -1, 0, 0, nil, nil, nil}
}

func newTestRunStatusRunning(executionId int, folderId int, totalTests int, completedTests int) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusRunning, folderId, totalTests, completedTests, nil, nil, nil}
}

func newTestRunStatusDone(executionId int, folderId int, totalTests int, testSet *api.TestSet, result *api.TestExecution) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusDone, folderId, totalTests, totalTests, testSet, result, nil}
}

func newTestRunStatusError(executionId int, err error) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusError, -1, 0, 0, nil, nil, err}
}
