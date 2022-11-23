package commandline

import (
	"strings"

	"github.com/urfave/cli/v2"
)

type AutoCompleteHandler struct {
}

func (a AutoCompleteHandler) Find(commandText string, commands []*cli.Command) []string {
	words := strings.Split(commandText, " ")
	if len(words) < 2 {
		return []string{}
	}

	command := &cli.Command{
		Name:        "uipathcli",
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
		return a.searchFlags(strings.TrimLeft(lastWord, "-"), command)
	}
	return a.searchCommands(lastWord, command.Subcommands)
}

func (a AutoCompleteHandler) findCommand(name string, commands []*cli.Command) *cli.Command {
	for _, command := range commands {
		if command.Name == name {
			return command
		}
	}
	return nil
}

func (a AutoCompleteHandler) searchCommands(word string, commands []*cli.Command) []string {
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
	return a.removeDuplicates(result)
}

func (a AutoCompleteHandler) searchFlags(word string, command *cli.Command) []string {
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
	return a.removeDuplicates(result)
}

func (a AutoCompleteHandler) removeDuplicates(values []string) []string {
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
