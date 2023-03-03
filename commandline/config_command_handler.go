package commandline

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
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
const successfullyConfiguredMessage = "Successfully configured uipath CLI"
const successfullySetMessage = "Successfully set config value"

const CredentialsAuth = "credentials"
const LoginAuth = "login"
const PatAuth = "pat"

func (h ConfigCommandHandler) Set(key string, value string, profileName string) error {
	config := h.getOrCreateProfile(profileName)

	keyParts := strings.Split(key, ".")
	if key == "organization" {
		config.Organization = value
	} else if key == "tenant" {
		config.Tenant = value
	} else if key == "uri" {
		uri, err := url.Parse(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'uri': %v", err)
		}
		config.Uri = uri
	} else if key == "insecure" {
		insecure, err := h.ConvertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'insecure': %v", err)
		}
		config.Insecure = insecure
	} else if key == "debug" {
		debug, err := h.ConvertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'debug': %v", err)
		}
		config.Debug = debug
	} else if len(keyParts) == 2 && keyParts[0] == "header" {
		config.Header[keyParts[1]] = value
	} else if len(keyParts) == 2 && keyParts[0] == "path" {
		config.Path[keyParts[1]] = value
	} else if len(keyParts) == 2 && keyParts[0] == "query" {
		config.Query[keyParts[1]] = value
	} else if key == "auth.grantType" {
		config.Auth.Config["grantType"] = value
	} else if key == "auth.scopes" {
		config.Auth.Config["scopes"] = value
	} else if len(keyParts) == 3 && keyParts[0] == "auth" && keyParts[1] == "properties" {
		properties, ok := config.Auth.Config["properties"].(map[interface{}]interface{})
		if properties == nil || !ok {
			properties = map[interface{}]interface{}{}
		}
		properties[keyParts[2]] = value
		config.Auth.Config["properties"] = properties
	} else {
		return fmt.Errorf("Unknown config key '%s'", key)
	}

	err := h.ConfigProvider.Update(profileName, config)
	if err != nil {
		return err
	}
	fmt.Fprintln(h.StdOut, successfullySetMessage)
	return nil
}

func (h ConfigCommandHandler) ConvertToBool(value string) (bool, error) {
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
	authType, err := h.readAuthTypeInput(config, reader)
	if err != nil {
		return nil
	}

	authChanged := false
	if authType == CredentialsAuth {
		clientId, clientSecret, err := h.readCredentialsInput(config, reader)
		if err != nil {
			return nil
		}
		authChanged = config.ConfigureCredentialsAuth(clientId, clientSecret)
	} else if authType == LoginAuth {
		clientId, redirectUri, scopes, err := h.readLoginInput(config, reader)
		if err != nil {
			return nil
		}
		authChanged = config.ConfigureLoginAuth(clientId, redirectUri, scopes)
	} else if authType == PatAuth {
		pat, err := h.readPatInput(config, reader)
		if err != nil {
			return nil
		}
		authChanged = config.ConfigurePatAuth(pat)
	}

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

func (h ConfigCommandHandler) readAuthTypeInput(config config.Config, reader *bufio.Reader) (string, error) {
	authType := h.getAuthType(config)
	for {
		message := fmt.Sprintf(`Authentication type [%s]:
  [1] credentials - Client Id and Client Secret
  [2] login - OAuth login using the browser
  [3] pat - Personal Access Token
Select:`, h.getDisplayValue(authType, false))
		input, err := h.readUserInput(message, reader)
		if err != nil {
			return "", nil
		}
		switch input {
		case "":
			return authType, nil
		case "1":
			return CredentialsAuth, nil
		case "2":
			return LoginAuth, nil
		case "3":
			return PatAuth, nil
		}
	}
}
