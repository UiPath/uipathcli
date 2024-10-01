package commandline

import (
	"fmt"

	"github.com/UiPath/uipathcli/config"
)

const FlagNameDebug = "debug"
const FlagNameProfile = "profile"
const FlagNameUri = "uri"
const FlagNameOrganization = "organization"
const FlagNameTenant = "tenant"
const FlagNameInsecure = "insecure"
const FlagNameOutputFormat = "output"
const FlagNameQuery = "query"
const FlagNameWait = "wait"
const FlagNameWaitTimeout = "wait-timeout"
const FlagNameFile = "file"
const FlagNameIdentityUri = "identity-uri"
const FlagNameVersion = "version"
const FlagNameHelp = "help"

const FlagValueFromStdIn = "-"
const FlagValueOutputFormatJson = "json"
const FlagValueOutputFormatText = "text"

var FlagNamesPredefined = []string{
	FlagNameDebug,
	FlagNameProfile,
	FlagNameUri,
	FlagNameOrganization,
	FlagNameTenant,
	FlagNameInsecure,
	FlagNameOutputFormat,
	FlagNameQuery,
	FlagNameWait,
	FlagNameWaitTimeout,
	FlagNameFile,
	FlagNameIdentityUri,
	FlagNameVersion,
	FlagNameHelp,
}

// The FlagBuilder can be used to prepare a list of flags for a CLI command.
// The builder takes care that flags with the same name are deduped.
type FlagBuilder struct {
	flags map[string]*FlagDefinition
	order []string
}

func (b *FlagBuilder) AddFlag(flag *FlagDefinition) *FlagBuilder {
	name := flag.Name
	if _, found := b.flags[name]; !found {
		b.flags[name] = flag
		b.order = append(b.order, name)
	}
	return b
}

func (b *FlagBuilder) AddFlags(flags []*FlagDefinition) *FlagBuilder {
	for _, flag := range flags {
		b.AddFlag(flag)
	}
	return b
}

func (b *FlagBuilder) AddDefaultFlags(hidden bool) *FlagBuilder {
	b.AddFlags(b.defaultFlags(hidden))
	return b
}

func (b *FlagBuilder) AddHelpFlag() *FlagBuilder {
	b.AddFlag(b.helpFlag())
	return b
}

func (b *FlagBuilder) AddVersionFlag(hidden bool) *FlagBuilder {
	b.AddFlag(b.versionFlag(hidden))
	return b
}

func (b FlagBuilder) Build() []*FlagDefinition {
	flags := []*FlagDefinition{}
	for _, name := range b.order {
		flags = append(flags, b.flags[name])
	}
	return flags
}

func (b FlagBuilder) defaultFlags(hidden bool) []*FlagDefinition {
	return []*FlagDefinition{
		NewFlag(FlagNameDebug, "Enable debug output", FlagTypeBoolean).
			WithEnvVarName("UIPATH_DEBUG").
			WithDefaultValue(false).
			WithHidden(hidden),
		NewFlag(FlagNameProfile, "Config profile to use", FlagTypeString).
			WithEnvVarName("UIPATH_PROFILE").
			WithDefaultValue(config.DefaultProfile).
			WithHidden(hidden),
		NewFlag(FlagNameUri, "Server Base-URI", FlagTypeString).
			WithEnvVarName("UIPATH_URI").
			WithHidden(hidden),
		NewFlag(FlagNameOrganization, "Organization name", FlagTypeString).
			WithEnvVarName("UIPATH_ORGANIZATION").
			WithHidden(hidden),
		NewFlag(FlagNameTenant, "Tenant name", FlagTypeString).
			WithEnvVarName("UIPATH_TENANT").
			WithHidden(hidden),
		NewFlag(FlagNameInsecure, "Disable HTTPS certificate check", FlagTypeBoolean).
			WithEnvVarName("UIPATH_INSECURE").
			WithDefaultValue(false).
			WithHidden(hidden),
		NewFlag(FlagNameOutputFormat, fmt.Sprintf("Set output format: %s (default), %s", FlagValueOutputFormatJson, FlagValueOutputFormatText), FlagTypeString).
			WithEnvVarName("UIPATH_OUTPUT").
			WithDefaultValue("").
			WithHidden(hidden),
		NewFlag(FlagNameQuery, "Perform JMESPath query on output", FlagTypeString).
			WithDefaultValue("").
			WithHidden(hidden),
		NewFlag(FlagNameWait, "Waits for the provided condition (JMESPath expression)", FlagTypeString).
			WithDefaultValue("").
			WithHidden(hidden),
		NewFlag(FlagNameWaitTimeout, "Time to wait in seconds for condition", FlagTypeInteger).
			WithDefaultValue(30).
			WithHidden(hidden),
		NewFlag(FlagNameFile, "Provide input from file (use - for stdin)", FlagTypeString).
			WithDefaultValue("").
			WithHidden(hidden),
		NewFlag(FlagNameIdentityUri, "Identity Server URI", FlagTypeString).
			WithEnvVarName("UIPATH_IDENTITY_URI").
			WithHidden(hidden),
		b.versionFlag(hidden),
	}
}

func (b FlagBuilder) versionFlag(hidden bool) *FlagDefinition {
	return NewFlag(FlagNameVersion, "Specific service version", FlagTypeString).
		WithEnvVarName("UIPATH_VERSION").
		WithDefaultValue("").
		WithHidden(hidden)
}

func (b FlagBuilder) helpFlag() *FlagDefinition {
	return NewFlag(FlagNameHelp, "Show help", FlagTypeBoolean).
		WithDefaultValue(false).
		WithHidden(true)
}

func NewFlagBuilder() *FlagBuilder {
	return &FlagBuilder{map[string]*FlagDefinition{}, []string{}}
}
