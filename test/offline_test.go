package test

import (
	"fmt"
	"runtime"
	"testing"
)

func TestOfflineDownloadsModules(t *testing.T) {
	context := NewContextBuilder().
		Build()
	result := RunCli([]string{"config", "offline"}, context)

	dotnetRuntime := map[string]string{
		"linux-amd64":   "dotnet-runtime-linux-x64.tar.gz",
		"windows-amd64": "dotnet-runtime-win-x64.zip",
		"darwin-amd64":  "dotnet-runtime-osx-x64.tar.gz",
		"linux-arm64":   "dotnet-runtime-linux-arm64.tar.gz",
		"windows-arm64": "dotnet-runtime-win-arm64.zip",
		"darwin-arm64":  "dotnet-runtime-osx-arm64.tar.gz",
	}
	expectedOutput := fmt.Sprintf(`
[ succeeded ] uipcli: successfully downloaded from https://github.com/UiPath/uipathcli/releases/download/plugins-v2.0.0/UiPath.CLI.24.12.9208.28468.nupkg
[ succeeded ] uipcli-win: successfully downloaded from https://github.com/UiPath/uipathcli/releases/download/plugins-v2.0.0/UiPath.CLI.Windows.24.12.9208.28468.nupkg
[ succeeded ] dotnet8-%s-%s: successfully downloaded from https://aka.ms/dotnet/8.0/%s

Successfully downloaded all modules for offline mode!`, runtime.GOOS, runtime.GOARCH, dotnetRuntime[runtime.GOOS+"-"+runtime.GOARCH])
	if result.StdOut != expectedOutput {
		t.Errorf("Expected output that modules were downloaded successfully '%s', but got: '%s'", expectedOutput, result.StdOut)
	}
}
