package utils

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/UiPath/uipathcli/log"
)

func TestProgressBarUpdateProgressShowsZeroAtTheBeginning(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))

	progressBar.UpdateProgress("downloading...", 0, 1000, 0)

	if output.String() != "\rdownloading...   0% |                    | (0.0/1.0 kB, 0 B/s)" {
		t.Errorf("Should display progress of 10 percent, but got: %v", output.String())
	}
}

func TestProgressBarUpdateProgressMultipleTimes(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))

	progressBar.UpdateProgress("downloading...", 0, 1000, 1)
	progressBar.UpdateProgress("downloading...", 200, 1000, 1024)

	if output.String() != "\rdownloading...   0% |                    | (0.0/1.0 kB, 1 B/s)"+
		"\rdownloading...  20% |████                | (0.2/1.0 kB, 1.0 kB/s)" {
		t.Errorf("Should display progress of 10 percent, but got: %v", output.String())
	}
}

func TestProgressBarShowsReadsFromProgressReader(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))
	progressReader := NewProgressReader(data(1000), func(progress Progress) {
		progressBar.UpdateProgress("download", progress.BytesRead, 1000, progress.BytesPerSecond)
	})

	_, _ = progressReader.Read(make([]byte, 100))

	lastLine := lastLine(output)
	if !strings.HasPrefix(lastLine, "download  10% |██                  | (0.1/1.0 kB,") {
		t.Errorf("Should display progress of 10 percent, but got: %v", lastLine)
	}
}

func TestProgressBarShowsMultipleReadsFromProgressReader(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))
	progressReader := NewProgressReader(data(1000), func(progress Progress) {
		progressBar.UpdateProgress("download", progress.BytesRead, 1000, progress.BytesPerSecond)
	})

	_, _ = progressReader.Read(make([]byte, 500))
	_, _ = progressReader.Read(make([]byte, 500))
	_, _ = progressReader.Read(make([]byte, 1))

	lastLine := lastLine(output)
	if !strings.HasPrefix(lastLine, "download 100% |████████████████████| (1.0/1.0 kB,") {
		t.Errorf("Should display progress of 20 percent, but got: %v", lastLine)
	}
}

func TestProgressBarUpdatePercentageShowsSimpleBar(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))

	progressBar.UpdatePercentage("building...", 0)

	lastLine := lastLine(output)
	if !strings.HasPrefix(lastLine, "building...      |                    |") {
		t.Errorf("Should display simple bar, but got: %v", lastLine)
	}
}

func TestProgressBarUpdatePercentageMovesBar(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))
	progressBar.UpdatePercentage("building...", 0)

	progressBar.UpdatePercentage("building...", 10)

	lastLine := lastLine(output)
	if !strings.HasPrefix(lastLine, "building...      |██                  |") {
		t.Errorf("Should display simple bar, but got: %v", lastLine)
	}
}

func TestProgressBarUpdatePercentageTo100(t *testing.T) {
	var output bytes.Buffer
	progressBar := NewProgressBar(log.NewDefaultLogger(&output))
	progressBar.UpdatePercentage("building...", 0)
	progressBar.UpdatePercentage("building...", 50)
	progressBar.UpdatePercentage("building...", 100)

	lastLine := lastLine(output)
	if !strings.HasPrefix(lastLine, "building...      |████████████████████|") {
		t.Errorf("Should display simple bar, but got: %v", lastLine)
	}
}

func lastLine(output bytes.Buffer) string {
	lines := strings.Split(output.String(), "\r")
	return lines[len(lines)-1]
}

func data(length int) io.Reader {
	data := make([]byte, length)
	_, err := rand.Read(data)
	if err != nil {
		panic(fmt.Errorf("Error generating random data: %w", err))
	}
	return bytes.NewReader(data)
}
