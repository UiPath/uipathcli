package commandline

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/UiPath/uipathcli/auth"
	"github.com/UiPath/uipathcli/config"
	"github.com/UiPath/uipathcli/log"
	"github.com/UiPath/uipathcli/utils/api"
	"github.com/UiPath/uipathcli/utils/network"
)

// configCommandHandler implements command for configuring the CLI interactively.
type configCommandHandler struct {
	StdIn          io.Reader
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
	cfg := h.getOrCreateProfile(input.Profile)
	reader := bufio.NewReader(h.StdIn)

	builder := config.NewConfigBuilder(cfg)
	err := h.readAuthInput(builder, reader, input.AuthType)
	if err != nil {
		return nil
	}

	updatedCfg, _ := builder.Build()

	token, err := h.performAuth(input, updatedCfg)
	if err != nil {
		return h.handleAuthError(err, input.Profile, builder, reader)
	}

	organization, err := h.getOrganizationInfo(input, token)
	if err != nil {
		return h.handleAuthError(err, input.Profile, builder, reader)
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

func (h configCommandHandler) getOrganizationInfo(input configCommandInput, token *auth.AuthToken) (*api.Organization, error) {
	if token == nil {
		return nil, nil
	}
	jwtInfo, err := auth.NewJwtParser().Parse(token.Value)
	if err != nil {
		return nil, err
	}
	if jwtInfo.PartId == "" {
		return nil, nil
	}

	settings := network.NewHttpClientSettings(input.Debug, input.OperationId, map[string]string{}, getTenantsTimeout, getTenantsMaxAttempts, input.Insecure)
	omsClient := api.NewOmsClient(input.BaseUri, token, *settings, h.Logger)
	return omsClient.GetOrganizationInfo(jwtInfo.PartId)
}

func (h configCommandHandler) performAuth(input configCommandInput, cfg config.Config) (*auth.AuthToken, error) {
	authContext := h.newAuthContext(input, cfg.Auth)
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

func (h configCommandHandler) handleAuthError(authErr error, profileName string, builder *config.ConfigBuilder, reader *bufio.Reader) error {
	h.Logger.LogError(fmt.Sprintf(`Could not retrieve organization details: %v
	
Please make sure the credentials are correct or enter the organization and tenant information manually:
`, authErr))
	err := h.readOrgTenantInput(builder, reader)
	if err != nil {
		return nil
	}
	return h.updateConfigIfNeeded(builder, profileName)
}

func (h configCommandHandler) readAuthInput(builder *config.ConfigBuilder, reader *bufio.Reader, authType string) error {
	if authType == "" {
		authType = h.readAuthTypeInput(builder.Config, reader)
	}
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

func (h configCommandHandler) readTenantInput(builder *config.ConfigBuilder, reader *bufio.Reader, organization api.Organization) error {
	tenant := builder.Config.Tenant

	tenantList := ""
	for i, tenant := range organization.Tenants {
		tenantList += fmt.Sprintf("  [%d] %s\n", i+1, tenant.Name)
	}

	for {
		message := fmt.Sprintf(`Tenant [%s]:
%sSelect:`, h.getDisplayValue(tenant, false), tenantList)
		input, err := h.readUserInput(message, reader)
		if err != nil {
			return err
		}
		if input == nil {
			return nil
		}
		i, err := strconv.Atoi(*input)
		if err == nil && i <= len(organization.Tenants) {
			tenant := organization.Tenants[i-1]
			builder.
				WithOrganization(&organization.Name).
				WithTenant(&tenant.Name)
			return nil
		}
	}
}

func newConfigCommandHandler(
	stdIn io.Reader,
	stdOut io.Writer,
	logger log.Logger,
	configProvider config.ConfigProvider,
	authenticators []auth.Authenticator,
) *configCommandHandler {
	return &configCommandHandler{
		StdIn:          stdIn,
		StdOut:         stdOut,
		Logger:         logger,
		ConfigProvider: configProvider,
		Authenticators: authenticators,
	}
}
