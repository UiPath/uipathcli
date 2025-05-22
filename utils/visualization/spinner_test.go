package visualization

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/UiPath/uipathcli/log"
)

func TestSpinnerOutputsWaitingCircle(t *testing.T) {
	var writer bytes.Buffer
	spinner := NewSpinner(log.NewDefaultLogger(&writer), "progress: ")
	time.Sleep(1 * time.Second)
	spinner.Close()

	output := writer.String()
	if !strings.HasPrefix(output, "\rprogress: |\rprogress: /\rprogress: -\rprogress: \\\rprogress: |") {
		t.Errorf("Spinner should display waiting circle, but got: %v", output)
	}
}

func TestSpinnerStopsOnClose(t *testing.T) {
	var writer bytes.Buffer
	spinner := NewSpinner(log.NewDefaultLogger(&writer), "progress: ")
	time.Sleep(200 * time.Millisecond)
	spinner.Close()
	time.Sleep(200 * time.Millisecond)

	output1 := writer.String()
	time.Sleep(200 * time.Millisecond)
	output2 := writer.String()
	if output1 != output2 {
		t.Errorf("Spinner should stop outputting, but got additional output. before: %v after: %v", output1, output2)
	}
}

func TestSpinnerHidesOnClose(t *testing.T) {
	var writer bytes.Buffer
	spinner := NewSpinner(log.NewDefaultLogger(&writer), "progress: ")
	time.Sleep(200 * time.Millisecond)
	spinner.Close()
	time.Sleep(200 * time.Millisecond)

	output := writer.String()
	if !strings.HasSuffix(output, "\r           \r") {
		t.Errorf("Spinner should hide on close, but got: %v", output)
	}
}
