package commandline

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/UiPath/uipathcli/config"
)

// configCommandHandler implements commands for configuring the CLI.
// The CLI can be configured interactively or by setting config values
// programmatically.
//
// Example:
// uipath config ==> interactive configuration of the CLI
// uipath config set ==> stores a value in the configuration file
type configCommandHandler struct {
	StdIn          io.Reader
	StdOut         io.Writer
	ConfigProvider config.ConfigProvider
}

const notSetMessage = "not set"
const maskMessage = "*******"
const successfullyConfiguredMessage = "Successfully configured uipath CLI"
const successfullySetMessage = "Successfully set config value"

func (h configCommandHandler) Set(key string, value string, profileName string) error {
	cfg := h.getOrCreateProfile(profileName)
	err := h.setConfigValue(&cfg, key, value)
	if err != nil {
		return err
	}
	err = h.ConfigProvider.Update(profileName, cfg)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(h.StdOut, successfullySetMessage)
	return nil
}

func (h configCommandHandler) setConfigValue(cfg *config.Config, key string, value string) error {
	keyParts := strings.Split(key, ".")
	if key == "serviceVersion" {
		cfg.SetServiceVersion(value)
		return nil
	} else if key == "organization" {
		cfg.SetOrganization(value)
		return nil
	} else if key == "tenant" {
		cfg.SetTenant(value)
		return nil
	} else if key == "uri" {
		return cfg.SetUri(value)
	} else if key == "insecure" {
		insecure, err := h.convertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'insecure': %w", err)
		}
		cfg.SetInsecure(insecure)
		return nil
	} else if key == "debug" {
		debug, err := h.convertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for 'debug': %w", err)
		}
		cfg.SetDebug(debug)
		return nil
	} else if key == "auth.grantType" {
		cfg.SetAuthGrantType(value)
		return nil
	} else if key == "auth.scopes" {
		cfg.SetAuthScopes(value)
		return nil
	} else if h.isHeaderKey(keyParts) {
		cfg.SetHeader(keyParts[1], value)
		return nil
	} else if h.isParameterKey(keyParts) {
		cfg.SetParameter(keyParts[1], value)
		return nil
	} else if h.isAuthPropertyKey(keyParts) {
		cfg.SetAuthProperty(keyParts[2], value)
		return nil
	}
	return fmt.Errorf("Unknown config key '%s'", key)
}

func (h configCommandHandler) isHeaderKey(keyParts []string) bool {
	return len(keyParts) == 2 && keyParts[0] == "header"
}

func (h configCommandHandler) isParameterKey(keyParts []string) bool {
	return len(keyParts) == 2 && keyParts[0] == "parameter"
}

func (h configCommandHandler) isAuthPropertyKey(keyParts []string) bool {
	return len(keyParts) == 3 && keyParts[0] == "auth" && keyParts[1] == "properties"
}

func (h configCommandHandler) convertToBool(value string) (bool, error) {
	if strings.EqualFold(value, "true") {
		return true, nil
	}
	if strings.EqualFold(value, "false") {
		return false, nil
	}
	return false, fmt.Errorf("Invalid boolean value: %s", value)
}

func (h configCommandHandler) Configure(auth string, profileName string) error {
	switch auth {
	case config.AuthTypeCredentials:
		return h.configureCredentials(profileName)
	case config.AuthTypeLogin:
		return h.configureLogin(profileName)
	case config.AuthTypePat:
		return h.configurePat(profileName)
	case "":
		return h.configure(profileName)
	}
	return fmt.Errorf("Invalid auth, supported values: %s, %s, %s", config.AuthTypeCredentials, config.AuthTypeLogin, config.AuthTypePat)
}

func (h configCommandHandler) configure(profileName string) error {
	cfg := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	builder := config.NewConfigBuilder(cfg)
	err := h.readOrgTenantInput(builder, reader)
	if err != nil {
		return nil
	}
	err = h.readAuthInput(builder, reader)
	if err != nil {
		return nil
	}

	return h.updateConfigIfNeeded(builder, profileName)
}

func (h configCommandHandler) readAuthInput(builder *config.ConfigBuilder, reader *bufio.Reader) error {
	authType := h.readAuthTypeInput(builder.Config, reader)
	switch authType {
	case config.AuthTypeCredentials:
		return h.readCredentialsInput(builder, reader)
	case config.AuthTypeLogin:
		return h.readLoginInput(builder, reader)
	case config.AuthTypePat:
		return h.readPatInput(builder, reader)
	default:
		return nil
	}
}

func (h configCommandHandler) configureCredentials(profileName string) error {
	cfg := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	builder := config.NewConfigBuilder(cfg)
	err := h.readOrgTenantInput(builder, reader)
	if err != nil {
		return nil
	}
	err = h.readCredentialsInput(builder, reader)
	if err != nil {
		return nil
	}

	return h.updateConfigIfNeeded(builder, profileName)
}

func (h configCommandHandler) configureLogin(profileName string) error {
	cfg := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	builder := config.NewConfigBuilder(cfg)
	err := h.readOrgTenantInput(builder, reader)
	if err != nil {
		return nil
	}
	err = h.readLoginInput(builder, reader)
	if err != nil {
		return nil
	}

	return h.updateConfigIfNeeded(builder, profileName)
}

func (h configCommandHandler) configurePat(profileName string) error {
	cfg := h.getOrCreateProfile(profileName)
	reader := bufio.NewReader(h.StdIn)

	builder := config.NewConfigBuilder(cfg)
	err := h.readOrgTenantInput(builder, reader)
	if err != nil {
		return nil
	}
	err = h.readPatInput(builder, reader)
	if err != nil {
		return nil
	}

	return h.updateConfigIfNeeded(builder, profileName)
}

func (h configCommandHandler) updateConfigIfNeeded(builder *config.ConfigBuilder, profileName string) error {
	updatedConfig, changed := builder.Build()
	if !changed {
		return nil
	}

	err := h.ConfigProvider.Update(profileName, updatedConfig)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintln(h.StdOut, successfullyConfiguredMessage)
	return nil
}

func (h configCommandHandler) getOrCreateProfile(profileName string) config.Config {
	cfg := h.ConfigProvider.Config(profileName)
	if cfg == nil {
		return h.ConfigProvider.New()
	}
	return *cfg
}

func (h configCommandHandler) getDisplayValue(value string, masked bool) string {
	if value == "" {
		return notSetMessage
	}
	if masked {
		return h.maskValue(value)
	}
	return value
}

func (h configCommandHandler) maskValue(value string) string {
	if len(value) < 10 {
		return maskMessage
	}
	return maskMessage + value[len(value)-4:]
}

func (h configCommandHandler) readUserInput(message string, reader *bufio.Reader) (*string, error) {
	_, err := fmt.Fprint(h.StdOut, message+" ")
	if err != nil {
		return nil, err
	}
	value, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	value = strings.Trim(value, "\r\n")
	if value == "" {
		return nil, nil
	}
	value = strings.Trim(value, " \t")
	return &value, nil
}

func (h configCommandHandler) readOrgTenantInput(builder *config.ConfigBuilder, reader *bufio.Reader) error {
	cfg := builder.Config
	message := fmt.Sprintf("Enter organization [%s]:", h.getDisplayValue(cfg.Organization, false))
	organization, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}

	message = fmt.Sprintf("Enter tenant [%s]:", h.getDisplayValue(cfg.Tenant, false))
	tenant, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}

	builder.
		WithOrganization(organization).
		WithTenant(tenant)
	return nil
}

func (h configCommandHandler) readCredentialsInput(builder *config.ConfigBuilder, reader *bufio.Reader) error {
	cfg := builder.Config
	message := fmt.Sprintf("Enter client id [%s]:", h.getDisplayValue(cfg.ClientId(), true))
	clientId, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}

	message = fmt.Sprintf("Enter client secret [%s]:", h.getDisplayValue(cfg.ClientSecret(), true))
	clientSecret, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}

	builder.WithCredentials(clientId, clientSecret)
	return nil
}

func (h configCommandHandler) readLoginInput(builder *config.ConfigBuilder, reader *bufio.Reader) error {
	cfg := builder.Config
	message := fmt.Sprintf("Enter client id [%s]:", h.getDisplayValue(cfg.ClientId(), true))
	clientId, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}
	message = fmt.Sprintf("Enter client secret (only for confidential apps) [%s]:", h.getDisplayValue(cfg.ClientSecret(), true))
	clientSecret, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}
	message = fmt.Sprintf("Enter redirect uri [%s]:", h.getDisplayValue(cfg.RedirectUri(), false))
	redirectUri, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}
	message = fmt.Sprintf("Enter scopes [%s]:", h.getDisplayValue(cfg.Scopes(), false))
	scopes, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}

	builder.WithLogin(clientId, clientSecret, redirectUri, scopes)
	return nil
}

func (h configCommandHandler) readPatInput(builder *config.ConfigBuilder, reader *bufio.Reader) error {
	cfg := builder.Config
	message := fmt.Sprintf("Enter personal access token [%s]:", h.getDisplayValue(cfg.Pat(), true))
	pat, err := h.readUserInput(message, reader)
	if err != nil {
		return err
	}

	builder.WithPat(pat)
	return nil
}

func (h configCommandHandler) readAuthTypeInput(cfg config.Config, reader *bufio.Reader) string {
	authType := cfg.AuthType()
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
		if input == nil {
			return authType
		}
		switch *input {
		case "":
		case "1":
			return config.AuthTypeCredentials
		case "2":
			return config.AuthTypeLogin
		case "3":
			return config.AuthTypePat
		}
	}
}

func newConfigCommandHandler(stdIn io.Reader, stdOut io.Writer, configProvider config.ConfigProvider) *configCommandHandler {
	return &configCommandHandler{
		StdIn:          stdIn,
		StdOut:         stdOut,
		ConfigProvider: configProvider,
	}
}
