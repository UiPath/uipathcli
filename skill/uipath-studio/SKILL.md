---
name: UiPath Studio
description: >
  This skill should be used when the user asks to "create a UiPath agent",
  "build an agent", "scaffold an agent project", "create a solution",
  "pack a solution", "unpack a .uis file", "deploy to UiPath",
  "publish a solution", "add a tool to an agent", "create evaluations",
  "add an evaluator", "set agent model", "set agent prompt",
  "create a coded agent", "create a low-code agent",
  "work with UiPath agents", "work with UiPath Maestro",
  "create a Maestro project", "create an evaluation set",
  "add a web search tool", "configure agent model",
  "set up agent evaluations",
  "add escalation", "add HITL", "create index", "add context resource",
  "validate agent structure", "create UiPath connection",
  or mentions UiPath Studio Web, Maestro, .uis files, agent.json,
  entry-points.json, project.uiproj, evaluation sets, or the UiPath CLI
  studio commands. Provides comprehensive guidance for creating, evaluating,
  deploying, and managing UiPath agents and solutions from the command line.
version: 0.1.0
---

# UiPath Studio — Agent & Solution Development

Create, evaluate, deploy, and manage UiPath agents and Maestro solutions
entirely from the command line.

## Core Concepts

### Solutions
A **Solution** is the top-level container in UiPath Maestro. It holds multiple
projects of different types: Agent, Process, WebApp, CaseManagement,
BusinessRules, Connector, ProcessOrchestration, Api.

A `.uis` file is a ZIP archive containing the entire solution.

### Agents
Two flavors exist:
- **Low-code** (`type: lowCode`) — visual builder with system/user prompts,
  tool bindings, and flow layout
- **Coded** (`type: coded`, `targetRuntime: python`) — Python entry point with
  `@traced` decorators, pydantic models, and the `uipath` SDK

### Evaluations
Every agent supports evals with built-in evaluator types (exact-match,
contains, json-similarity, llm-judge, trajectory) and custom Python evaluators.

## Workflow: Creating an Agent Solution

### Step 1: Scaffold the Solution

Generate `SolutionStorage.json` and `<Name>.uipx` manifest. Use the script:

```bash
bash $SKILL_DIR/scripts/solution_create.sh "MySolution" "Agent"
```

Or create manually following `examples/solution/`.

### Step 2: Scaffold the Agent Project

For low-code agents, generate the full directory structure with `agent.json`,
`entry-points.json`, `project.uiproj`, and default evals. Reference:
`examples/low-code-agent/`

For coded agents, also generate `source_code/main.py`, `pyproject.toml`, and
`uipath.json`. Reference: `examples/coded-agent/`

Key file: `agent.json` — controls model, prompts, I/O schemas, tool bindings.
See `references/agent-structure.md` for all fields.

### Step 3: Add Tools & Resources

Add tools to the agent's `resources/` directory. Each tool is a subdirectory
with a `resource.json`. Five tool types exist:

| Type | resourceType | Use Case |
|------|-------------|----------|
| Integration | `tool` (external) | Web Search, Web Reader, API calls |
| Agent | `tool` (solution) | Agent-calling-agent |
| Internal | `tool` (built-in) | Analyze Files |
| Context | `context` | RAG/Index semantic search |
| Escalation | `escalation` | HITL via Action Center |

See `references/tool-types.md` for complete schemas and examples.

### Step 4: Configure Evaluations

Create eval sets in `evals/eval-sets/` and evaluators in `evals/evaluators/`.
See `references/evaluation-framework.md` for all evaluator types and schemas.

### Step 5: Pack & Deploy

**Prerequisite**: Authentication must be configured before deploying. Run
`uipath config --auth login` for interactive setup or configure
`~/.uipath/config` manually. See `references/cli-architecture.md` for details.

Pack the solution directory into a `.uis` file:

```bash
bash $SKILL_DIR/scripts/solution_pack.sh ./MySolution MySolution.uis
```

Deploy using the UiPath CLI or Studio Web API. See
`references/studio-web-api.md` for deployment endpoints.

## File Structure Quick Reference

### Low-Code Agent
```
Agent/
  agent.json                    # Model, prompts, settings, I/O schemas
  entry-points.json             # Entry point definitions
  project.uiproj                # {ProjectType: "Agent"}
  flow-layout.json              # Visual layout
  .agent-builder/agent.json     # Builder metadata with resources
  .agent-builder/bindings.json  # Connection bindings
  .project/JitCustomTypes.json
  resources/<Tool>/resource.json
  evals/eval-sets/*.json
  evals/evaluators/*.json
```

### Coded Agent
```
Agent/
  agent.json                    # targetRuntime: "python", type: "coded"
  entry-points.json
  project.uiproj
  source_code/main.py           # @traced async def main(input) -> Output
  source_code/pyproject.toml    # uipath>=2.1.87
  source_code/uipath.json
  coded-evals/eval-sets/*.json
  coded-evals/evaluators/*.json
  coded-evals/evaluators/custom/ # Python BaseEvaluator classes
```

### Solution Wrapper
```
<SolutionName>/
  SolutionStorage.json          # {SolutionId, Projects: [...]}
  <SolutionName>.uipx           # Manifest with project types
  Agent/                        # Agent project(s)
  resources/solution_folder/    # Deployment resources
    package/<Name>.json         # Package resource
    process/agent/<Name>.json   # Process resource
    connection/...              # Connection resources
    index/...                   # Index resources
```

## UiPath CLI Commands

The CLI binary is `uipath`. Existing studio commands:

```bash
uipath studio package pack --source <dir> [--output-type Process|Library]
uipath studio package publish --source <nupkg> --feed <url>
uipath studio package restore --source <dir>
uipath studio package analyze --source <dir> [--governance-file <path>]
uipath studio test run --source <dir> [--junit-results <path>]
```

Configuration: `~/.uipath/config` with profiles. Auth via PAT, OAuth, or
client credentials. See `references/cli-architecture.md`.

## Validation

Validate an agent project structure before packing:

```bash
bash $SKILL_DIR/scripts/validate_agent.sh ./Agent
```

## Additional Resources

### Reference Files
- **`references/agent-structure.md`** — Complete agent.json schema, all fields,
  model options, prompt templating with contentTokens
- **`references/solution-structure.md`** — Solution manifest, SolutionStorage,
  .uipx format, project type registry
- **`references/tool-types.md`** — All 5 tool types with full JSON schemas,
  connection bindings, guardrails
- **`references/evaluation-framework.md`** — All evaluator types, eval set
  format, custom Python evaluators, coded-evals structure
- **`references/studio-web-api.md`** — Studio Web API endpoints for push/pull,
  publish, deploy, debug, file operations, resource builder
- **`references/cli-architecture.md`** — CLI plugin system, how to add new
  commands, configuration, authentication

### Example Files
- **`examples/low-code-agent/`** — Complete low-code agent with Web Search tool
- **`examples/coded-agent/`** — Complete coded Python agent with custom evaluator
- **`examples/solution/`** — Full solution wrapper with deployment resources

### Scripts
- **`scripts/solution_create.sh`** — Scaffold a new solution
- **`scripts/solution_pack.sh`** — Pack solution directory into .uis
- **`scripts/solution_unpack.sh`** — Unpack .uis into directory
- **`scripts/validate_agent.sh`** — Validate agent project structure
