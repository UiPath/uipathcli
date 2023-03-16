package commandline

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/UiPath/uipathcli/config"
)

// The ConfigCommandHandler implements commands for configuring the CLI.
// The CLI can be configured interactively or by setting config values
// programmatically.
//
// Example:
// uipath config ==> interactive configuration of the CLI
// uipath config set ==> stores a value in the configuration file
type ConfigCommandHandler struct {
	StdIn          io.Reader
	StdOut         io.Writer
	ConfigProvider config.ConfigProvider
}

const notSetMessage = "not set"
const maskMessage = "*******"
const successfullyConfiguredMessage = "Successfully configured uipath CLI"
const successfullySetMessage = "Successfully set config value"

const CredentialsAuth = "credentials"
const LoginAuth = "login"
const PatAuth = "pat"

func (h ConfigCommandHandler) Set(key string, value string, profileName string) error {
	config := h.getOrCreateProfile(profileName)
	err := h.setConfigValue(&config, key, value)
	if err != nil {
		return err
	}
	err = h.ConfigProvider.Update(profileName, config)
	if err != nil {
		return err
	}
	fmt.Fprintln(h.StdOut, successfullySetMessage)
	return nil
}

func (h ConfigCommandHandler) setConfigValue(config *config.Config, key string, value string) error {
	keyParts := strings.Split(key, ".")
	if key == "organization" {
		config.ConfigureOrgTenant(value, "")
		return nil
	} else if key == "tenant" {
		config.ConfigureOrgTenant("", value)
		return nil
	} else if key == "uri" {
		return config.SetUri(value)
	} else if key == "insecure" {
		insecure, err := h.convertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'insecure': %w", err)
		}
		config.SetInsecure(insecure)
		return nil
	} else if key == "debug" {
		debug, err := h.convertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'debug': %w", err)
		}
		config.SetDebug(debug)
		return nil
	} else if key == "auth.grantType" {
		config.SetAuthGrantType(value)
		return nil
	} else if key == "auth.scopes" {
		config.SetAuthScopes(value)
		return nil
	} else if h.isHeaderKey(keyParts) {
		config.SetHeader(keyParts[1], value)
		return nil
	} else if h.isPathKey(keyParts) {
		config.SetPath(keyParts[1], value)
		return nil
	} else if h.isQueryKey(keyParts) {
		config.SetQuery(keyParts[1], value)
		return nil
	} else if h.isAuthPropertyKey(keyParts) {
		config.SetAuthProperty(keyParts[2], value)
		return nil
	}
	return fmt.Errorf("Unknown config key '%s'", key)
}

func (h ConfigCommandHandler) isHeaderKey(keyParts []string) bool {
	return len(keyParts) == 2 && keyParts[0] == "header"
}

func (h ConfigCommandHandler) isPathKey(keyParts []string) bool {
	return len(keyParts) == 2 && keyParts[0] == "path"
}

func (h ConfigCommandHandler) isQueryKey(keyParts []string) bool {
	return len(keyParts) == 2 && keyParts[0] == "query"
}

func (h ConfigCommandHandler) isAuthPropertyKey(keyParts []string) bool {
	return len(keyParts) == 3 && keyParts[0] == "auth" && keyParts[1] == "properties"
}

func (h ConfigCommandHandler) convertToBool(value string) (bool, error) {
	if strings.EqualFold(value, "true") {
		return true, nil
	}
	if strings.EqualFold(value, "false") {
		return false, nil
	}
	return false, fmt.Errorf("Invalid boolean value: %s", value)
}

func (h ConfigCommandHandler) Configure(auth string, profileName string) error {
	switch auth {
	case CredentialsAuth:
		return h.configureCredentials(profileName)
	case LoginAuth:
		return h.configureLogin(profileName)
	case PatAuth:
		return h.configurePat(profileName)
	case "":
		return h.configure(profileName)
	}
	return fmt.Errorf("Invalid auth, supported values: %s, %s, %s", CredentialsAuth, LoginAuth, PatAuth)
}

func (h ConfigCommandHandler) configure(profileName string) error {
	config := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	organization, tenant, err := h.readOrgTenantInput(config, reader)
	if err != nil {
		return nil
	}

	authChanged := h.readAuthInput(config, reader)
	orgTenantChanged := config.ConfigureOrgTenant(organization, tenant)

	if orgTenantChanged || authChanged {
		err = h.ConfigProvider.Update(profileName, config)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, successfullyConfiguredMessage)
	}
	return nil
}

func (h ConfigCommandHandler) readAuthInput(config config.Config, reader *bufio.Reader) bool {
	authType := h.readAuthTypeInput(config, reader)
	switch authType {
	case CredentialsAuth:
		clientId, clientSecret, err := h.readCredentialsInput(config, reader)
		if err != nil {
			return false
		}
		return config.ConfigureCredentialsAuth(clientId, clientSecret)
	case LoginAuth:
		clientId, redirectUri, scopes, err := h.readLoginInput(config, reader)
		if err != nil {
			return false
		}
		return config.ConfigureLoginAuth(clientId, redirectUri, scopes)
	case PatAuth:
		pat, err := h.readPatInput(config, reader)
		if err != nil {
			return false
		}
		return config.ConfigurePatAuth(pat)
	default:
		return false
	}
}

func (h ConfigCommandHandler) configureCredentials(profileName string) error {
	config := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	organization, tenant, err := h.readOrgTenantInput(config, reader)
	if err != nil {
		return nil
	}
	clientId, clientSecret, err := h.readCredentialsInput(config, reader)
	if err != nil {
		return nil
	}

	orgTenantChanged := config.ConfigureOrgTenant(organization, tenant)
	authChanged := config.ConfigureCredentialsAuth(clientId, clientSecret)

	if orgTenantChanged || authChanged {
		err = h.ConfigProvider.Update(profileName, config)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, successfullyConfiguredMessage)
	}
	return nil
}

func (h ConfigCommandHandler) configureLogin(profileName string) error {
	config := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	organization, tenant, err := h.readOrgTenantInput(config, reader)
	if err != nil {
		return nil
	}
	clientId, redirectUri, scopes, err := h.readLoginInput(config, reader)
	if err != nil {
		return nil
	}

	orgTenantChanged := config.ConfigureOrgTenant(organization, tenant)
	authChanged := config.ConfigureLoginAuth(clientId, redirectUri, scopes)

	if orgTenantChanged || authChanged {
		err = h.ConfigProvider.Update(profileName, config)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, successfullyConfiguredMessage)
	}
	return nil
}

func (h ConfigCommandHandler) configurePat(profileName string) error {
	config := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	organization, tenant, err := h.readOrgTenantInput(config, reader)
	if err != nil {
		return nil
	}
	pat, err := h.readPatInput(config, reader)
	if err != nil {
		return nil
	}

	orgTenantChanged := config.ConfigureOrgTenant(organization, tenant)
	authChanged := config.ConfigurePatAuth(pat)

	if orgTenantChanged || authChanged {
		err = h.ConfigProvider.Update(profileName, config)
		if err != nil {
			return err
		}
		fmt.Fprintln(h.StdOut, successfullyConfiguredMessage)
	}
	return nil
}

func (h ConfigCommandHandler) getAuthType(config config.Config) string {
	if config.Pat() != "" {
		return PatAuth
	}
	if config.ClientId() != "" && config.RedirectUri() != "" && config.Scopes() != "" {
		return LoginAuth
	}
	if config.ClientId() != "" && config.ClientSecret() != "" {
		return CredentialsAuth
	}
	return ""
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

func (h ConfigCommandHandler) readUserInput(message string, reader *bufio.Reader) (string, error) {
	fmt.Fprint(h.StdOut, message+" ")
	value, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.Trim(value, " \r\n\t"), nil
}

func (h ConfigCommandHandler) readOrgTenantInput(config config.Config, reader *bufio.Reader) (string, string, error) {
	message := fmt.Sprintf("Enter organization [%s]:", h.getDisplayValue(config.Organization, false))
	organization, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", err
	}

	message = fmt.Sprintf("Enter tenant [%s]:", h.getDisplayValue(config.Tenant, false))
	tenant, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", err
	}

	return organization, tenant, nil
}

func (h ConfigCommandHandler) readCredentialsInput(config config.Config, reader *bufio.Reader) (string, string, error) {
	message := fmt.Sprintf("Enter client id [%s]:", h.getDisplayValue(config.ClientId(), true))
	clientId, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", err
	}

	message = fmt.Sprintf("Enter client secret [%s]:", h.getDisplayValue(config.ClientSecret(), true))
	clientSecret, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", err
	}

	return clientId, clientSecret, nil
}

func (h ConfigCommandHandler) readLoginInput(config config.Config, reader *bufio.Reader) (string, string, string, error) {
	message := fmt.Sprintf("Enter client id [%s]:", h.getDisplayValue(config.ClientId(), true))
	clientId, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", "", err
	}
	message = fmt.Sprintf("Enter redirect uri [%s]:", h.getDisplayValue(config.RedirectUri(), false))
	redirectUri, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", "", err
	}
	message = fmt.Sprintf("Enter scopes [%s]:", h.getDisplayValue(config.Scopes(), false))
	scopes, err := h.readUserInput(message, reader)
	if err != nil {
		return "", "", "", err
	}

	return clientId, redirectUri, scopes, nil
}

func (h ConfigCommandHandler) readPatInput(config config.Config, reader *bufio.Reader) (string, error) {
	message := fmt.Sprintf("Enter personal access token [%s]:", h.getDisplayValue(config.Pat(), true))
	return h.readUserInput(message, reader)
}

func (h ConfigCommandHandler) readAuthTypeInput(config config.Config, reader *bufio.Reader) string {
	authType := h.getAuthType(config)
	for {
		message := fmt.Sprintf(`Authentication type [%s]:
  [1] credentials - Client Id and Client Secret
  [2] login - OAuth login using the browser
  [3] pat - Personal Access Token
Select:`, h.getDisplayValue(authType, false))
		input, err := h.readUserInput(message, reader)
		if err != nil {
			return ""
		}
		switch input {
		case "":
			return authType
		case "1":
			return CredentialsAuth
		case "2":
			return LoginAuth
		case "3":
			return PatAuth
		}
	}
}
