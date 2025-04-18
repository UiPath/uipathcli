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
	FolderId        int
}

func newTestRunParams(
	executionId int,
	uipcli *studio.Uipcli,
	logger log.Logger,
	source string,
	destination string,
	timeout time.Duration,
	attachRobotLogs bool,
	folderId int,
) *testRunParams {
	return &testRunParams{
		executionId,
		uipcli,
		logger,
		source,
		destination,
		timeout,
		attachRobotLogs,
		folderId,
	}
}
