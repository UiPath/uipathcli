package test

import (
	"os"
	"testing"
)

func TestEnableAutocompleteInvalidShellShowsError(t *testing.T) {
	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	result := RunCli([]string{"autocomplete", "enable", "--shell", "invalid"}, context)

	expectedError := "Invalid shell, supported values: powershell, bash\n"
	if result.StdErr != expectedError {
		t.Errorf("Should show invalid shell error, got: %v", result.StdErr)
	}
}

func TestEnableAutocompletePowershellShowsSuccessOutput(t *testing.T) {
	profilePath := CreateFile(t)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	result := RunCli([]string{"autocomplete", "enable", "--shell", "powershell", "--file", profilePath}, context)

	expectedOutput := `Shell: powershell
Profile: ` + profilePath + `

Successfully enabled command completion! Restart your shell for the changes to take effect.
`
	if result.StdOut != expectedOutput {
		t.Errorf("Should show enabled command completion message, got: %v", result.StdOut)
	}
}

func TestEnableAutocompletePowershellCreatesProfileFile(t *testing.T) {
	profilePath := CreateFile(t)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	RunCli([]string{"autocomplete", "enable", "--shell", "powershell", "--file", profilePath}, context)

	content, _ := os.ReadFile(profilePath)

	expectedFileContent := `
$uipath_auto_complete = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $padLength = $cursorPosition - $commandAst.Extent.StartOffset
    $textToComplete = $commandAst.ToString().PadRight($padLength, ' ').Substring(0, $padLength)
    $command, $params = $commandAst.ToString() -split " ", 2
    & $command autocomplete complete --command "$textToComplete" | foreach-object {
        [system.management.automation.completionresult]::new($_, $_, 'parametervalue', $_)
    }
}
Register-ArgumentCompleter -Native -CommandName uipath -ScriptBlock $uipath_auto_complete
`
	if string(content) != expectedFileContent {
		t.Errorf("Should create profile file with correct content, got: %v", string(content))
	}
}

func TestEnableAutocompletePowershellUpdatesExistingProfileFile(t *testing.T) {
	profilePath := CreateFileWithContent(t, "existing content\nshould not change")

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	RunCli([]string{"autocomplete", "enable", "--shell", "powershell", "--file", profilePath}, context)

	content, _ := os.ReadFile(profilePath)

	expectedFileContent := `existing content
should not change
$uipath_auto_complete = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $padLength = $cursorPosition - $commandAst.Extent.StartOffset
    $textToComplete = $commandAst.ToString().PadRight($padLength, ' ').Substring(0, $padLength)
    $command, $params = $commandAst.ToString() -split " ", 2
    & $command autocomplete complete --command "$textToComplete" | foreach-object {
        [system.management.automation.completionresult]::new($_, $_, 'parametervalue', $_)
    }
}
Register-ArgumentCompleter -Native -CommandName uipath -ScriptBlock $uipath_auto_complete
`
	if string(content) != expectedFileContent {
		t.Errorf("Should update profile file with correct content, got: %v", string(content))
	}
}

func TestEnableAutocompletePowershellNoChangesIfEnabledAlready(t *testing.T) {
	initialContent := `
$uipath_auto_complete = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $command, $params = $commandAst.ToString() -split " ", 2
    & $command autocomplete complete --command "$commandAst" | foreach-object {
        [system.management.automation.completionresult]::new($_, $_, 'parametervalue', $_)
    }
}
Register-ArgumentCompleter -Native -CommandName uipath -ScriptBlock $uipath_auto_complete
`
	profilePath := CreateFileWithContent(t, initialContent)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	RunCli([]string{"autocomplete", "enable", "--shell", "powershell", "--file", profilePath}, context)

	content, _ := os.ReadFile(profilePath)
	if string(content) != initialContent {
		t.Errorf("Should not update profile file when auto-complete is already enabled, got: %v", string(content))
	}
}

func TestEnableAutocompletePowershellShowsAlreadyEnabledOutput(t *testing.T) {
	initialContent := `
$uipath_auto_complete = {
    param($wordToComplete, $commandAst, $cursorPosition)
    $command, $params = $commandAst.ToString() -split " ", 2
    & $command autocomplete complete --command "$commandAst" | foreach-object {
        [system.management.automation.completionresult]::new($_, $_, 'parametervalue', $_)
    }
}
Register-ArgumentCompleter -Native -CommandName uipath -ScriptBlock $uipath_auto_complete
`
	profilePath := CreateFileWithContent(t, initialContent)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	result := RunCli([]string{"autocomplete", "enable", "--shell", "powershell", "--file", profilePath}, context)

	exepectedOutput := `Shell: powershell
Profile: ` + profilePath + `

Command completion is already enabled.
`
	if result.StdOut != exepectedOutput {
		t.Errorf("Should show output that auto-complete is already enabled, got: %v", result.StdOut)
	}
}

func TestEnableAutocompleteBashShowsSuccessOutput(t *testing.T) {
	profilePath := CreateFile(t)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	result := RunCli([]string{"autocomplete", "enable", "--shell", "bash", "--file", profilePath}, context)

	expectedOutput := `Shell: bash
Profile: ` + profilePath + `

Successfully enabled command completion! Restart your shell for the changes to take effect.
`
	if result.StdOut != expectedOutput {
		t.Errorf("Should show enabled command completion message, got: %v", result.StdOut)
	}
}

func TestEnableAutocompleteBashCreatesProfileFile(t *testing.T) {
	profilePath := CreateFile(t)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	RunCli([]string{"autocomplete", "enable", "--shell", "bash", "--file", profilePath}, context)

	content, _ := os.ReadFile(profilePath)

	expectedFileContent := `
function _uipath_auto_complete()
{
  local executable="${COMP_WORDS[0]}"
  local cur="${COMP_WORDS[COMP_CWORD]}" IFS=$'\n'
  local candidates
  read -d '' -ra candidates < <($executable autocomplete complete --command "${COMP_LINE}" 2>/dev/null)
  read -d '' -ra COMPREPLY < <(compgen -W "${candidates[*]:-}" -- "$cur")
}
complete -f -F _uipath_auto_complete uipath
`
	if string(content) != expectedFileContent {
		t.Errorf("Should create profile file with correct content, got: %v", string(content))
	}
}

func TestEnableAutocompleteBashUpdatesExistingProfileFile(t *testing.T) {
	profilePath := CreateFileWithContent(t, "\nexisting content\nshould not change\n")

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	RunCli([]string{"autocomplete", "enable", "--shell", "bash", "--file", profilePath}, context)

	content, _ := os.ReadFile(profilePath)

	expectedFileContent := `
existing content
should not change

function _uipath_auto_complete()
{
  local executable="${COMP_WORDS[0]}"
  local cur="${COMP_WORDS[COMP_CWORD]}" IFS=$'\n'
  local candidates
  read -d '' -ra candidates < <($executable autocomplete complete --command "${COMP_LINE}" 2>/dev/null)
  read -d '' -ra COMPREPLY < <(compgen -W "${candidates[*]:-}" -- "$cur")
}
complete -f -F _uipath_auto_complete uipath
`
	if string(content) != expectedFileContent {
		t.Errorf("Should update profile file with correct content, got: %v", string(content))
	}
}

func TestEnableAutocompleteBashNoChangesIfEnabledAlready(t *testing.T) {
	initialContent := `
function _uipath_auto_complete()
{
  local executable="${COMP_WORDS[0]}"
  local cur="${COMP_WORDS[COMP_CWORD]}" IFS=$'\n'
  local candidates
  read -d '' -ra candidates < <($executable autocomplete complete --command "${COMP_LINE}" 2>/dev/null)
  read -d '' -ra COMPREPLY < <(compgen -W "${candidates[*]:-}" -- "$cur")
}
complete -f -F _uipath_auto_complete uipath
`
	profilePath := CreateFileWithContent(t, initialContent)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	RunCli([]string{"autocomplete", "enable", "--shell", "bash", "--file", profilePath}, context)

	content, _ := os.ReadFile(profilePath)
	if string(content) != initialContent {
		t.Errorf("Should not update profile file when auto-complete is already enabled, got: %v", string(content))
	}
}

func TestEnableAutocompleteBashShowsAlreadyEnabledOutput(t *testing.T) {
	initialContent := `
function _uipath_auto_complete()
{
  local executable="${COMP_WORDS[0]}"
  local cur="${COMP_WORDS[COMP_CWORD]}" IFS=$'\n'
  local candidates
  read -d '' -ra candidates < <($executable autocomplete complete --command "${COMP_LINE}" 2>/dev/null)
  read -d '' -ra COMPREPLY < <(compgen -W "${candidates[*]:-}" -- "$cur")
}
complete -f -F _uipath_auto_complete uipath
`
	profilePath := CreateFileWithContent(t, initialContent)

	context := NewContextBuilder().
		WithDefinition("myservice", "").
		Build()

	result := RunCli([]string{"autocomplete", "enable", "--shell", "bash", "--file", profilePath}, context)

	exepectedOutput := `Shell: bash
Profile: ` + profilePath + `

Command completion is already enabled.
`
	if result.StdOut != exepectedOutput {
		t.Errorf("Should show output that auto-complete is already enabled, got: %v", result.StdOut)
	}
}
