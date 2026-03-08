# Solution Structure Reference

Complete specification of UiPath Solution packaging and deployment resources.

## .uis File Format

A `.uis` file is a standard ZIP archive containing the entire solution. It can
be created by zipping the solution directory and renaming to `.uis`.

## SolutionStorage.json

Maps project IDs to their relative paths within the solution:

```json
{
  "SolutionId": "<uuid>",
  "Projects": [
    {
      "ProjectId": "<uuid>",
      "ProjectRelativePath": "Agent/project.uiproj"
    },
    {
      "ProjectId": "<uuid>",
      "ProjectRelativePath": "RPA Workflow/project.uiproj"
    }
  ]
}
```

## Solution Manifest (.uipx)

The `<SolutionName>.uipx` file is a JSON manifest declaring all projects:

```json
{
  "DocVersion": "1.0.0",
  "StudioMinVersion": "2025.04.0",
  "SolutionId": "<uuid>",
  "Projects": [
    {
      "Type": "Agent",
      "ProjectRelativePath": "Agent/project.uiproj",
      "Id": "<uuid>"
    },
    {
      "Type": "Process",
      "ProjectRelativePath": "RPA Workflow/project.uiproj",
      "Id": "<uuid>"
    }
  ]
}
```

### Supported Project Types

| Type | Description | Key Files |
|------|-------------|-----------|
| `Agent` | AI agent (low-code or coded) | `agent.json`, `entry-points.json` |
| `Process` | RPA workflow | `Main.xaml`, `project.json` |
| `WebApp` | Web application (for HITL UIs) | `.app/` directory, `Main.xaml` |
| `CaseManagement` | Case management flow | `case.stage.json`, `.bpmn` |
| `BusinessRules` | DMN business rules | `*.dmn` |
| `Connector` | Custom connector | connector definitions |
| `ProcessOrchestration` | BPMN process orchestration | `Process.bpmn` |
| `Api` | API workflow (serverless) | `Workflow.json` |

## Deployment Resources

The `resources/solution_folder/` directory contains deployment descriptors.

### Package Resource

`resources/solution_folder/package/<Name>.json`:

```json
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "Agent",
    "kind": "package",
    "apiVersion": "orchestrator.uipath.com/v1",
    "projectKey": "<project-uuid>",
    "dependencies": [],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{ "fullyQualifiedName": "solution_folder" }],
    "spec": {
      "fileName": null,
      "fileReference": null,
      "name": "Agent",
      "description": null
    },
    "locks": [],
    "key": "<package-uuid>"
  }
}
```

### Process Resource

`resources/solution_folder/process/agent/<Name>.json`:

```json
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "Agent",
    "kind": "process",
    "type": "agent",
    "apiVersion": "orchestrator.uipath.com/v1",
    "projectKey": "<project-uuid>",
    "dependencies": [
      { "name": "Agent", "kind": "package" }
    ],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{ "fullyQualifiedName": "solution_folder" }],
    "spec": {
      "entryPointUniqueId": null,
      "type": "Agent",
      "name": "Agent",
      "description": null,
      "package": { "key": "<package-uuid>" },
      "packageName": "<SolutionName>.agent.Agent",
      "inputArguments": "{}",
      "hiddenForAttendedUser": false,
      "alwaysRunning": false,
      "autoStartProcess": false,
      "targetFrameworkValue": "Portable",
      "agentMemory": false,
      "retentionAction": "Delete",
      "retentionPeriod": 30,
      "staleRetentionAction": "Delete",
      "staleRetentionPeriod": 180,
      "tags": []
    },
    "locks": [],
    "key": "<process-uuid>"
  }
}
```

### Process Types by Project Type

| Project Type | Process `type` | Process `kind` |
|-------------|----------------|----------------|
| Agent | `agent` | `process` |
| Process/RPA | — | `process` |
| WebApp | — | `process` (under `webApp/`) |
| CaseManagement | — | `process` (under `caseManagement/`) |
| ProcessOrchestration | — | `process` (under `processOrchestration/`) |
| Api | — | `process` (under `api/`) |

### Connection Resource

`resources/solution_folder/connection/<connector-key>/<name>.json`:

```json
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "UiPath GenAI Activities",
    "kind": "connection",
    "apiVersion": "elements.uipath.com/v1",
    "dependencies": [],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{ "fullyQualifiedName": "solution_folder" }],
    "spec": {
      "connectorKey": "uipath-uipath-airdk",
      "connectorVersion": null,
      "name": "UiPath GenAI Activities"
    },
    "locks": [],
    "key": "<connection-uuid>"
  }
}
```

### Index Resource

`resources/solution_folder/index/<Name>.json`:

```json
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "MyIndex",
    "kind": "index",
    "apiVersion": "ecs.uipath.com/v1",
    "dependencies": [],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{ "fullyQualifiedName": "solution_folder" }],
    "spec": {
      "name": "MyIndex",
      "description": "Semantic search index",
      "indexConfigurationJson": "{...}"
    },
    "locks": [],
    "key": "<index-uuid>"
  }
}
```

The `indexConfigurationJson` is a JSON string containing:

```json
{
  "Version": 2,
  "Provider": 3,
  "DataSource": {
    "Type": 1,
    "Properties": {
      "folderName": "Shared",
      "directoryPath": "/",
      "storageBucketName": "MyIndex",
      "storageBucketId": "00000000-0000-0000-0000-000000000000",
      "fileNameGlob": "*"
    }
  },
  "EmbeddingModel": "text-embedding-3-large",
  "ExtractionStrategy": null,
  "UserFields": []
}
```

### App Version Resource

`resources/solution_folder/appVersion/<AppName>.json`:

```json
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "MyApp",
    "kind": "appVersion",
    "apiVersion": "apps.uipath.com/v2",
    "projectKey": "<project-uuid>",
    "dependencies": [],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{ "fullyQualifiedName": "solution_folder" }],
    "spec": {
      "name": "MyApp",
      "description": null,
      "isAppPublic": false
    },
    "locks": [],
    "key": "<appversion-uuid>"
  }
}
```

## Multi-Project Solution Example (EverythingBagel)

A solution with all project types:

```json
{
  "DocVersion": "1.0.0",
  "StudioMinVersion": "2025.04.0",
  "SolutionId": "<uuid>",
  "Projects": [
    { "Type": "BusinessRules", "ProjectRelativePath": "Business Rules/project.uiproj", "Id": "<uuid>" },
    { "Type": "WebApp", "ProjectRelativePath": "SimpleApprovalApp/project.uiproj", "Id": "<uuid>" },
    { "Type": "CaseManagement", "ProjectRelativePath": "Agentic case/project.uiproj", "Id": "<uuid>" },
    { "Type": "Process", "ProjectRelativePath": "RPA Workflow/project.uiproj", "Id": "<uuid>" },
    { "Type": "Connector", "ProjectRelativePath": "Connector/project.uiproj", "Id": "<uuid>" },
    { "Type": "ProcessOrchestration", "ProjectRelativePath": "Agentic Process/project.uiproj", "Id": "<uuid>" },
    { "Type": "Agent", "ProjectRelativePath": "Agent/project.uiproj", "Id": "<uuid>" },
    { "Type": "Api", "ProjectRelativePath": "API Workflow/project.uiproj", "Id": "<uuid>" }
  ]
}
```

## Project Type Key Files

### ProcessOrchestration — Process.bpmn

BPMN 2.0 XML with UiPath extensions:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<bpmn:definitions xmlns:bpmn="http://www.omg.org/spec/BPMN/20100524/MODEL"
                  xmlns:uipath="http://uipath.org/schema/bpmn">
  <bpmn:process id="Process_1" name="Process" isExecutable="false">
    <bpmn:startEvent id="Event_start" name="Start event">
      <bpmn:extensionElements>
        <uipath:entryPointId value="<uuid>" />
      </bpmn:extensionElements>
    </bpmn:startEvent>
  </bpmn:process>
</bpmn:definitions>
```

### CaseManagement — case.stage.json

```json
{
  "root": {
    "id": "root",
    "type": "case-management:root",
    "name": "My Case",
    "caseIdentifierType": "constant",
    "caseIdentifier": "CASE",
    "caseAppEnabled": true
  },
  "nodes": [
    {
      "id": "trigger_1",
      "type": "case-management:Trigger",
      "position": { "x": 160, "y": 198.5 },
      "data": { "parentElement": { "id": "root", "type": "case-management:root" } }
    },
    {
      "id": "stage_1",
      "type": "case-management:Stage",
      "position": { "x": 326, "y": 200 },
      "data": {
        "label": "Stage 1",
        "parentElement": { "id": "root", "type": "case-management:root" },
        "tasks": []
      }
    }
  ],
  "edges": [
    {
      "id": "edge_initial",
      "source": "trigger_1",
      "target": "stage_1",
      "type": "case-management:TriggerEdge"
    }
  ]
}
```

### Api — Workflow.json

ServerlessV2 DSL:

```json
{
  "document": {
    "dsl": "1.0.0",
    "name": "Workflow",
    "version": "0.0.1",
    "namespace": "default",
    "metadata": { "variables": [] }
  },
  "do": [
    {
      "Sequence_1": {
        "do": [],
        "metadata": { "fullName": "Sequence", "activityType": "Sequence" }
      }
    }
  ],
  "evaluate": { "mode": "strict", "language": "javascript" }
}
```
