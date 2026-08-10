# Agent Structure Reference

Complete specification of all files in a UiPath agent project.

## agent.json — Core Agent Definition

The central file controlling agent behavior, model, prompts, and I/O.

### Low-Code Agent

```json
{
  "version": "1.1.0",
  "settings": {
    "model": "anthropic.claude-haiku-4-5-20251001-v1:0",
    "maxTokens": 16384,
    "temperature": 0,
    "engine": "basic-v2",
    "maxIterations": 25
  },
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string" }
    },
    "required": ["query"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "content": { "type": "string", "description": "Output content" }
    }
  },
  "metadata": {
    "storageVersion": "44.0.0",
    "isConversational": false,
    "showProjectCreationExperience": true,
    "targetRuntime": "pythonAgent"
  },
  "type": "lowCode",
  "projectId": "<uuid>",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant.",
      "contentTokens": [
        { "type": "simpleText", "rawString": "You are a helpful assistant." }
      ]
    },
    {
      "role": "user",
      "content": "query: {{query}}",
      "contentTokens": [
        { "type": "simpleText", "rawString": "query: " },
        { "type": "variable", "rawString": "input.query" },
        { "type": "simpleText", "rawString": "" }
      ]
    }
  ]
}
```

### Coded Agent

```json
{
  "version": "1.0.0",
  "metadata": {
    "storageVersion": "27.0.0",
    "targetRuntime": "python",
    "isConversational": false,
    "codeVersion": "1.0.10",
    "author": "user@example.com",
    "pushDate": "2025-10-24T20:00:07.198305+00:00"
  },
  "inputSchema": {
    "type": "object",
    "properties": {
      "query": { "type": "string" }
    },
    "required": ["query"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "result": { "type": "string" }
    },
    "required": ["result"]
  },
  "bindings": {
    "version": "2.0",
    "resources": []
  },
  "settings": {},
  "entryPoints": [{}],
  "type": "coded"
}
```

### Field Reference

| Field | Type | Description |
|-------|------|-------------|
| `version` | string | Schema version. `"1.1.0"` for low-code, `"1.0.0"` for coded |
| `type` | string | `"lowCode"` or `"coded"` |
| `settings.model` | string | LLM model identifier |
| `settings.maxTokens` | integer | Maximum output tokens (default: 16384) |
| `settings.temperature` | number | Sampling temperature (0-1, default: 0) |
| `settings.engine` | string | Agent engine version (default: `"basic-v2"`) |
| `settings.maxIterations` | integer | Maximum tool-use iterations (default: 25) |
| `inputSchema` | object | JSON Schema for agent inputs |
| `outputSchema` | object | JSON Schema for agent outputs |
| `messages` | array | System and user prompt messages (low-code only) |
| `metadata.storageVersion` | string | Internal storage version |
| `metadata.targetRuntime` | string | `"pythonAgent"` (low-code) or `"python"` (coded) |
| `metadata.isConversational` | boolean | Whether agent maintains conversation state |
| `metadata.codeVersion` | string | Code version (coded agents only) |
| `projectId` | string | UUID linking to project (low-code only) |
| `bindings` | object | Resource/connection bindings (coded agents) |

### Supported Models

| Model ID | Provider |
|----------|----------|
| `gpt-4o-2024-11-20` | OpenAI |
| `gpt-4.1-2025-04-14` | OpenAI |
| `anthropic.claude-haiku-4-5-20251001-v1:0` | Anthropic |
| `anthropic.claude-sonnet-4-20250514-v1:0` | Anthropic |

### Message Content Tokens

Messages use `contentTokens` for variable interpolation:

| Token Type | Description | Example |
|-----------|-------------|---------|
| `simpleText` | Static text | `{"type":"simpleText","rawString":"Hello "}` |
| `variable` | Input variable reference | `{"type":"variable","rawString":"input.query"}` |

The `content` field contains the rendered template with `{{variableName}}`
placeholders. The `contentTokens` array provides the structured representation.

### Input Schema Special Types

For file/attachment inputs, use the `job-attachment` type:

```json
{
  "type": "object",
  "properties": {
    "file": { "$ref": "#/definitions/job-attachment" }
  },
  "definitions": {
    "job-attachment": {
      "type": "object",
      "required": ["ID"],
      "x-uipath-resource-kind": "JobAttachment",
      "properties": {
        "ID": { "type": "string", "description": "Orchestrator attachment key" },
        "FullName": { "type": "string", "description": "File name" },
        "MimeType": { "type": "string", "description": "MIME type" },
        "Metadata": {
          "type": "object",
          "description": "Dictionary of metadata",
          "additionalProperties": { "type": "string" }
        }
      }
    }
  }
}
```

## entry-points.json

Defines the agent's callable entry points with input/output schemas.

```json
{
  "$schema": "https://cloud.uipath.com/draft/2024-12/entry-point",
  "$id": "entry-points.json",
  "entryPoints": [
    {
      "filePath": "main.py",
      "uniqueId": "<uuid>",
      "type": "agent",
      "input": {
        "type": "object",
        "properties": {
          "query": { "type": "string" }
        },
        "required": ["query"]
      },
      "output": {
        "type": "object",
        "properties": {
          "result": { "type": "string" }
        },
        "required": ["result"]
      }
    }
  ]
}
```

For low-code agents, `filePath` is omitted. For coded agents, it points to the
Python entry point (e.g., `main.py`).

## project.uiproj

Minimal project descriptor:

```json
{
  "ProjectType": "Agent",
  "Name": "Agent",
  "Description": null,
  "MainFile": null
}
```

## .agent-builder/ (Low-Code Only)

### agent.json
Extended version of the top-level `agent.json` that includes full resource
definitions inline (with `inputSchema`, `outputSchema`, connection properties).
Contains `id` field matching `projectId`.

### bindings.json
Connection bindings for external tools:

```json
{
  "version": "2.0",
  "resources": [
    {
      "resource": "connection",
      "key": "<connection-uuid>",
      "value": {
        "connectionId": {
          "defaultValue": "<connection-uuid>",
          "isExpression": false,
          "displayName": "Connection ID"
        }
      },
      "metadata": {
        "connector": "uipath-uipath-airdk",
        "useConnectionService": "true",
        "bindingsVersion": "2.2",
        "solutionsSupport": "true"
      }
    }
  ]
}
```

### entry-points.json
Same structure as top-level `entry-points.json`.

## source_code/ (Coded Agents Only)

### main.py

```python
import logging
from pydantic.dataclasses import dataclass
from uipath.tracing import traced

logger = logging.getLogger(__name__)

@dataclass
class AgentInput:
    query: str

@dataclass
class AgentOutput:
    result: str

@traced()
async def main(input: AgentInput) -> AgentOutput:
    # Agent logic here
    return AgentOutput(result=f"Processed: {input.query}")
```

Key patterns:
- Use `@dataclass` from `pydantic.dataclasses` for I/O types
- Use `@traced()` decorator for observability
- The `main` function is the entry point — must be async
- Use `@mockable(example_calls=...)` for tool simulation in evals

### pyproject.toml

```toml
[project]
name = "my-agent"
version = "0.0.1"
description = "My agent description"
authors = [{ name = "Author", email = "author@example.com" }]
dependencies = ["uipath>=2.1.87"]
requires-python = ">=3.10"
```

### uipath.json

Runtime configuration matching the entry-points.json schema:

```json
{
  "entryPoints": [
    {
      "filePath": "main.py",
      "uniqueId": "<uuid>",
      "type": "agent",
      "input": { ... },
      "output": { ... }
    }
  ],
  "bindings": {
    "version": "2.0",
    "resources": []
  }
}
```

## flow-layout.json (Low-Code Only)

Visual layout metadata for the agent builder UI. Auto-generated.
Typically an empty object `{}` or contains node positioning data.

## .project/JitCustomTypes.json

JIT compilation custom types. Usually empty: `{}`
