package commandline

import (
	"bufio"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/network"
	"github.com/UiPath/uipathcli/utils/stream"
	"github.com/UiPath/uipathcli/utils/visualization"
)

// configCommandHandler implements command for configuring the CLI interactively.
type configCommandHandler struct {
	Input          stream.Stream
	StdOut         io.Writer
	Logger         log.Logger
	ConfigProvider config.ConfigProvider
	Authenticators []auth.Authenticator
}

const notSetMessage = "not set"
const maskMessage = "*******"
const successfullyConfiguredMessage = "Successfully configured uipath CLI"

const getTenantsTimeout = time.Duration(60) * time.Second
const getTenantsMaxAttempts = 3

func (h configCommandHandler) Configure(input configCommandInput) error {
	stdin, err := h.Input.Data()
	if err != nil {
		return err
	}
	defer func() { _ = stdin.Close() }()

	cfg := h.getOrCreateProfile(input.Profile)
	reader := bufio.NewReader(stdin)

	builder := config.NewConfigBuilder(cfg)
	authType := input.AuthType
	if authType == "" {
		authType = h.readAuthTypeInput(builder, reader)
	}
	err = h.readAuthInput(builder, reader, authType)
	if err != nil {
		return nil
	}

	organization, err := h.getOrganization(builder, input)
	if err != nil {
		return h.handleAuthError(err, builder, reader, input.Profile, authType)
	}
	if organization == nil {
		err := h.readOrgTenantInput(builder, reader)
		if err != nil {
			return nil
		}
		return h.updateConfigIfNeeded(builder, input.Profile)
	}
	err = h.readTenantInput(builder, reader, *organization)
	if err != nil {
		return nil
	}
	return h.updateConfigIfNeeded(builder, input.Profile)
}

func (h configCommandHandler) getOrganization(builder *config.ConfigBuilder, input configCommandInput) (*api.Organization, error) {
	cfg, _ := builder.Build()
	token, err := h.performAuth(input, cfg.Auth)
	if err != nil {
		return nil, err
	}
	if token == nil {
		return nil, nil
	}

	organizationId, err := h.getOrganizationIdFromToken(token)
	if err != nil {
		return nil, err
	}
	if organizationId == "" {
		return nil, nil
	}

	spinner := visualization.NewSpinner(h.Logger, "Tenant [loading...]: ")
	defer spinner.Close()
	settings := network.NewHttpClientSettings(input.Debug, input.OperationId, map[string]string{}, getTenantsTimeout, getTenantsMaxAttempts, input.Insecure)
	omsClient := api.NewOmsClient(input.BaseUri, token, *settings, h.Logger)
	return omsClient.GetOrganizationInfo(organizationId)
}

func (h configCommandHandler) getOrganizationIdFromToken(token *auth.AuthToken) (string, error) {
	jwtInfo, err := auth.NewJwtParser().Parse(token.Value)
	if err != nil {
		return "", err
	}
	return jwtInfo.PrtId, nil
}

func (h configCommandHandler) performAuth(input configCommandInput, authConfig map[string]interface{}) (*auth.AuthToken, error) {
	authContext := h.newAuthContext(input, authConfig)
	for _, authenticator := range h.Authenticators {
		result := authenticator.Auth(*authContext)
		if result.Error != nil {
			return nil, result.Error
		}
		if result.Token != nil {
			return result.Token, nil
		}
	}
	return nil, nil
}

func (h configCommandHandler) newAuthContext(input configCommandInput, authConfig map[string]interface{}) *auth.AuthenticatorContext {
	authRequest := auth.NewAuthenticatorRequest(input.BaseUri.String(), map[string]string{})
	return auth.NewAuthenticatorContext(authConfig, input.IdentityUri, input.OperationId, input.Insecure, input.Debug, *authRequest, h.Logger)
}

func (h configCommandHandler) handleAuthError(authErr error, builder *config.ConfigBuilder, reader *bufio.Reader, profileName string, authType string) error {
	errorMessageHint := ""
	switch authType {
	case config.AuthTypeLogin:
		errorMessageHint = "Please make sure the login information is correct or enter the organization and tenant information manually."
	default:
		errorMessageHint = "Please make sure the credentials are correct or enter the organization and tenant information manually."
	}
	h.Logger.LogError(fmt.Sprintf(`
Authentication Failed!
%v
%s

`, authErr, errorMessageHint))
	err := h.readOrgTenantInput(builder, reader)
	if err != nil {
		return nil
	}
	return h.updateConfigIfNeeded(builder, profileName)
}

func (h configCommandHandler) readAuthInput(builder *config.ConfigBuilder, reader *bufio.Reader, authType string) error {
	switch authType {
	case config.AuthTypeCredentials:
		return h.readCredentialsInput(builder, reader)
	case config.AuthTypeLogin:
		return h.readLoginInput(builder, reader)
	case config.AuthTypePat:
		return h.readPatInput(builder, reader)
	case "":
		return nil
	}
	return fmt.Errorf("Invalid auth, supported values: %s, %s, %s", config.AuthTypeCredentials, config.AuthTypeLogin, config.AuthTypePat)
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

func (h configCommandHandler) readAuthTypeInput(builder *config.ConfigBuilder, reader *bufio.Reader) string {
	authType := builder.Config.AuthType()
	options := []string{
		"credentials - Client Id and Client Secret",
		"login - OAuth login using the browser",
		"pat - Personal Access Token",
	}
	i, _, err := h.readInput(reader, "Authentication type", authType, options)
	if err != nil {
		return ""
	}
	switch i {
	case 0:
		return config.AuthTypeCredentials
	case 1:
		return config.AuthTypeLogin
	case 2:
		return config.AuthTypePat
	default:
		return authType
	}
}

func (h configCommandHandler) readTenantInput(builder *config.ConfigBuilder, reader *bufio.Reader, organization api.Organization) error {
	if len(organization.Tenants) == 1 {
		builder.
			WithOrganization(&organization.Name).
			WithTenant(&organization.Tenants[0].Name)
		return nil
	}

	tenants := []string{}
	for _, tenant := range organization.Tenants {
		tenants = append(tenants, tenant.Name)
	}
	sort.Strings(tenants)
	i, selected, err := h.readInput(reader, "Tenant", builder.Config.Tenant, tenants)
	if err != nil {
		return err
	}

	if i >= 0 {
		builder.
			WithOrganization(&organization.Name).
			WithTenant(&selected)
	}
	return nil
}

func (h configCommandHandler) readInput(reader *bufio.Reader, message string, value string, options []string) (int, string, error) {
	optionList := ""
	for i, option := range options {
		optionList += fmt.Sprintf("  [%d] %s\n", i+1, option)
	}

	for {
		message := fmt.Sprintf(`%s [%s]:
%sSelect:`, message, h.getDisplayValue(value, false), optionList)
		input, err := h.readUserInput(message, reader)
		if err != nil {
			return -1, "", err
		}
		if input == nil {
			return -1, "", nil
		}
		i, err := strconv.Atoi(*input)
		if err == nil && i <= len(options) {
			selectedIndex := i - 1
			selectedValue := options[selectedIndex]
			return selectedIndex, selectedValue, nil
		}
	}
}

func newConfigCommandHandler(
	input stream.Stream,
	stdOut io.Writer,
	logger log.Logger,
	configProvider config.ConfigProvider,
	authenticators []auth.Authenticator,
) *configCommandHandler {
	return &configCommandHandler{
		Input:          input,
		StdOut:         stdOut,
		Logger:         logger,
		ConfigProvider: configProvider,
		Authenticators: authenticators,
	}
}
