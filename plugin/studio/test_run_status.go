package studio

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
	TotalTests     int
	CompletedTests int
	Result         *testRunResult
	Err            error
}

func newTestRunStatusUploading(executionId int) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusUploading, 0, 0, nil, nil}
}

func newTestRunStatusRunning(executionId int, totalTests int, completedTests int) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusRunning, totalTests, completedTests, nil, nil}
}

func newTestRunStatusDone(executionId int, totalTests int, result *testRunResult) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusDone, totalTests, totalTests, result, nil}
}

func newTestRunStatusError(executionId int, err error) *testRunStatus {
	return &testRunStatus{executionId, TestRunStatusError, 0, 0, nil, err}
}
