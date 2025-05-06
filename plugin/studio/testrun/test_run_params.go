package testrun

import (
	"time"

	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/plugin/studio"
)

type testRunParams struct {
	ExecutionId     int
	Uipcli          *studio.Uipcli
	Logger          log.Logger
	Source          string
	Destination     string
	Timeout         time.Duration
	AttachRobotLogs bool
	Folder          string
}

func newTestRunParams(
	executionId int,
	uipcli *studio.Uipcli,
	logger log.Logger,
	source string,
	destination string,
	timeout time.Duration,
	attachRobotLogs bool,
	folder string,
) *testRunParams {
	return &testRunParams{
		executionId,
		uipcli,
		logger,
		source,
		destination,
		timeout,
		attachRobotLogs,
		folder,
	}
}
