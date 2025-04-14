package studio

import (
	"time"

	"github.com/UiPath/uipathcli/log"
)

type testRunParams struct {
	ExecutionId     int
	Uipcli          *uipcli
	Logger          log.Logger
	Source          string
	Destination     string
	Timeout         time.Duration
	AttachRobotLogs bool
}

func newTestRunParams(
	executionId int,
	uipcli *uipcli,
	logger log.Logger,
	source string,
	destination string,
	timeout time.Duration,
	attachRobotLogs bool) *testRunParams {
	return &testRunParams{executionId, uipcli, logger, source, destination, timeout, attachRobotLogs}
}
