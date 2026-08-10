# Evaluation Framework Reference

Complete specification of UiPath agent evaluation system.

## Directory Structure

### Low-Code Agents

```
Agent/evals/
  eval-sets/
    evaluation-set-default.json
  evaluators/
    evaluator-default.json
    evaluator-default-trajectory.json
```

### Coded Agents

```
Agent/coded-evals/
  eval-sets/
    default.json
  evaluators/
    exact-match.json
    contains.json
    json-similarity.json
    llm-judge-semantic-similarity.json
    llm-judge-strict-json-similarity.json
    trajectory.json
    custom/
      my_evaluator.py
      types/
        my-evaluator-types.json
Agent/evals/
  eval-sets/
    evaluation-set-default.json
  evaluators/
    evaluator-default.json
    evaluator-default-trajectory.json
```

Coded agents have BOTH `coded-evals/` (Python SDK format) and `evals/`
(platform format). The `coded-evals/` directory uses the Python evaluator SDK.

## Evaluation Sets

### Low-Code Format

`evals/eval-sets/evaluation-set-default.json`:

```json
{
  "fileName": "evaluation-set-default.json",
  "id": "<uuid>",
  "name": "Default Evaluation Set",
  "batchSize": 10,
  "evaluatorRefs": ["<evaluator-id>"],
  "evaluations": [
    {
      "id": "<uuid>",
      "name": "Test Case Name",
      "inputs": {
        "query": "What is UiPath?"
      },
      "expectedOutput": {
        "content": "UiPath is a leading enterprise automation platform..."
      },
      "simulationInstructions": "",
      "expectedAgentBehavior": "The agent should search the web and provide a summary.",
      "simulateInput": false,
      "inputGenerationInstructions": "",
      "simulateTools": false,
      "toolsToSimulate": [],
      "evalSetId": "<eval-set-uuid>",
      "createdAt": "2026-01-22T18:47:25.622Z",
      "updatedAt": "2026-01-22T18:50:44.767Z",
      "source": "manual"
    }
  ]
}
```

### Coded Format

`coded-evals/eval-sets/default.json`:

```json
{
  "version": "1.0",
  "id": "MyEvalSet",
  "name": "My Evaluation Set",
  "evaluatorRefs": [
    "ContainsEvaluator",
    "ExactMatchEvaluator",
    "LLMJudgeOutputEvaluator",
    "TrajectoryEvaluator"
  ],
  "evaluations": [
    {
      "id": "test-1",
      "name": "Addition Test",
      "inputs": {
        "a": 1,
        "b": 4,
        "operator": "+"
      },
      "evaluationCriterias": {
        "ContainsEvaluator": { "searchText": "5" },
        "ExactMatchEvaluator": {
          "expectedOutput": { "result": "5.0" }
        },
        "LLMJudgeOutputEvaluator": {
          "expectedOutput": { "result": 5.0 }
        },
        "TrajectoryEvaluator": {
          "expectedAgentBehavior": "The agent should correctly add 1 + 4."
        }
      },
      "updatedAt": "2025-10-29T18:22:31.492Z"
    }
  ],
  "fileName": "default.json",
  "updatedAt": "2025-10-29T18:22:31.492Z"
}
```

### Evaluation Fields

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Unique evaluation ID |
| `name` | string | Human-readable test name |
| `inputs` | object | Input values matching agent inputSchema |
| `expectedOutput` | object | Expected output (low-code format) |
| `evaluationCriterias` | object | Per-evaluator criteria (coded format) |
| `simulationInstructions` | string | Instructions for input simulation |
| `expectedAgentBehavior` | string | Description of expected agent trajectory |
| `simulateInput` | boolean | Whether to generate synthetic inputs |
| `simulateTools` | boolean | Whether to simulate tool responses |
| `toolsToSimulate` | array | Specific tools to simulate |

## Built-in Evaluator Types

### 1. Exact Match (`uipath-exact-match`)

Checks if output exactly matches expected value.

```json
{
  "version": "1.0",
  "id": "ExactMatchEvaluator",
  "description": "Checks if the response text exactly matches the expected value.",
  "evaluatorTypeId": "uipath-exact-match",
  "evaluatorConfig": {
    "name": "ExactMatchEvaluator",
    "targetOutputKey": "result",
    "negated": false,
    "ignoreCase": false,
    "defaultEvaluationCriteria": {
      "expectedOutput": { "result": "5.0" }
    }
  }
}
```

### 2. Contains (`uipath-contains`)

Checks if output contains a substring.

```json
{
  "version": "1.0",
  "id": "ContainsEvaluator",
  "description": "Checks if the response text includes the expected text.",
  "evaluatorTypeId": "uipath-contains",
  "evaluatorConfig": {
    "name": "ContainsEvaluator",
    "targetOutputKey": "result",
    "negated": false,
    "ignoreCase": false,
    "defaultEvaluationCriteria": {
      "searchText": "expected substring"
    }
  }
}
```

### 3. JSON Similarity (`uipath-json-similarity`)

Compares JSON structures for similarity.

```json
{
  "version": "1.0",
  "id": "JsonSimilarityEvaluator",
  "description": "Compares JSON output for structural similarity.",
  "evaluatorTypeId": "uipath-json-similarity",
  "evaluatorConfig": {
    "name": "JsonSimilarityEvaluator",
    "targetOutputKey": "*",
    "defaultEvaluationCriteria": {
      "expectedOutput": { "result": 5 }
    }
  }
}
```

### 4. LLM Judge — Semantic Similarity (`uipath-llm-judge-output-semantic-similarity`)

Uses an LLM to judge semantic equivalence.

```json
{
  "version": "1.0",
  "id": "LLMJudgeOutputEvaluator",
  "description": "Uses an LLM to judge semantic similarity between expected and actual output.",
  "evaluatorTypeId": "uipath-llm-judge-output-semantic-similarity",
  "evaluatorConfig": {
    "name": "LLMJudgeOutputEvaluator",
    "targetOutputKey": "*",
    "model": "gpt-4.1-2025-04-14",
    "prompt": "Compare the following outputs and evaluate their semantic similarity.\n\nActual Output: {{ActualOutput}}\nExpected Output: {{ExpectedOutput}}\n\nProvide a score from 0-100.",
    "temperature": 0.0,
    "defaultEvaluationCriteria": {
      "expectedOutput": { "result": 5.0 }
    }
  }
}
```

### 5. LLM Judge — Strict JSON Similarity (`uipath-llm-judge-strict-json-similarity`)

Strict JSON comparison via LLM. Same structure as semantic similarity but with
stricter comparison prompt.

### 6. Trajectory (`uipath-llm-judge-trajectory-similarity`)

Evaluates the agent's execution path and decision sequence.

```json
{
  "version": "1.0",
  "id": "TrajectoryEvaluator",
  "description": "Evaluates the agent's execution trajectory and decision sequence.",
  "evaluatorTypeId": "uipath-llm-judge-trajectory-similarity",
  "evaluatorConfig": {
    "name": "TrajectoryEvaluator",
    "model": "gpt-4.1-2025-04-14",
    "prompt": "Evaluate the agent's execution trajectory based on the expected behavior.\n\nExpected Agent Behavior: {{ExpectedAgentBehavior}}\nAgent Run History: {{AgentRunHistory}}\n\nProvide a score from 0-100.",
    "temperature": 0.0,
    "defaultEvaluationCriteria": {
      "expectedAgentBehavior": "The agent should correctly perform the task."
    }
  }
}
```

### 7. Tool Call Arguments (`uipath-tool-call-arguments`)

Validates that the agent called tools with correct arguments.

```json
{
  "version": "1.0",
  "id": "ToolCallArgumentsEvaluator",
  "description": "Validates tool call arguments match expected values.",
  "evaluatorTypeId": "uipath-tool-call-arguments",
  "evaluatorConfig": {
    "name": "ToolCallArgumentsEvaluator",
    "defaultEvaluationCriteria": {
      "expectedToolCalls": [
        {
          "toolName": "Web Search",
          "arguments": { "query": "expected search query" }
        }
      ]
    }
  }
}
```

## Custom Python Evaluators (Coded Agents)

### Evaluator Config

`coded-evals/evaluators/my-evaluator.json`:

```json
{
  "version": "1.0",
  "id": "MyCustomEvaluator",
  "evaluatorTypeId": "file://types/my-evaluator-types.json",
  "evaluatorSchema": "file://my_evaluator.py:MyCustomEvaluator",
  "description": "Custom evaluator description",
  "evaluatorConfig": {
    "name": "MyCustomEvaluator",
    "defaultEvaluationCriteria": {
      "customField": "default-value"
    },
    "negated": false
  }
}
```

### Types File

`coded-evals/evaluators/custom/types/my-evaluator-types.json`:

JSON schema for the custom evaluation criteria fields.

### Python Implementation

`coded-evals/evaluators/custom/my_evaluator.py`:

```python
import json
from uipath.eval.evaluators import BaseEvaluator, BaseEvaluationCriteria, BaseEvaluatorConfig
from uipath.eval.models import AgentExecution, EvaluationResult, NumericEvaluationResult
from opentelemetry.sdk.trace import ReadableSpan

class MyEvaluationCriteria(BaseEvaluationCriteria):
    """Evaluation criteria for the custom evaluator."""
    customField: str

class MyEvaluatorConfig(BaseEvaluatorConfig[MyEvaluationCriteria]):
    """Configuration for the custom evaluator."""
    name: str = "MyCustomEvaluator"
    negated: bool = False
    default_evaluation_criteria: MyEvaluationCriteria = MyEvaluationCriteria(
        customField="default"
    )

class MyCustomEvaluator(
    BaseEvaluator[MyEvaluationCriteria, MyEvaluatorConfig, type(None)]
):
    """Custom evaluator implementation."""

    @classmethod
    def get_evaluator_id(cls) -> str:
        return "MyCustomEvaluator"

    async def evaluate(
        self,
        agent_execution: AgentExecution,
        evaluation_criteria: MyEvaluationCriteria
    ) -> EvaluationResult:
        # Access agent output
        output = agent_execution.output

        # Access agent trace spans
        for span in agent_execution.agent_trace:
            if span.name == "target_operation":
                input_value = json.loads(
                    span.attributes.get("input.value", "{}")
                )
                # Evaluate...

        # Return score 0.0-1.0
        score = 1.0 if condition else 0.0
        if self.evaluator_config.negated:
            score = 1.0 - score

        return NumericEvaluationResult(score=score)
```

### Key Classes

| Class | Description |
|-------|-------------|
| `BaseEvaluator` | Base class for all custom evaluators |
| `BaseEvaluationCriteria` | Base for criteria models (pydantic) |
| `BaseEvaluatorConfig` | Base for evaluator configuration |
| `AgentExecution` | Contains `output`, `agent_trace` (spans) |
| `NumericEvaluationResult` | Result with `score` (0.0-1.0) |
| `ReadableSpan` | OpenTelemetry trace span |

### Mockable Functions for Eval Simulation

```python
from uipath.eval.mocks import ExampleCall, mockable

EXAMPLES = [
    ExampleCall(id="example1", input='{"query":"test"}', output='{"result":"mock"}')
]

@traced()
@mockable(example_calls=EXAMPLES)
async def my_tool(query: str) -> dict:
    # Real implementation
    ...
```

The `@mockable` decorator allows evals to simulate tool responses.

## Evaluator Quick Reference

| ID | Type | Score | Criteria Key |
|----|------|-------|-------------|
| `uipath-exact-match` | Deterministic | 0/1 | `expectedOutput` |
| `uipath-contains` | Deterministic | 0/1 | `searchText` |
| `uipath-json-similarity` | Deterministic | 0-1 | `expectedOutput` |
| `uipath-llm-judge-output-semantic-similarity` | LLM | 0-100 | `expectedOutput` |
| `uipath-llm-judge-strict-json-similarity` | LLM | 0-100 | `expectedOutput` |
| `uipath-llm-judge-trajectory-similarity` | LLM | 0-100 | `expectedAgentBehavior` |
| `uipath-tool-call-arguments` | Deterministic | 0/1 | `expectedToolCalls` |
| `file://` custom | Custom Python | 0-1 | Custom fields |
