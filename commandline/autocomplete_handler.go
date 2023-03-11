package commandline

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v2"
)

const directoryPermissions = 0755
const filePermissions = 0644

const Powershell = "powershell"
const Bash = "bash"

const completeHandlerEnabledCheck = "uipath_auto_complete"

const powershellCompleteHandler = `
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

const bashCompleteHandler = `
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

// autoCompleteHandler parses the autocomplete command and provides suggestions for the available commands.
// It tries to perform a prefix- as well as contains-match based on the current context.
// Example:
// uipath autocomplete complete --command "uipath o"
// returns:
// oms
// orchestrator
// documentunderstanding
type autoCompleteHandler struct {
}

func (a autoCompleteHandler) EnableCompleter(shell string, filePath string) (string, error) {
	if shell != Powershell && shell != Bash {
		return "", fmt.Errorf("Invalid shell, supported values: %s, %s", Powershell, Bash)
	}

	profileFilePath, err := a.profileFilePath(shell, filePath)
	if err != nil {
		return "", err
	}
	completeHandler := a.completeHandler(shell)
	return a.enableCompleter(shell, profileFilePath, completeHandlerEnabledCheck, completeHandler)
}

func (a autoCompleteHandler) profileFilePath(shell string, filePath string) (string, error) {
	if filePath != "" {
		return filePath, nil
	}
	if shell == Powershell {
		return PowershellProfilePath()
	}
	return BashrcPath()
}

func (a autoCompleteHandler) completeHandler(shell string) string {
	if shell == Powershell {
		return powershellCompleteHandler
	}
	return bashCompleteHandler
}

func (a autoCompleteHandler) enableCompleter(shell string, filePath string, enabledCheck string, completerHandler string) (string, error) {
	err := a.ensureDirectoryExists(filePath)
	if err != nil {
		return "", err
	}
	enabled, err := a.completerEnabled(filePath, enabledCheck)
	if err != nil {
		return "", err
	}
	if enabled {
		output := fmt.Sprintf("Shell: %s\nProfile: %s\n\nCommand completion is already enabled.", shell, filePath)
		return output, nil
	}
	err = a.writeCompleterHandler(filePath, completerHandler)
	if err != nil {
		return "", err
	}
	output := fmt.Sprintf("Shell: %s\nProfile: %s\n\nSuccessfully enabled command completion! Restart your shell for the changes to take effect.", shell, filePath)
	return output, nil
}

func (a autoCompleteHandler) ensureDirectoryExists(filePath string) error {
	err := os.MkdirAll(filepath.Dir(filePath), directoryPermissions)
	if err != nil {
		return fmt.Errorf("Error creating profile folder: %v", err)
	}
	return nil
}

func (a autoCompleteHandler) completerEnabled(filePath string, enabledCheck string) (bool, error) {
	content, err := os.ReadFile(filePath)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("Error reading profile file: %v", err)
	}
	return strings.Contains(string(content), enabledCheck), nil
}

func (a autoCompleteHandler) writeCompleterHandler(filePath string, completerHandler string) error {
	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return fmt.Errorf("Error opening profile file: %v", err)
	}
	defer file.Close()
	if _, err := file.WriteString(completerHandler); err != nil {
		return fmt.Errorf("Error writing profile file: %v", err)
	}
	return nil
}

func (a autoCompleteHandler) Find(commandText string, commands []*cli.Command, exclude []string) []string {
	words := strings.Split(commandText, " ")
	if len(words) < 2 {
		return []string{}
	}

	command := &cli.Command{
		Name:        "uipath",
		Subcommands: commands,
	}

	for _, word := range words[1 : len(words)-1] {
		if strings.HasPrefix(word, "-") {
			break
		}
		command = a.findCommand(word, command.Subcommands)
		if command == nil {
			return []string{}
		}
	}

	lastWord := words[len(words)-1]
	if strings.HasPrefix(lastWord, "-") {
		return a.searchFlags(strings.TrimLeft(lastWord, "-"), command, append(exclude, words...))
	}
	return a.searchCommands(lastWord, command.Subcommands, exclude)
}

func (a autoCompleteHandler) findCommand(name string, commands []*cli.Command) *cli.Command {
	for _, command := range commands {
		if command.Name == name {
			return command
		}
	}
	return nil
}

func (a autoCompleteHandler) searchCommands(word string, commands []*cli.Command, exclude []string) []string {
	result := []string{}
	for _, command := range commands {
		if strings.HasPrefix(command.Name, word) {
			result = append(result, command.Name)
		}
	}
	for _, command := range commands {
		if strings.Contains(command.Name, word) {
			result = append(result, command.Name)
		}
	}
	return a.removeDuplicates(a.removeExcluded(result, exclude))
}

func (a autoCompleteHandler) searchFlags(word string, command *cli.Command, exclude []string) []string {
	result := []string{}
	for _, flag := range command.Flags {
		flagNames := flag.Names()
		for _, flagName := range flagNames {
			if strings.HasPrefix(flagName, word) {
				result = append(result, "--"+flagName)
			}
		}
	}
	for _, flag := range command.Flags {
		flagNames := flag.Names()
		for _, flagName := range flagNames {
			if strings.Contains(flagName, word) {
				result = append(result, "--"+flagName)
			}
		}
	}
	return a.removeDuplicates(a.removeExcluded(result, exclude))
}

func (a autoCompleteHandler) removeDuplicates(values []string) []string {
	keys := make(map[string]bool)
	result := []string{}

	for _, entry := range values {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			result = append(result, entry)
		}
	}
	return result
}

func (a autoCompleteHandler) removeExcluded(values []string, exclude []string) []string {
	result := []string{}
	for _, entry := range values {
		if !a.contains(exclude, entry) {
			result = append(result, entry)
		}
	}
	return result
}

func (a autoCompleteHandler) contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func newAutoCompleteHandler() *autoCompleteHandler {
	return &autoCompleteHandler{}
}
