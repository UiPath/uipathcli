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

const notSetMessage = "not set"
const maskMessage = "*******"

const CredentialsAuth = "credentials"
const LoginAuth = "login"

func (h ConfigCommandHandler) Configure(auth string, profileName string) error {
	switch auth {
	case CredentialsAuth:
		return h.configureCredentials(profileName)
	case LoginAuth:
		return h.configureLogin(profileName)
	}
	return fmt.Errorf("Invalid auth, supported values: %s, %s", CredentialsAuth, LoginAuth)
}

func (h ConfigCommandHandler) configureCredentials(profileName string) error {
	config := h.getOrCreateProfile(profileName)

	message := fmt.Sprintf("Enter client id [%s]:", h.getDisplayValue(config.ClientId(), true))
	clientId, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter client secret [%s]:", h.getDisplayValue(config.ClientSecret(), true))
	clientSecret, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter organization [%s]:", h.getDisplayValue(config.Organization(), false))
	organization, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter tenant [%s]:", h.getDisplayValue(config.Tenant(), false))
	tenant, err := h.readUserInput(message)
	if err != nil {
		return nil
	}

	authChanged := config.ConfigureCredentialsAuth(clientId, clientSecret)
	orgTenantChanged := config.ConfigureOrgTenant(organization, tenant)

	if authChanged || orgTenantChanged {
		err = h.ConfigProvider.Update(profileName, config.Auth.Config, config.Path)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, "Successfully configured uipathcli")
	}
	return nil
}

func (h ConfigCommandHandler) configureLogin(profileName string) error {
	config := h.getOrCreateProfile(profileName)

	message := fmt.Sprintf("Enter client id [%s]:", h.getDisplayValue(config.ClientId(), true))
	clientId, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter redirect uri [%s]:", h.getDisplayValue(config.RedirectUri(), false))
	redirectUri, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter scopes [%s]:", h.getDisplayValue(config.Scopes(), false))
	scopes, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter organization [%s]:", h.getDisplayValue(config.Organization(), false))
	organization, err := h.readUserInput(message)
	if err != nil {
		return nil
	}
	message = fmt.Sprintf("Enter tenant [%s]:", h.getDisplayValue(config.Tenant(), false))
	tenant, err := h.readUserInput(message)
	if err != nil {
		return nil
	}

	authChanged := config.ConfigureLoginAuth(clientId, redirectUri, scopes)
	orgTenantChanged := config.ConfigureOrgTenant(organization, tenant)

	if authChanged || orgTenantChanged {
		err = h.ConfigProvider.Update(profileName, config.Auth.Config, config.Path)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, "Successfully configured uipathcli")
	}
	return nil
}

func (h ConfigCommandHandler) getOrCreateProfile(profileName string) config.Config {
	config := h.ConfigProvider.Config(profileName)
	if config == nil {
		return h.ConfigProvider.New()
	}
	return *config
}

func (h ConfigCommandHandler) getDisplayValue(value string, masked bool) string {
	if value == "" {
		return notSetMessage
	}
	if masked {
		return h.maskValue(value)
	}
	return value
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
