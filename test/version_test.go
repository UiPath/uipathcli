package test

import (
	"strings"
	"testing"
)

func TestVersionOutput(t *testing.T) {
	context := NewContextBuilder().
		Build()

	result := RunCli([]string{"--version"}, context)

	if !strings.HasPrefix(result.StdOut, "uipathcli main") {
		t.Errorf("Did not return version information, got: %v", result.StdOut)
	}
}
