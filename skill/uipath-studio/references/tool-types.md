# Tool Types Reference

Complete schemas for all agent tool/resource types in UiPath.

Tools are added to an agent by creating a `resource.json` file in
`Agent/resources/<ToolName>/resource.json`.

For low-code agents, tools are also registered inline in the
`.agent-builder/agent.json` under the `resources` array.

## 1. Integration Tool (External)

External tools that call APIs through UiPath connectors.

### Web Search Example

```json
{
  "$resourceType": "tool",
  "name": "Web Search",
  "description": "Web search executes a search of the public domain using a natural language search query.",
  "location": "external",
  "type": "integration",
  "inputSchema": {
    "type": "object",
    "properties": {
      "provider": {
        "type": "string",
        "title": "Search Engine",
        "enum": ["GoogleCustomSearch"],
        "oneOf": [{ "const": "GoogleCustomSearch", "title": "GoogleCustomSearch" }]
      },
      "query": {
        "type": "string",
        "title": "Search",
        "description": "The natural language query to search the web for"
      },
      "num": {
        "type": "integer",
        "title": "Number of results",
        "description": "The number of results. Default to 10."
      }
    },
    "required": ["provider", "query"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "results[*]": {
        "type": "array",
        "title": "Results",
        "items": { "$ref": "#/definitions/results[*]" }
      },
      "formattedResults": {
        "type": "string",
        "title": "Formatted results"
      }
    },
    "definitions": {
      "results[*]": {
        "type": "object",
        "properties": {
          "title": { "type": "string" },
          "snippet": { "type": "string" },
          "url": { "type": "string" }
        }
      }
    }
  },
  "settings": {},
  "properties": {
    "toolPath": "/v2/webSearch",
    "objectName": "v2::webSearch",
    "toolDisplayName": "Web Search",
    "method": "POST",
    "connection": {
      "id": "<connection-uuid>",
      "name": "UiPath GenAI Activities",
      "state": "enabled",
      "connector": {
        "key": "uipath-uipath-airdk",
        "name": "UiPath GenAI Activities",
        "enabled": true
      },
      "folder": {
        "key": "<folder-uuid>",
        "path": "<folder-uuid>"
      },
      "solutionProperties": {
        "resourceKey": "<connection-resource-uuid>"
      }
    },
    "parameters": [
      {
        "name": "provider",
        "displayName": "Search Engine",
        "type": "string",
        "fieldLocation": "body",
        "fieldVariant": "static",
        "value": "GoogleCustomSearch",
        "dynamic": false,
        "position": "primary",
        "sortOrder": 1,
        "required": true
      },
      {
        "name": "query",
        "displayName": "Search",
        "type": "string",
        "fieldLocation": "body",
        "fieldVariant": "dynamic",
        "value": "{{prompt}}",
        "dynamic": true,
        "position": "primary",
        "sortOrder": 2,
        "required": true
      }
    ],
    "bodyStructure": { "contentType": "json" }
  },
  "guardrail": { "policies": [] },
  "id": "<tool-uuid>",
  "isPreview": false,
  "isEnabled": true
}
```

### Common Integration Tools

| Tool | toolPath | Connector |
|------|----------|-----------|
| Web Search | `/v2/webSearch` | `uipath-uipath-airdk` |
| Web Reader | `/v1/webRead` | `uipath-uipath-airdk` |
| Web Summary | `/v1/webSummary` | `uipath-uipath-airdk` |

### Parameter Field Variants

| Variant | Description |
|---------|-------------|
| `static` | Fixed value, not changeable by LLM |
| `dynamic` | LLM provides the value, `{{prompt}}` in template |

## 2. Agent Tool (Solution)

References another agent within the same solution.

```json
{
  "$resourceType": "tool",
  "id": "<tool-uuid>",
  "referenceKey": "<target-agent-process-uuid>",
  "name": "Agent 2",
  "type": "agent",
  "description": "Used to add things",
  "location": "solution",
  "isEnabled": true,
  "inputSchema": {
    "type": "object",
    "properties": {
      "number1": { "type": "number", "description": "The first number" },
      "number2": { "type": "number", "description": "The second number" }
    },
    "required": ["number1", "number2"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "sum": { "type": "number", "description": "The sum" }
    }
  },
  "settings": {},
  "guardrail": { "policies": [] },
  "argumentProperties": {},
  "properties": {
    "processName": "Agent 2",
    "folderPath": "solution_folder"
  }
}
```

The `referenceKey` links to the target agent's process resource key in
`resources/solution_folder/process/agent/<Name>.json`.

## 3. Internal Tool (Built-in)

Built-in UiPath tools that do not require external connections.

### Analyze Files Tool

```json
{
  "$resourceType": "tool",
  "referenceKey": null,
  "name": "Analyze Files",
  "type": "internal",
  "description": "Analyze one or more files with an LLM to extract, synthesize, or answer queries about their content.",
  "isEnabled": true,
  "inputSchema": {
    "type": "object",
    "properties": {
      "attachments": {
        "type": "array",
        "items": { "$ref": "#/definitions/job-attachment" },
        "description": "Array of files to process"
      },
      "analysisTask": {
        "type": "string",
        "description": "The task or question for processing the files"
      }
    },
    "required": ["attachments", "analysisTask"],
    "definitions": {
      "job-attachment": {
        "type": "object",
        "properties": {
          "ID": { "type": "string", "description": "Orchestrator attachment key" },
          "FullName": { "type": "string", "description": "File name" },
          "MimeType": { "type": "string", "description": "MIME type" },
          "Metadata": {
            "type": "object",
            "additionalProperties": { "type": "string" }
          }
        },
        "required": ["ID"],
        "x-uipath-resource-kind": "JobAttachment"
      }
    }
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "analysis": {
        "type": "string",
        "description": "Analysis result"
      }
    },
    "required": ["analysis"]
  },
  "settings": {},
  "guardrail": { "policies": [] },
  "argumentProperties": {},
  "properties": {
    "toolType": "analyze-attachments"
  },
  "id": "<tool-uuid>"
}
```

## 4. Context Resource (RAG/Index)

Provides semantic search over indexed documents.

```json
{
  "$resourceType": "context",
  "name": "MyKnowledgeBase",
  "description": "Semantic search over support tickets",
  "folderPath": "Solution Folder",
  "indexName": "MyKnowledgeBase",
  "id": "<resource-uuid>",
  "referenceKey": null,
  "settings": {
    "query": {
      "description": "The query for the Semantic strategy.",
      "variant": "dynamic"
    },
    "folderPathPrefix": { "variant": "static" },
    "threshold": 0.5,
    "resultCount": 3,
    "retrievalMode": "semantic",
    "fileExtension": { "value": "All" }
  }
}
```

### Context Settings

| Field | Type | Description |
|-------|------|-------------|
| `threshold` | number | Similarity threshold (0-1) |
| `resultCount` | integer | Number of results to return |
| `retrievalMode` | string | `"semantic"`, `"keyword"`, or `"hybrid"` |
| `fileExtension` | object | File type filter |

## 5. Escalation (HITL — Human-in-the-Loop)

Routes to Action Center for human approval/review.

```json
{
  "$resourceType": "escalation",
  "id": "<escalation-uuid>",
  "name": "AskConfirmation",
  "description": "",
  "channels": [
    {
      "id": "<channel-uuid>",
      "name": "Channel",
      "description": "Channel description",
      "inputSchema": {
        "type": "object",
        "properties": {
          "Content": { "type": "string" },
          "Comment": { "type": "string", "description": "User comments" }
        }
      },
      "outputSchema": {
        "type": "object",
        "properties": {
          "Comment": { "type": "string", "description": "User comments" }
        }
      },
      "outcomeMapping": {
        "approve": "continue",
        "reject": "continue"
      },
      "recipients": [
        {
          "type": 1,
          "value": "<user-uuid>",
          "displayName": "Reviewer Name"
        }
      ],
      "type": "actionCenter",
      "properties": {
        "appName": "SimpleApprovalApp",
        "appVersion": 1,
        "resourceKey": "<app-resource-uuid>",
        "isActionableMessageEnabled": true,
        "actionableMessageMetaData": {
          "fieldSet": {
            "type": "fieldSet",
            "id": "<uuid>",
            "fields": [
              { "id": "Content", "name": "Content", "type": "Fact" },
              { "id": "Comment", "name": "Comment", "type": "Input.Text" }
            ]
          },
          "actionSet": {
            "type": "actionSet",
            "id": "<uuid>",
            "actions": [
              { "id": "approve", "name": "approve", "title": "approve", "type": "Action.Http", "isPrimary": true },
              { "id": "reject", "name": "reject", "title": "reject", "type": "Action.Http", "isPrimary": true }
            ]
          }
        }
      }
    }
  ],
  "isAgentMemoryEnabled": false,
  "governanceProperties": { "isEscalatedAtRuntime": false },
  "escalationType": 0,
  "properties": {}
}
```

### Outcome Mapping Options

| Outcome | Maps To | Description |
|---------|---------|-------------|
| `approve` | `continue` | Resume agent execution |
| `reject` | `continue` | Resume (agent handles rejection) |
| `approve` | `stop` | Stop agent on approval |
| `reject` | `stop` | Stop agent on rejection |

### Recipient Types

| Type | Description |
|------|-------------|
| `1` | Specific user (by UUID) |
| `2` | Group |
| `3` | Dynamic (resolved at runtime) |

## Guardrails

All tool types support guardrail policies:

```json
{
  "guardrail": {
    "policies": [
      {
        "name": "content-filter",
        "enabled": true,
        "config": { ... }
      }
    ]
  }
}
```

Currently, an empty `policies` array is the default.
