package commandline

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/UiPath/uipathcli/config"
)

type ConfigCommandHandler struct {
	StdIn          io.Reader
	StdOut         io.Writer
	ConfigProvider config.ConfigProvider
}

var notSetMessage = "not set"
var maskMessage = "*******"

func (h ConfigCommandHandler) Configure(profileName string) error {
	config := h.ConfigProvider.Config(profileName)
	message := fmt.Sprintf("Enter client id [%s]:", h.authConfigValue("clientId", config))
	clientId, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter client secret [%s]:", h.authConfigValue("clientSecret", config))
	clientSecret, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter organization [%s]:", h.pathValue("organization", config))
	organization, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter tenant [%s]:", h.pathValue("tenant", config))
	tenant, err := h.readUserInput(message)
	if err != nil {
		return nil
	}

	if clientId != "" || clientSecret != "" || organization != "" || tenant != "" {
		err = h.ConfigProvider.Update(profileName, clientId, clientSecret, organization, tenant)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, "Successfully configured uipathcli")
	}
	return nil
}

func (h ConfigCommandHandler) authConfigValue(name string, config *config.Config) string {
	value := ""
	if config != nil && config.Auth.Config[name] != nil {
		value, _ = config.Auth.Config[name].(string)
	}
	if value != "" {
		return h.maskValue(value)
	}
	return notSetMessage
}

func (h ConfigCommandHandler) pathValue(name string, config *config.Config) string {
	if config != nil && config.Path[name] != "" {
		return config.Path[name]
	}
	return notSetMessage
}

func (h ConfigCommandHandler) maskValue(value string) string {
	if len(value) < 10 {
		return maskMessage
	}
	return maskMessage + value[len(value)-4:]
}

func (h ConfigCommandHandler) readUserInput(message string) (string, error) {
	reader := bufio.NewReader(h.StdIn)
	fmt.Fprint(h.StdOut, message+" ")
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(value, " \r\n\t"), nil
}
