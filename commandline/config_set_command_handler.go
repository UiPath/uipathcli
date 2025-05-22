package commandline

import (
	"fmt"
	"io"
	"strings"

	"github.com/UiPath/uipathcli/config"
)

// configSetCommandHandler implements command for setting config values.
//
// Example:
// uipath config set --key uri --value https://myserver
//
// The command also supports setting values in different profiles.
//
// Example:
// uipath config set --key uri --value https://myserver --profile onprem
type configSetCommandHandler struct {
	StdOut         io.Writer
	ConfigProvider config.ConfigProvider
}

const successfullySetMessage = "Successfully set config value"

const ConfigKeyServiceVersion = "serviceVersion"
const ConfigKeyOrganization = "organization"
const ConfigKeyTenant = "tenant"
const ConfigKeyUri = "uri"
const ConfigKeyInsecure = "insecure"
const ConfigKeyDebug = "debug"
const ConfigKeyAuthGrantType = "auth.grantType"
const ConfigKeyAuthScopes = "auth.scopes"
const ConfigKeyAuthUri = "auth.uri"
const ConfigKeyAuthProperties = "auth.properties."
const ConfigKeyHeader = "header."
const ConfigKeyParameter = "parameter."

var ConfigKeys = []string{
	ConfigKeyServiceVersion,
	ConfigKeyOrganization,
	ConfigKeyTenant,
	ConfigKeyUri,
	ConfigKeyInsecure,
	ConfigKeyDebug,
	ConfigKeyAuthGrantType,
	ConfigKeyAuthScopes,
	ConfigKeyAuthUri,
	ConfigKeyAuthProperties,
	ConfigKeyHeader,
	ConfigKeyParameter,
}

func (h configSetCommandHandler) Set(key string, value string, profileName string) error {
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

func (h configSetCommandHandler) setConfigValue(cfg *config.Config, key string, value string) error {
	keyParts := strings.Split(key, ".")
	if key == ConfigKeyServiceVersion {
		cfg.SetServiceVersion(value)
		return nil
	} else if key == ConfigKeyOrganization {
		cfg.SetOrganization(value)
		return nil
	} else if key == ConfigKeyTenant {
		cfg.SetTenant(value)
		return nil
	} else if key == ConfigKeyUri {
		return cfg.SetUri(value)
	} else if key == ConfigKeyInsecure {
		insecure, err := h.convertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for '%s': %w", ConfigKeyInsecure, err)
		}
		cfg.SetInsecure(insecure)
		return nil
	} else if key == ConfigKeyDebug {
		debug, err := h.convertToBool(value)
		if err != nil {
			return fmt.Errorf("Invalid value for '%s': %w", ConfigKeyDebug, err)
		}
		cfg.SetDebug(debug)
		return nil
	} else if key == ConfigKeyAuthGrantType {
		cfg.SetAuthGrantType(value)
		return nil
	} else if key == ConfigKeyAuthScopes {
		cfg.SetAuthScopes(value)
		return nil
	} else if key == ConfigKeyAuthUri {
		return cfg.SetAuthUri(value)
	} else if h.isHeaderKey(key, keyParts) {
		cfg.SetHeader(keyParts[1], value)
		return nil
	} else if h.isParameterKey(key, keyParts) {
		cfg.SetParameter(keyParts[1], value)
		return nil
	} else if h.isAuthPropertyKey(key, keyParts) {
		cfg.SetAuthProperty(keyParts[2], value)
		return nil
	}
	return fmt.Errorf("Unknown config key '%s'", key)
}

func (h configSetCommandHandler) isHeaderKey(key string, keyParts []string) bool {
	return strings.HasPrefix(key, ConfigKeyHeader) && len(keyParts) == 2
}

func (h configSetCommandHandler) isParameterKey(key string, keyParts []string) bool {
	return strings.HasPrefix(key, ConfigKeyParameter) && len(keyParts) == 2
}

func (h configSetCommandHandler) isAuthPropertyKey(key string, keyParts []string) bool {
	return strings.HasPrefix(key, ConfigKeyAuthProperties) && len(keyParts) == 3
}

func (h configSetCommandHandler) convertToBool(value string) (bool, error) {
	if strings.EqualFold(value, "true") {
		return true, nil
	}
	if strings.EqualFold(value, "false") {
		return false, nil
	}
	return false, fmt.Errorf("Invalid boolean value: %s", value)
}

func (h configSetCommandHandler) getOrCreateProfile(profileName string) config.Config {
	cfg := h.ConfigProvider.Config(profileName)
	if cfg == nil {
		return h.ConfigProvider.New()
	}
	return *cfg
}

func newConfigSetCommandHandler(
	stdOut io.Writer,
	configProvider config.ConfigProvider,
) *configSetCommandHandler {
	return &configSetCommandHandler{
		StdOut:         stdOut,
		ConfigProvider: configProvider,
	}
}
