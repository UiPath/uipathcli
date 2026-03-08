# UiPath CLI Architecture Reference

How the UiPath CLI (`uipathcli`) is structured and how to extend it.

## Overview

The CLI is written in Go and uses two execution models:
1. **OpenAPI-generated commands** — auto-generated from YAML definitions
2. **Plugin commands** — hand-crafted Go code for complex operations

## Directory Structure

```
uipathcli/
  main.go                       # Entry point, registers plugins
  definitions/                  # Embedded OpenAPI YAML specs
    orchestrator.yaml           # Orchestrator API (45K lines)
    du.framework.yaml           # Document Understanding API
    identity.yaml               # Identity Server API
    identity.token.yaml         # Token endpoint
    studio.yaml                 # Studio (currently empty)
  plugin/                       # Plugin command implementations
    command_plugin.go           # CommandPlugin interface
    command.go                  # Command metadata struct
    execution_context.go        # Plugin execution context
    digitizer/                  # DU digitize command
    orchestrator/
      download/                 # Bucket download
      upload/                   # Bucket upload
    studio/
      pack/                     # Package pack
      publish/                  # Package publish
      analyze/                  # Package analyze
      restore/                  # Package restore
      testrun/                  # Test run
  commandline/                  # CLI framework
    cli.go                      # Main CLI runner
    command_builder.go          # Command tree builder
    definition_provider.go      # Definition loading
  executor/                     # Execution engines
    http_executor.go            # For OpenAPI commands
    plugin_executor.go          # For plugin commands
  auth/                         # Authentication
    pat_authenticator.go        # Personal Access Token
    oauth_authenticator.go      # Browser-based OAuth
    bearer_authenticator.go     # Client credentials
  config/                       # Configuration management
  parser/                       # OpenAPI parser
  output/                       # Response formatting
```

## Plugin Interface

To add a new command, implement `plugin.CommandPlugin`:

```go
type CommandPlugin interface {
    Command() Command
    Execute(ctx ExecutionContext, writer output.OutputWriter, logger log.Logger) error
}
```

### Command Metadata

```go
func NewCommand(service string) *Command

cmd := plugin.NewCommand("studio").
    WithCategory("solution", "Solution management", "Manage UiPath solutions").
    WithOperation("pack", "Pack solution", "Package solution directory into .uis").
    WithParameter(plugin.NewParameter("source", plugin.ParameterTypeString, "Source directory", true)).
    WithParameter(plugin.NewParameter("output", plugin.ParameterTypeString, "Output .uis file", false))
```

### Parameter Types

| Type | Constant |
|------|----------|
| String | `plugin.ParameterTypeString` |
| Integer | `plugin.ParameterTypeInteger` |
| Boolean | `plugin.ParameterTypeBoolean` |
| Binary (file) | `plugin.ParameterTypeBinary` |
| String Array | `plugin.ParameterTypeStringArray` |

### Execution Context

```go
type ExecutionContext struct {
    Organization string
    Tenant       string
    BaseUri      url.URL
    Auth         AuthToken
    Parameters   []ExecutionParameter
    Debug        bool
    Settings     map[string]interface{}
}
```

Access parameters:
```go
source := ctx.Parameters.Get("source")  // returns string value
```

### Registration in main.go

```go
cli := commandline.NewCli(
    // ...
    *commandline.NewDefinitionProvider(
        // ...
        []plugin.CommandPlugin{
            // Existing plugins
            plugin_studio_pack.NewPackagePackCommand(),
            // New plugins
            plugin_studio_solution.NewSolutionPackCommand(),
            plugin_studio_solution.NewSolutionUnpackCommand(),
            plugin_studio_agent.NewAgentInitCommand(),
        },
    ),
    // ...
)
```

## Configuration

### Config File: `~/.uipath/config`

```yaml
profiles:
  - name: default
    organization: my-org
    tenant: my-tenant
    uri: https://cloud.uipath.com
    auth:
      clientId: <id>
      clientSecret: <secret>
    header: {}
    parameter: {}
  - name: alpha
    organization: my-org
    tenant: my-tenant
    uri: https://alpha.uipath.com
    auth:
      properties:
        - grantType: authorization_code
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `UIPATH_PROFILE` | Active profile name |
| `UIPATH_ORGANIZATION` | Organization override |
| `UIPATH_TENANT` | Tenant override |
| `UIPATH_CLIENT_ID` | Client ID override |
| `UIPATH_CLIENT_SECRET` | Client secret override |
| `UIPATH_PAT` | Personal Access Token |
| `UIPATH_URI` | Base URI override |
| `UIPATH_OUTPUT` | Output format (json/text) |
| `UIPATH_DEBUG` | Enable debug logging |
| `UIPATH_INSECURE` | Skip TLS verification |
| `UIPATH_CONFIGURATION_PATH` | Config file path |
| `UIPATH_DEFINITIONS_PATH` | Definitions directory path |

### Authentication Methods

1. **PAT** — `UIPATH_PAT` env var or `auth.pat` in config
2. **OAuth** — Browser-based login, caches tokens in `~/.uipath/cache/`
3. **Client Credentials** — `clientId` + `clientSecret` → Bearer token
4. **Login** — `auth.properties.grantType: authorization_code`

Auth is tried in order: PAT → OAuth → Bearer. First success wins.

## Existing CLI Commands

### Studio Package Commands
```bash
uipath studio package pack --source <dir> [--output-type Process|Library] [--auto-version] [--output <path>]
uipath studio package publish --source <nupkg> [--organization-feed] [--tenant-feed]
uipath studio package restore --source <dir>
uipath studio package analyze --source <dir> [--governance-file <path>] [--treat-warnings-as-errors]
uipath studio test run --source <dir> [--junit-results <path>] [--uipath-results <path>] [--attach-robot-logs]
```

### Orchestrator Commands
```bash
uipath orchestrator buckets upload --folder-id <id> --bucket-id <key> --path <remote> --file <local>
uipath orchestrator buckets download --folder-id <id> --bucket-id <key> --path <remote>
uipath orchestrator jobs start-jobs --folder-id <id> --start-info "..."
uipath orchestrator releases get --folder-id <id>
uipath orchestrator processes upload-package --folder-id <id> --file <nupkg>
```

### Document Understanding Commands
```bash
uipath du digitization digitize --project-id <id> --file <path>
uipath du extraction extract --project-id <id> --document-id <id> --extractor-id <id>
uipath du classification classify --project-id <id> --document-id <id> --classifier-id <id>
```

### Configuration Commands
```bash
uipath config                    # Interactive configuration
uipath config --auth login       # OAuth browser login
uipath config --auth credentials # Client credentials setup
uipath config --auth pat         # PAT setup
```

### Common Flags
```bash
--profile <name>       # Select profile
--output json|text     # Output format
--query <jmespath>     # JMESPath query on output
--uri <url>            # Override base URI
--debug                # Show HTTP request/response
--insecure             # Skip TLS verification
--wait                 # Wait for async operations
--wait-timeout <sec>   # Wait timeout
--file @<path>         # Read file as input
```

## Adding OpenAPI-Generated Commands

To add auto-generated commands, populate the YAML definition file:

1. Download the Swagger spec:
   ```bash
   curl -o swagger.json https://alpha.uipath.com/{org}/studio_/backend/swagger/v1/swagger.json
   ```

2. Convert to OpenAPI 3.0 YAML (if needed)

3. Place in `definitions/studio.yaml` (or create `definitions/studio.web.yaml`)

4. Custom parameter naming via `x-uipathcli-name` extension:
   ```yaml
   parameters:
     - name: X-UIPATH-OrganizationUnitId
       x-uipathcli-name: folder-id
   ```

5. Custom operation naming via `x-uipathcli-name` extension on operations

The parser automatically generates CLI commands from the OpenAPI paths.

## URL Construction

Default: `https://cloud.uipath.com/{organization}/{tenant}/{service}_/...`

Services: `orchestrator_`, `du_`, `studio_`, `identity_`

For alpha: `https://alpha.uipath.com/{organization}/{tenant}/{service}_/...`

Override with `--uri` flag or `uri` in profile config.

## Output Transformation

Use JMESPath queries to transform output:

```bash
uipath orchestrator releases get --folder-id 123 --query "value[].{Name:Name,Key:Key}"
```

Text output is tab-separated for Unix tool compatibility:

```bash
uipath orchestrator folders get --output text | awk -F'\t' '{print $2}'
```
