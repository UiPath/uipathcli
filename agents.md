# UiPath CLI — Coding Agent Reference

Complete reference for AI coding agents to use the `uipath` CLI for automating
UiPath workflows: managing solutions, agents, packages, orchestrator resources,
and document understanding.

## Quick Setup

```bash
# Install (Linux amd64)
curl -sL "https://github.com/UiPath/uipathcli/releases/latest/download/uipathcli-linux-amd64.tar.gz" | tar -xzv

# Authenticate (pick one)
uipath config --auth login        # OAuth browser login (interactive)
uipath config --auth credentials  # Client credentials (automated/CI)
uipath config --auth pat          # Personal access token

# Verify
uipath orchestrator users get
```

### Environment Variable Authentication (CI/CD)

```bash
export UIPATH_ORGANIZATION="my-org"
export UIPATH_TENANT="DefaultTenant"
export UIPATH_PAT="rt_..."                        # Option A: PAT
# OR
export UIPATH_CLIENT_ID="..." UIPATH_CLIENT_SECRET="..."  # Option B: Credentials
```

### Config File (`~/.uipath/config`)

```yaml
profiles:
  - name: default
    organization: my-org
    tenant: DefaultTenant
    auth:
      clientId: <id>
      clientSecret: <secret>
  - name: alpha
    uri: https://alpha.uipath.com
    organization: my-org
    tenant: DefaultTenant
    auth:
      pat: rt_...
```

Use `--profile alpha` or `UIPATH_PROFILE=alpha` to switch profiles.

---

## Command Structure

```
uipath <service> [<category>] <operation> [--param value] [global-flags]
```

### Global Flags (All Commands)

| Flag | Env Variable | Default | Description |
|------|-------------|---------|-------------|
| `--debug` | `UIPATH_DEBUG` | false | Show HTTP request/response details |
| `--profile` | `UIPATH_PROFILE` | default | Config profile name |
| `--uri` | `UIPATH_URI` | https://cloud.uipath.com | Server base URL |
| `--organization` | `UIPATH_ORGANIZATION` | | Organization name |
| `--tenant` | `UIPATH_TENANT` | | Tenant name |
| `--output` | `UIPATH_OUTPUT` | json | Output format: `json`, `text` |
| `--query` | | | JMESPath expression for output filtering |
| `--wait` | | | JMESPath condition to poll until true |
| `--wait-timeout` | | 30 | Seconds to wait before timeout |
| `--file` | | | Input file path (use `-` for stdin) |
| `--insecure` | `UIPATH_INSECURE` | false | Skip TLS cert verification |
| `--identity-uri` | `UIPATH_IDENTITY_URI` | | Identity server URL |
| `--call-timeout` | `UIPATH_CALL_TIMEOUT` | 60 | HTTP call timeout (seconds) |
| `--max-attempts` | `UIPATH_MAX_ATTEMPTS` | 3 | Retry count for failed requests |

---

## Complete Command Reference

### Studio Solution Commands

Manage UiPath Maestro solutions (.uis files). Solutions are containers holding
agent projects, processes, web apps, and other project types.

#### `uipath studio solution pack`
Pack a solution directory into a .uis file (ZIP archive).

```bash
uipath studio solution pack --source ./MySolution --destination ./output.uis
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | `.` | Path to solution directory |
| `--destination` | string | no | `<dirname>.uis` | Output .uis file path |

**Requires:** `SolutionStorage.json` in source directory.
**Excludes:** `.git/`, `__pycache__/`, `*.pyc` (includes `.agent-builder/`, `.project/`).

**Output:**
```json
{"status":"Succeeded","package":"/path/to/output.uis","solutionId":"...","name":"","size":12345}
```

---

#### `uipath studio solution unpack`
Extract a .uis file into a solution directory.

```bash
uipath studio solution unpack --source ./solution.uis --destination ./MySolution
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | | Path to .uis file |
| `--destination` | string | no | `<basename>` | Output directory |

**Output:**
```json
{"status":"Succeeded","directory":"/path/to/MySolution","solutionId":"...","projectCount":1}
```

---

#### `uipath studio solution push`
Upload a .uis solution file to UiPath Studio Web.

```bash
uipath studio solution push --source ./solution.uis
uipath studio solution push --source ./solution.uis --solution-id abc-123  # Update existing
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | | Path to .uis file |
| `--solution-id` | string | no | | Solution ID to update (omit for new) |

**Requires:** `--organization`

---

#### `uipath studio solution pull`
Download a solution from Studio Web as a .uis file.

```bash
uipath studio solution pull --solution-id abc-123 --destination ./solution.uis
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--solution-id` | string | yes | | Solution ID to download |
| `--destination` | string | no | `<solution-id>.uis` | Output file path |

**Requires:** `--organization`

---

#### `uipath studio solution list`
List all solutions in Studio Web.

```bash
uipath studio solution list
uipath studio solution list --query "solutions[?status == 'active']"
```

**Requires:** `--organization`

---

#### `uipath studio solution publish`
Publish a solution in Studio Web for deployment to Orchestrator.

```bash
uipath studio solution publish --solution-id abc-123
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--solution-id` | string | yes | | Solution ID to publish |

**Requires:** `--organization`

---

### Studio Package Commands

Build, analyze, test, and publish UiPath Studio automation projects (.nupkg).

#### `uipath studio package pack`
Package a Studio project into a .nupkg file.

```bash
uipath studio package pack --source ./MyProject --destination ./output
uipath studio package pack --auto-version true
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | `.` | Path to project.json or folder |
| `--destination` | string | yes | `.` | Output folder |
| `--package-version` | string | no | | Specific version string |
| `--auto-version` | boolean | no | false | Auto-generate version |
| `--output-type` | string | no | | Force type: Process, Library, Tests, Objects |
| `--split-output` | boolean | no | false | Split runtime and design libraries |
| `--release-notes` | string | no | | Release notes |

**Output:**
```json
{"status":"Succeeded","package":"/path/to/Package.1.0.0.nupkg","name":"MyProject","description":"...","projectId":"...","version":"1.0.0"}
```

---

#### `uipath studio package publish`
Publish a .nupkg package to Orchestrator and create/update a release.

```bash
uipath studio package publish --source ./MyProject.1.0.0.nupkg --folder Shared
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | `.` | Path to .nupkg file or directory |
| `--folder` | string | no | `Shared` | Orchestrator folder name |
| `--folder-id` | integer | no | | Folder ID (alternative to name) |

**Requires:** `--organization`, `--tenant`

---

#### `uipath studio package analyze`
Run static analysis on a project using governance rules.

```bash
uipath studio package analyze --source ./MyProject
uipath studio package analyze --query "violations[?severity == 'Error'].[errorCode, description]" --output text
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | `.` | Path to project.json or folder |
| `--stop-on-rule-violation` | boolean | no | true | Exit with error on violations |
| `--treat-warnings-as-errors` | boolean | no | false | Treat warnings as errors |
| `--governance-file` | string | no | `uipath.policy.default.json` | Governance policy file |

**Output:**
```json
{"status":"Succeeded","violations":[{"errorCode":"ST-USG-010","severity":"Warning","description":"...","filePath":"Main.xaml"}]}
```

---

#### `uipath studio package restore`
Restore project dependencies.

```bash
uipath studio package restore --source ./MyProject --destination ./packages
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string | yes | `.` | Path to project.json or folder |
| `--destination` | string | yes | `./packages` | Output folder for dependencies |

---

#### `uipath studio test run`
Run test cases on connected Orchestrator, with multi-project parallel support.

```bash
uipath studio test run --source ./MyProject
uipath studio test run --source "./project1,./project2" --attach-robot-logs true --results-output junit
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--source` | string[] | yes | `.` | Comma-separated project paths |
| `--timeout` | integer | no | 3600 | Max wait time (seconds) |
| `--results-output` | string | no | `uipath` | Output format: `uipath`, `junit` |
| `--attach-robot-logs` | boolean | no | false | Attach robot logs to results |
| `--folder` | string | no | `Shared` | Orchestrator folder |
| `--folder-id` | integer | no | | Folder ID (hidden) |

**Requires:** `--organization`, `--tenant`

---

### Orchestrator Commands

#### `uipath orchestrator buckets download`
Download a file from an Orchestrator storage bucket.

```bash
uipath orchestrator buckets download --folder-id 2000021 --key 12345 --path "documents/invoice.pdf"
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--folder-id` | integer | yes | | Folder/OrgUnit ID |
| `--key` | integer | yes | | Bucket ID |
| `--path` | string | yes | | File path in bucket |

**Requires:** `--organization`, `--tenant`

---

#### `uipath orchestrator buckets upload`
Upload a file to an Orchestrator storage bucket.

```bash
uipath orchestrator buckets upload --folder-id 2000021 --key 12345 --path "docs/report.pdf" --file ./report.pdf
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--folder-id` | integer | yes | | Folder/OrgUnit ID |
| `--key` | integer | yes | | Bucket ID |
| `--path` | string | yes | | Target path in bucket |
| `--file` | binary | yes | | File to upload (use `-` for stdin) |

**Requires:** `--organization`, `--tenant`

---

#### Auto-Generated Orchestrator Commands

The CLI auto-generates commands from the Orchestrator OpenAPI specification.
Common operations include:

```bash
# Folders
uipath orchestrator folders get
uipath orchestrator folders get --query "value[0].Id"

# Users
uipath orchestrator users get
uipath orchestrator users get --query "value[?Type == 'User']"

# Jobs
uipath orchestrator jobs start-jobs --folder-id <id> --start-info "ReleaseKey=<key>"
uipath orchestrator jobs get-by-id --folder-id <id> --key <jobId>
uipath orchestrator jobs stop-jobs --folder-id <id> --strategy SoftStop --job-ids "123,456"

# Assets
uipath orchestrator assets get --folder-id <id>
uipath orchestrator assets post --folder-id <id> --name "MyAsset" --value-type Text --string-value "value"

# Releases
uipath orchestrator releases get --folder-id <id>

# Queues
uipath orchestrator queue-items get --folder-id <id>
uipath orchestrator queue-items post --folder-id <id> --item-data "Name=MyQueue"

# Processes
uipath orchestrator processes get --folder-id <id>

# Machines
uipath orchestrator machines get

# Robots
uipath orchestrator robots get --folder-id <id>

# Logs
uipath orchestrator robot-logs get --folder-id <id>
```

---

### Document Understanding Commands

#### `uipath du digitization digitize`
Digitize a document (synchronous wrapper over async API).

```bash
uipath du digitization digitize --file invoice.pdf
uipath du digitization digitize --project-id "abc-123" --file invoice.jpg --content-type "image/jpeg"
```

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `--project-id` | string | no | `00000000-...` | DU project ID |
| `--file` | binary | yes | | File to digitize |
| `--content-type` | string | no | `application/octet-stream` | MIME type |

**Requires:** `--organization`, `--tenant`

#### Auto-Generated DU Commands

```bash
# Digitization
uipath du digitization start --file invoice.pdf
uipath du digitization get --document-id <id> --wait "status == 'Succeeded'"

# Classification
uipath du classification classify --file - --query "classificationResults[0].DocumentTypeId"

# Extraction
uipath du extraction extract --file - --query "extractionResult.ResultsDocument.Fields[?not_null(Values)]"

# Generative Extraction
uipath du extraction extract --project-id "00000000-0000-0000-0000-000000000001" \
  --extractor-id "generative_extractor" --document-id "$docId" \
  --prompts "id=total; question=The total amount"

# Projects
uipath du discovery projects
```

---

## Input & Output Patterns

### Parameter Types

| Type | Example | Notes |
|------|---------|-------|
| string | `--name "My Asset"` | |
| integer | `--folder-id 12345` | |
| boolean | `--auto-version true` | |
| binary | `--file invoice.pdf` | File path, or `-` for stdin |
| string[] | `--source "p1,p2"` | Comma-separated or repeated flags |
| object | `--start-info "Key=val;Key2=val2"` | Semicolon-separated, or raw JSON |

### Nested Objects

```bash
# Semicolon syntax
uipath orchestrator jobs start-jobs --folder-id 2000021 \
  --start-info "ReleaseKey=abc-123; RunAsMe=false; RuntimeType=Unattended"

# JSON syntax
uipath orchestrator jobs start-jobs --folder-id 2000021 \
  --start-info '{"releaseKey":"abc-123","runAsMe":false,"runtimeType":"Unattended"}'
```

### Piping / stdin

```bash
# Pipe file to command
cat invoice.pdf | uipath du digitization digitize --file - --content-type "application/pdf"

# Chain commands
uipath du digitization start --file invoice.jpg | uipath du classification classify --file -

# Here-doc
uipath orchestrator jobs start-jobs --folder-id 2000021 --file - <<EOF
{"startInfo":{"releaseKey":"abc-123","runAsMe":false}}
EOF
```

### JMESPath Queries

```bash
# Extract single field
uipath orchestrator folders get --query "value[0].Id"

# Filter results
uipath orchestrator users get --query "value[?Type == 'User']"

# Multi-field projection
uipath orchestrator users get --query "value[].[Name, CreationTime]" --output text

# Sort
uipath orchestrator users get --query "sort_by(value, &CreationTime) | [-1].Name"

# Count
uipath orchestrator folders get --query "length(value)"
```

### Wait for Async Operations

```bash
# Poll until condition is true
documentId=$(uipath du digitization start --file invoice.jpg --query "documentId" --output text)
uipath du digitization get --document-id $documentId --wait "status == 'Succeeded'" --wait-timeout 300

# Wait for job completion
uipath orchestrator jobs get-by-id --folder-id $fid --key $jobId \
  --wait "Status == 'Completed' || Status == 'Faulted'"
```

### Text Output Format

```bash
# Tab-separated output (useful for shell processing)
uipath orchestrator users get --query "value[].[Name, CreationTime]" --output text | \
  while IFS=$'\t' read -r name creation_time; do
    echo "User ${name} was created at ${creation_time}"
  done
```

---

## Complete Workflow Examples

### 1. Build, Publish, and Run a Studio Project

```bash
# Build the package
uipath studio package pack --source ./MyProject --destination ./output --auto-version true

# Publish to Orchestrator
uipath studio package publish --source ./output --folder Shared

# Get folder ID and start a job
folderId=$(uipath orchestrator folders get --query "value[0].Id")
jobId=$(uipath orchestrator jobs start-jobs \
  --folder-id "$folderId" \
  --start-info "ReleaseName=MyProject" \
  --query "value[0].Id")

# Wait for job to complete
uipath orchestrator jobs get-by-id --folder-id "$folderId" --key "$jobId" \
  --wait "Status == 'Completed' || Status == 'Faulted'" --wait-timeout 300
```

### 2. Create and Deploy an Agent Solution

```bash
# 1. Create solution directory structure locally (or use the skill script)
mkdir -p MySolution/Agent

# 2. Pack the solution into .uis
uipath studio solution pack --source ./MySolution --destination ./MySolution.uis

# 3. Push to Studio Web
uipath studio solution push --source ./MySolution.uis

# 4. List to find the solution ID
solutionId=$(uipath studio solution list --query "solutions[0].solutionId" --output text)

# 5. Publish for deployment
uipath studio solution publish --solution-id "$solutionId"
```

### 3. Pull, Modify, and Re-push a Solution

```bash
# Pull existing solution
uipath studio solution pull --solution-id abc-123 --destination ./solution.uis

# Unpack
uipath studio solution unpack --source ./solution.uis --destination ./MySolution

# Make changes to agent files (agent.json, resources, etc.)
# ...

# Repack and push
uipath studio solution pack --source ./MySolution --destination ./updated.uis
uipath studio solution push --source ./updated.uis --solution-id abc-123
```

### 4. Analyze and Test a Project

```bash
# Static analysis with governance rules
uipath studio package analyze --source ./MyProject \
  --query "violations[?severity == 'Error'].[errorCode, description]" --output text

# Run tests with JUnit output
uipath studio test run --source ./MyProject --results-output junit --attach-robot-logs true

# Multi-project parallel test run
uipath studio test run --source "./project1,./project2" --timeout 600
```

### 5. Document Processing Pipeline

```bash
# Digitize → Classify → Extract
documentId=$(uipath du digitization start --file invoice.pdf --query "documentId" --output text)
uipath du digitization get --document-id "$documentId" --wait "status == 'Succeeded'"

# Classify
docType=$(uipath du classification classify --file - --query "classificationResults[0].DocumentTypeId" --output text <<< "$documentId")

# Extract fields
uipath du extraction extract --document-id "$documentId" \
  --query "extractionResult.ResultsDocument.Fields[?not_null(Values)].{field: FieldId, value: Values[0].Value, confidence: Values[0].Confidence}"
```

### 6. Manage Orchestrator Resources

```bash
folderId=$(uipath orchestrator folders get --query "value[0].Id")

# Create asset
uipath orchestrator assets post --folder-id "$folderId" \
  --name "ApiKey" --value-scope "Global" --value-type "Text" --string-value "sk-..."

# List assets
uipath orchestrator assets get --folder-id "$folderId" --query "value[].{Name: Name, Type: ValueType}"

# Upload file to bucket
uipath orchestrator buckets upload --folder-id "$folderId" --key 1 --path "data/input.csv" --file ./input.csv

# Download from bucket
uipath orchestrator buckets download --folder-id "$folderId" --key 1 --path "data/output.csv"
```

---

## Solution File Structure

A UiPath solution (.uis) is a ZIP archive with this structure:

```
MySolution/
├── SolutionStorage.json          # Solution metadata: {SolutionId, Projects: [...]}
├── MySolution.uipx               # Manifest: project types and paths
├── Agent/                        # Agent project
│   ├── agent.json                # Model, prompts, I/O schemas, type
│   ├── entry-points.json         # Entry point definitions
│   ├── project.uiproj            # {ProjectType: "Agent"}
│   ├── flow-layout.json          # Visual layout (low-code)
│   ├── .agent-builder/
│   │   ├── agent.json            # Builder metadata with resources
│   │   └── bindings.json         # Connection bindings
│   ├── .project/
│   │   └── JitCustomTypes.json
│   ├── resources/                # Agent tools
│   │   └── <ToolName>/
│   │       └── resource.json     # Tool definition
│   └── evals/                    # Evaluations (low-code)
│       ├── eval-sets/*.json
│       └── evaluators/*.json
└── resources/solution_folder/    # Deployment resources
    ├── package/<Name>.json       # Package resource
    ├── process/agent/<Name>.json # Process resource
    ├── connection/...            # Connection resources
    └── index/...                 # Index resources
```

### Agent Types

**Low-code** (`type: "lowCode"`): Visual builder with system/user prompts,
contentTokens, flow-layout.json, .agent-builder/ directory.

**Coded** (`type: "coded"`, `targetRuntime: "python"`): Python entry point at
`source_code/main.py` with `@traced` decorators, pydantic models, `uipath` SDK.
Evals go in `coded-evals/` instead of `evals/`.

### Supported Models

```
anthropic.claude-haiku-4-5-20251001-v1:0
anthropic.claude-sonnet-4-20250514-v1:0
gpt-4.1-2025-04-14
gpt-4.1-mini-2025-04-14
gpt-4o-2024-11-20
gpt-4o-mini-2024-07-18
gemini-2.5-flash-preview-04-17
gemini-2.0-flash-001
```

### Tool/Resource Types

| Type | $resourceType | Use Case |
|------|--------------|----------|
| Integration | `tool` (external) | Web Search, Web Reader, API calls |
| Agent | `tool` (solution) | Agent-calling-agent |
| Internal | `tool` (built-in) | Analyze Files |
| Context | `context` | RAG/Index semantic search |
| Escalation | `escalation` | HITL via Action Center |

### Project Types (in .uipx manifest)

Agent, Process, WebApp, CaseManagement, BusinessRules, Connector,
ProcessOrchestration, Api

---

## Authentication Reference

### Priority Order
1. **PAT** (`UIPATH_PAT` or `auth.pat`) — highest priority
2. **OAuth Login** (`auth.clientId` + `auth.redirectUri` + `auth.scopes`)
3. **Client Credentials** (`UIPATH_CLIENT_ID`/`UIPATH_CLIENT_SECRET` or config)

### Scope Requirements

| Operation | Required Scopes |
|-----------|----------------|
| Orchestrator read | `OR.Users.Read`, `OR.Folders.Read`, etc. |
| Orchestrator write | `OR.Jobs.Write`, `OR.Assets.Write`, etc. |
| Studio operations | Studio-specific scopes |
| Document Understanding | `Du.Digitization`, `Du.Classification`, etc. |

### Config Commands

```bash
uipath config                                   # Interactive setup
uipath config --auth credentials                # Client credentials flow
uipath config --auth login                      # OAuth browser flow
uipath config --auth pat                        # Personal access token
uipath config set --key organization --value x  # Set individual values
uipath config set --key uri --value "https://alpha.uipath.com" --profile alpha
uipath config cache clear                       # Clear token cache
```

---

## Error Handling

All commands return JSON with a `status` field (`Succeeded` or `Failed`).
Failed operations also include an `error` field.

```bash
# Check command exit code
uipath studio package pack --source ./MyProject
if [ $? -ne 0 ]; then
  echo "Pack failed"
fi

# Check status in output
status=$(uipath studio package publish --source ./pkg.nupkg --query "status" --output text)
if [ "$status" = "Failed" ]; then
  error=$(uipath studio package publish --source ./pkg.nupkg --query "error" --output text)
  echo "Publish failed: $error"
fi
```

### Common Errors

| Error | Cause | Fix |
|-------|-------|-----|
| `Organization is not set` | Missing --organization | Set in config or CLI flag |
| `Tenant is not set` | Missing --tenant | Set in config or CLI flag |
| `Package not found` | Invalid --source path | Check file exists |
| `Service returned status code '503'` | Server error with retry exhaustion | Check service status, increase --max-attempts |
| `Package already exists` | Version conflict on publish | Use --auto-version or bump version |

---

## Tips for Coding Agents

1. **Always capture IDs**: Use `--query` and `--output text` to capture IDs for
   chaining:
   ```bash
   folderId=$(uipath orchestrator folders get --query "value[0].Id" --output text)
   ```

2. **Use `--wait` for async ops**: Don't poll manually:
   ```bash
   uipath du digitization get --document-id $id --wait "status == 'Succeeded'"
   ```

3. **Debug failures**: Add `--debug` to see full HTTP request/response.

4. **Solution vs Package**: Solutions (.uis) are Maestro containers with agents.
   Packages (.nupkg) are traditional Studio automation projects.

5. **Studio Web needs only org**: Solution push/pull/list/publish require
   `--organization` but NOT `--tenant` (unlike Orchestrator commands).

6. **Pipe commands**: Chain operations using pipes and `--file -`:
   ```bash
   uipath du digitization start --file doc.pdf | uipath du extraction extract --file -
   ```

7. **JMESPath is powerful**: Filter, sort, project, and transform output without
   external tools like jq.

8. **Pack before push**: Always pack a solution directory before pushing:
   ```bash
   uipath studio solution pack --source ./Sol && uipath studio solution push --source ./Sol.uis
   ```
