//go:build windows

package restore

import (
	"reflect"
	"testing"

	"github.com/UiPath/uipathcli/plugin/studio"
	"github.com/UiPath/uipathcli/test"
)

func TestRestoreWindowsSuccessfully(t *testing.T) {
	context := test.NewContextBuilder().
		WithDefinition("studio", studio.StudioDefinition).
		WithCommandPlugin(NewPackageRestoreCommand()).
		Build()

	source := test.NewWindowsProject(t).
		Build()
	destination := test.CreateDirectory(t)
	result := test.RunCli([]string{"studio", "package", "restore", "--source", source, "--destination", destination}, context)

	stdout := test.ParseOutput(t, result.StdOut)
	expected := map[string]interface{}{
		"status":      "Succeeded",
		"error":       nil,
		"name":        "MyWindowsProcess",
		"description": "Blank Process",
		"projectId":   "94c4c9c1-68c3-45d4-be14-d4427f17eddd",
		"output":      destination,
	}
	if !reflect.DeepEqual(expected, stdout) {
		t.Errorf("Expected output '%v', but got: '%v'", expected, stdout)
	}
}
