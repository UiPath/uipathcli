#!/usr/bin/env bash
# Creates a new UiPath Solution scaffold with the specified project type.
#
# Usage: solution_create.sh <solution_name> <project_type> [project_name]
#
# Project types: Agent, Process, WebApp, CaseManagement, BusinessRules,
#                Connector, ProcessOrchestration, Api
#
# Example: solution_create.sh "MyAgent" "Agent"
#          solution_create.sh "MyWorkflow" "Agent" "ResearchBot"

set -euo pipefail

SOLUTION_NAME="${1:?Usage: solution_create.sh <solution_name> <project_type> [project_name]}"
PROJECT_TYPE="${2:?Usage: solution_create.sh <solution_name> <project_type> [project_name]}"
PROJECT_NAME="${3:-Agent}"

# Generate UUIDs
gen_uuid() {
    python3 -c "import uuid; print(str(uuid.uuid4()))" 2>/dev/null \
        || cat /proc/sys/kernel/random/uuid 2>/dev/null \
        || uuidgen 2>/dev/null \
        || echo "$(od -x /dev/urandom | head -1 | awk '{OFS="-"; print $2$3,$4,$5,$6,$7$8$9}')"
}

SOLUTION_ID=$(gen_uuid)
PROJECT_ID=$(gen_uuid)
PROJECT_UUID=$(gen_uuid)
PACKAGE_KEY=$(gen_uuid)
PROCESS_KEY=$(gen_uuid)

# Create directory structure
mkdir -p "${SOLUTION_NAME}/${PROJECT_NAME}"
mkdir -p "${SOLUTION_NAME}/resources/solution_folder/package"
mkdir -p "${SOLUTION_NAME}/resources/solution_folder/process/agent"

# SolutionStorage.json
cat > "${SOLUTION_NAME}/SolutionStorage.json" << EOF
{"SolutionId":"${SOLUTION_ID}","Projects":[{"ProjectId":"${PROJECT_ID}","ProjectRelativePath":"${PROJECT_NAME}/project.uiproj"}]}
EOF

# Solution manifest (.uipx)
cat > "${SOLUTION_NAME}/${SOLUTION_NAME}.uipx" << EOF
{
  "DocVersion": "1.0.0",
  "StudioMinVersion": "2025.04.0",
  "SolutionId": "${SOLUTION_ID}",
  "Projects": [
    {
      "Type": "${PROJECT_TYPE}",
      "ProjectRelativePath": "${PROJECT_NAME}/project.uiproj",
      "Id": "${PROJECT_UUID}"
    }
  ]
}
EOF

# project.uiproj
cat > "${SOLUTION_NAME}/${PROJECT_NAME}/project.uiproj" << EOF
{
  "ProjectType": "${PROJECT_TYPE}",
  "Name": "${PROJECT_NAME}",
  "Description": null,
  "MainFile": null
}
EOF

# Package resource
cat > "${SOLUTION_NAME}/resources/solution_folder/package/${PROJECT_NAME}.json" << EOF
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "${PROJECT_NAME}",
    "kind": "package",
    "apiVersion": "orchestrator.uipath.com/v1",
    "projectKey": "${PROJECT_UUID}",
    "dependencies": [],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{"fullyQualifiedName": "solution_folder"}],
    "spec": {
      "fileName": null,
      "fileReference": null,
      "name": "${PROJECT_NAME}",
      "description": null
    },
    "locks": [],
    "key": "${PACKAGE_KEY}"
  }
}
EOF

# Process resource
cat > "${SOLUTION_NAME}/resources/solution_folder/process/agent/${PROJECT_NAME}.json" << EOF
{
  "docVersion": "1.0.0",
  "resource": {
    "name": "${PROJECT_NAME}",
    "kind": "process",
    "type": "agent",
    "apiVersion": "orchestrator.uipath.com/v1",
    "projectKey": "${PROJECT_UUID}",
    "dependencies": [{"name": "${PROJECT_NAME}", "kind": "package"}],
    "runtimeDependencies": [],
    "files": [],
    "folders": [{"fullyQualifiedName": "solution_folder"}],
    "spec": {
      "entryPointUniqueId": null,
      "type": "Agent",
      "name": "${PROJECT_NAME}",
      "description": null,
      "package": {"key": "${PACKAGE_KEY}"},
      "packageName": "${SOLUTION_NAME}.agent.${PROJECT_NAME}",
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
    "key": "${PROCESS_KEY}"
  }
}
EOF

# Create Agent-specific files if project type is Agent
if [ "${PROJECT_TYPE}" = "Agent" ]; then
    ENTRY_POINT_ID=$(gen_uuid)

    mkdir -p "${SOLUTION_NAME}/${PROJECT_NAME}/.agent-builder"
    mkdir -p "${SOLUTION_NAME}/${PROJECT_NAME}/.project"
    mkdir -p "${SOLUTION_NAME}/${PROJECT_NAME}/evals/eval-sets"
    mkdir -p "${SOLUTION_NAME}/${PROJECT_NAME}/evals/evaluators"

    # agent.json
    cat > "${SOLUTION_NAME}/${PROJECT_NAME}/agent.json" << EOF
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
      "query": {"type": "string"}
    },
    "required": ["query"]
  },
  "outputSchema": {
    "type": "object",
    "properties": {
      "content": {"type": "string", "description": "Output content"}
    }
  },
  "metadata": {
    "storageVersion": "44.0.0",
    "isConversational": false,
    "showProjectCreationExperience": true,
    "targetRuntime": "pythonAgent"
  },
  "type": "lowCode",
  "projectId": "${PROJECT_ID}",
  "messages": [
    {
      "role": "system",
      "content": "You are a helpful assistant.",
      "contentTokens": [{"type": "simpleText", "rawString": "You are a helpful assistant."}]
    },
    {
      "role": "user",
      "content": "query: {{query}}",
      "contentTokens": [
        {"type": "simpleText", "rawString": "query: "},
        {"type": "variable", "rawString": "input.query"},
        {"type": "simpleText", "rawString": ""}
      ]
    }
  ]
}
EOF

    # entry-points.json
    cat > "${SOLUTION_NAME}/${PROJECT_NAME}/entry-points.json" << EOF
{
  "\$schema": "https://cloud.uipath.com/draft/2024-12/entry-point",
  "\$id": "entry-points.json",
  "entryPoints": [
    {
      "uniqueId": "${ENTRY_POINT_ID}",
      "type": "agent",
      "input": {
        "type": "object",
        "properties": {"query": {"type": "string"}},
        "required": ["query"]
      },
      "output": {
        "type": "object",
        "properties": {"content": {"type": "string", "description": "Output content"}}
      }
    }
  ]
}
EOF

    # flow-layout.json
    echo '{}' > "${SOLUTION_NAME}/${PROJECT_NAME}/flow-layout.json"

    # .project/JitCustomTypes.json
    echo '{}' > "${SOLUTION_NAME}/${PROJECT_NAME}/.project/JitCustomTypes.json"

    # .agent-builder files
    cat > "${SOLUTION_NAME}/${PROJECT_NAME}/.agent-builder/bindings.json" << EOF
{"version": "2.0", "resources": []}
EOF
    cp "${SOLUTION_NAME}/${PROJECT_NAME}/entry-points.json" \
       "${SOLUTION_NAME}/${PROJECT_NAME}/.agent-builder/entry-points.json"

    # Default evaluators
    EVAL_SET_ID=$(gen_uuid)
    EVAL_ID=$(gen_uuid)
    EVALUATOR_ID=$(gen_uuid)
    TRAJECTORY_ID=$(gen_uuid)

    cat > "${SOLUTION_NAME}/${PROJECT_NAME}/evals/eval-sets/evaluation-set-default.json" << EOF
{
  "fileName": "evaluation-set-default.json",
  "id": "${EVAL_SET_ID}",
  "name": "Default Evaluation Set",
  "batchSize": 10,
  "evaluatorRefs": ["${EVALUATOR_ID}"],
  "evaluations": []
}
EOF

    cat > "${SOLUTION_NAME}/${PROJECT_NAME}/evals/evaluators/evaluator-default.json" << EOF
{
  "version": "1.0",
  "id": "${EVALUATOR_ID}",
  "description": "Uses an LLM to judge semantic similarity.",
  "evaluatorTypeId": "uipath-llm-judge-output-semantic-similarity",
  "evaluatorConfig": {
    "name": "SemanticSimilarityEvaluator",
    "targetOutputKey": "*",
    "model": "gpt-4.1-2025-04-14",
    "prompt": "Compare the outputs and evaluate semantic similarity.\n\nActual: {{ActualOutput}}\nExpected: {{ExpectedOutput}}\n\nScore 0-100.",
    "temperature": 0.0,
    "defaultEvaluationCriteria": {"expectedOutput": {"content": ""}}
  }
}
EOF

    cat > "${SOLUTION_NAME}/${PROJECT_NAME}/evals/evaluators/evaluator-default-trajectory.json" << EOF
{
  "version": "1.0",
  "id": "${TRAJECTORY_ID}",
  "description": "Evaluates agent execution trajectory.",
  "evaluatorTypeId": "uipath-llm-judge-trajectory-similarity",
  "evaluatorConfig": {
    "name": "TrajectoryEvaluator",
    "model": "gpt-4.1-2025-04-14",
    "prompt": "Evaluate trajectory.\n\nExpected: {{ExpectedAgentBehavior}}\nHistory: {{AgentRunHistory}}\n\nScore 0-100.",
    "temperature": 0.0,
    "defaultEvaluationCriteria": {"expectedAgentBehavior": "The agent should correctly perform the task."}
  }
}
EOF
fi

echo "Solution '${SOLUTION_NAME}' created successfully with ${PROJECT_TYPE} project '${PROJECT_NAME}'"
echo "Directory: ${SOLUTION_NAME}/"
find "${SOLUTION_NAME}" -type f | sort | sed 's/^/  /'
