package studio

type testRunReport struct {
	TestCaseExecutions []testRunResult `json:"testSetExecutions"`
}

func newTestRunReport(testCaseExecutions []testRunResult) *testRunReport {
	return &testRunReport{testCaseExecutions}
}
