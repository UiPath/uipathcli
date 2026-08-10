# Studio Web API Reference

Studio Web API endpoints for managing solutions, projects, deployments, and
resources. Base URL: `https://cloud.uipath.com/{orgId}/studio_/backend`

Swagger UI: `https://alpha.uipath.com/studioweb/studio_/backend/swagger/index.html`

## Authentication

All endpoints require Bearer token authentication. Obtain tokens via:
- UiPath CLI OAuth flow (`uipath config --auth login`)
- Client credentials (`uipath identity token create`)
- Personal Access Token (PAT)

## Solution Management

### Create Solution
```
POST /api/external/Solution
POST /api/Solution
```

### Get Solution
```
GET /api/external/Solution/{solutionId}
GET /api/Solution/{solutionId}
```

### Update Solution
```
POST /api/external/Solution/Update/{solutionId}
POST /api/Solution/Update/{solutionId}
```

### Delete Solution
```
DELETE /api/external/Solution/{solutionId}
DELETE /api/Solution/{solutionId}
```

### Search Solutions
```
GET /api/Solution/SearchSolutionsAndProjects
GET /api/Solution/OrganizationSolutionsAndProjects
GET /api/Solution/Name/{solutionName}/Ids
```

### Import Solution from ZIP
```
POST /api/Solution/Import
POST /api/Solution/AddSolution
POST /api/Solution/ImportAgentAsSolution
```

### Export Solution to ZIP
```
GET /api/Solution/{solutionId}/Export
GET /api/Solution/Studio/{solutionId}/Export
```

### Overwrite Solution from ZIP
```
POST /api/Solution/{solutionId}/Overwrite
```

## Project Management within Solution

### Add Project
```
POST /api/external/Solution/{solutionId}/Projects
POST /api/Solution/{solutionId}/Projects
```

### Delete Project
```
DELETE /api/external/Solution/{solutionId}/{projectId}
DELETE /api/Solution/{solutionId}/{projectId}
```

### Duplicate Solution
```
POST /api/Solution/{solutionId}/duplicate
```

## Publishing (Traditional)

### Create Publish Request
```
POST /api/external/Solution/{solutionId}/Publish-Requests
POST /api/Solution/{solutionId}/Publish-Requests
```

### Get Publish Request Status
```
GET /api/Solution/{solutionId}/Publish-Requests/{publishRequestId}
```

### Publish Specific Project
```
POST /api/Solution/{solutionId}/Project-Publish-Requests
GET /api/Solution/{solutionId}/Project-Publish-Requests/{publishRequestId}
```

### Get Publish Status
```
GET /api/external/Solution/{solutionId}/Publish-Status
GET /api/Solution/{solutionId}/Publish-Status
```

### Get Published Versions
```
GET /api/Solution/{solutionId}/Published-Versions
GET /api/Solution/{solutionId}/Next-Publish-Version
```

## ResourceBuilder (Maestro Deploy/Debug)

The ResourceBuilder API handles Maestro-style deployments with resource
management, overwrites, and test configuration.

### Deploy Solution
```
POST /api/resourcebuilder/solutions/{solutionKey}/deploy
```

Request body: `SolutionDeploymentRequest` with `packageVersionKey`,
`installationFolderKey`, `authenticationInfo`.

### Debug Solution
```
POST /api/resourcebuilder/solutions/{solutionKey}/debug
```

### Debug Individual Project
```
POST /api/resourcebuilder/solutions/{solutionKey}/projects/{projectKey}/debug
```

### Apply Test Configuration
```
POST /api/resourcebuilder/solutions/{solutionKey}/applyTestConfiguration
```

### Get Deployment Info
```
GET /api/resourcebuilder/solutions/{solutionKey}/deployment/entities
GET /api/resourcebuilder/solutions/{solutionKey}/deployment/resources
GET /api/resourcebuilder/solutions/{solutionKey}/deployment/resource-stats
```

### Debug Provisioning Status
```
GET /api/Solution/{solutionId}/Debug-Provisioning-Status
PATCH /api/Solution/{solutionId}/Publish-Requests/{publishRequestId}
```

### Resource Management
```
GET /api/resourcebuilder/solutions/{solutionKey}/resources/search
GET /api/resourcebuilder/solutions/{solutionKey}/resources/{resourceKey}
DELETE /api/resourcebuilder/solutions/{solutionKey}/resources/{resourceKey}
GET /api/resourcebuilder/solutions/{solutionKey}/resources/{resourceKey}/configuration
PATCH /api/resourcebuilder/solutions/{solutionKey}/resources/{resourceKey}/configuration
POST /api/resourcebuilder/solutions/{solutionKey}/resources/{resourceKey}/sync-configuration
POST /api/resourcebuilder/solutions/{solutionKey}/resources/reference
POST /api/resourcebuilder/solutions/{solutionKey}/resources/virtual
```

### Resource Overwrites
```
GET /api/resourcebuilder/solutions/{solutionKey}/overwrites
PATCH /api/resourcebuilder/solutions/{solutionKey}/overwrites
POST /api/resourcebuilder/solutions/{solutionKey}/overwrite
```

### Validate
```
POST /api/resourcebuilder/solutions/{solutionKey}/validate/definition
GET /api/resourcebuilder/solutions/{solutionKey}/validate
POST /api/resourcebuilder/{solutionKey}/validate-bindings
```

### Publish Locations
```
GET /api/resourcebuilder/solutions/publish-location
```

## Snapshots (Version Control)

### Create Snapshot
```
POST /api/Solution/{solutionId}/Snapshot
```

### List Snapshots
```
GET /api/Solution/{solutionId}/Snapshots
```

### Export Snapshot
```
GET /api/Solution/{solutionId}/Snapshot/{snapshotId}/Export
```

### Open Snapshot (Readonly)
```
GET /api/Solution/{solutionId}/Snapshot/{snapshotId}/Open
```

### Restore from Snapshot
```
POST /api/Solution/{solutionId}/Restore/{snapshotId}
```

### Get File from Snapshot
```
GET /api/Solution/{solutionId}/Snapshot/{snapshotId}/FileOperations/Structure
GET /api/Solution/{solutionId}/Snapshot/{snapshotId}/FileOperations/File/{fileId}
```

## File Operations

### Get Project File Structure
```
GET /api/Project/{projectId}/FileOperations/Structure
```

### Create File
```
POST /api/Project/{projectId}/FileOperations/File
```

### Get File Contents
```
GET /api/Project/{projectId}/FileOperations/File/{fileId}
```

### Update File
```
PUT /api/Project/{projectId}/FileOperations/File/{fileId}
```

### Rename File
```
POST /api/Project/{projectId}/FileOperations/File/Rename
```

### Get/Set Entry Points
```
GET /api/Project/{projectId}/FileOperations/EntryPoints
POST /api/Project/{projectId}/FileOperations/EntryPoints
```

### Set Main File
```
PUT /api/Project/{projectId}/FileOperations/SetMain/{fileId}
```

### Create Folder
```
POST /api/Project/{projectId}/FileOperations/Folder
```

### Move Folder
```
POST /api/Project/{projectId}/FileOperations/Folder/Move
```

### Delete File or Folder
```
DELETE /api/Project/{projectId}/FileOperations/Delete/{itemId}
```

## External Project API

### Export Project Version
```
GET /api/ExternalProject/export-version/{originalProjectId}/{version}
```
Exports in `.uip` format.

### List Published Versions
```
GET /api/ExternalProject/versions/{projectId}
```

### Import Project Version
```
POST /api/ExternalProject/import-version
```
Expects `.uip` archive.

### Create from Snapshot
```
POST /api/ExternalProject/create-from-snapshot
```

## Build & Package

### Create Build
```
POST /api/Build/{projectId}
```

### Get Build Version
```
GET /api/Build/{projectId}/Version/{buildVersion}
```

### Build Payload
```
POST /api/Build/{projectId}/BuildPayload
POST /api/Build/BuildPayload
GET /api/Build/BuildPayload/{buildPayloadId}
```

## Sharing

### Share Solution
```
POST /api/ShareSolution
DELETE /api/ShareSolution
GET /api/ShareSolution/SharedEntities
```

### Share Project
```
POST /api/ShareProject
DELETE /api/ShareProject
GET /api/ShareProject/Users
```

## Solution Locking

### Acquire/Release Lock
```
POST /api/Solution/{solutionId}/Lock/{lockKey}
DELETE /api/Solution/{solutionId}/Lock/{lockKey}
GET /api/Solution/{solutionId}/LockInfo/{lockKey}
GET /api/Solution/{solutionId}/AllLocks
```

### Per-Resource Lock
```
PUT /api/Solution/{solutionId}/PerResourceLock/{lockKey}
DELETE /api/Solution/{solutionId}/PerResourceLock/{lockKey}
POST /api/Solution/{solutionId}/PerResourceLock
```

## Sessions

### Allocate Robot Session
```
POST /api/Session
```

### Allocate Designer Session
```
POST /api/Session/Designer
```

## Templates

### Search Templates
```
POST /api/Template/SearchTemplates
GET /api/Template/GetSystemTemplates
```

### Create from Template
```
POST /api/Template/CreateProjectFromTemplate
```

### Create Template from Project
```
POST /api/Template/CreateTemplateFromProject
```

## Local Solution API

For interpreting and generating local solution files without server storage:

```
POST /api/LocalSolution           # Interpret local solution files
PUT  /api/LocalSolution           # Generate new local solution files
POST /api/LocalSolution/Project   # Add project to local solution
DELETE /api/LocalSolution/Project # Remove project from local solution
PUT  /api/LocalSolution/Project   # Create local project files
PUT  /api/LocalSolution/Workflow  # Generate workflow file content
```
