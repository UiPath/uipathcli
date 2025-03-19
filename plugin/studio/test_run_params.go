package studio

import "time"

type testRunParams struct {
	NupkgPath      string
	ProcessKey     string
	ProcessVersion string
	Timeout        time.Duration
}

func newTestRunParams(
	nupkgPath string,
	processKey string,
	processVersion string,
	timeout time.Duration) *testRunParams {
	return &testRunParams{nupkgPath, processKey, processVersion, timeout}
}
