#!/usr/bin/env bash
# Validates a UiPath agent project structure.
#
# Usage: validate_agent.sh <agent_directory>
#
# Checks for required files, valid JSON, schema consistency, and common issues.
#
# Example: validate_agent.sh ./Agent
#          validate_agent.sh ./MySolution/Agent

set -euo pipefail

AGENT_DIR="${1:?Usage: validate_agent.sh <agent_directory>}"
AGENT_DIR="${AGENT_DIR%/}"

ERRORS=0
WARNINGS=0

error() { echo "  ERROR: $1" >&2; ERRORS=$((ERRORS + 1)); }
warn()  { echo "  WARN:  $1" >&2; WARNINGS=$((WARNINGS + 1)); }
ok()    { echo "  OK:    $1"; }

echo "Validating agent: ${AGENT_DIR}"
echo "=========================================="

# 1. Check required files
echo ""
echo "Required files:"

if [ -f "${AGENT_DIR}/agent.json" ]; then
    ok "agent.json exists"
else
    error "agent.json missing"
fi

if [ -f "${AGENT_DIR}/entry-points.json" ]; then
    ok "entry-points.json exists"
else
    error "entry-points.json missing"
fi

if [ -f "${AGENT_DIR}/project.uiproj" ]; then
    ok "project.uiproj exists"
else
    error "project.uiproj missing"
fi

# 2. Validate JSON files
echo ""
echo "JSON validation:"

for f in "${AGENT_DIR}/agent.json" "${AGENT_DIR}/entry-points.json" "${AGENT_DIR}/project.uiproj"; do
    if [ -f "$f" ]; then
        if python3 -c "import json; json.load(open('$f'))" 2>/dev/null; then
            ok "$(basename "$f") is valid JSON"
        else
            error "$(basename "$f") is NOT valid JSON"
        fi
    fi
done

# 3. Check project type
echo ""
echo "Project type:"

if [ -f "${AGENT_DIR}/project.uiproj" ]; then
    PROJECT_TYPE=$(python3 -c "import json; print(json.load(open('${AGENT_DIR}/project.uiproj')).get('ProjectType',''))" 2>/dev/null || echo "")
    if [ "${PROJECT_TYPE}" = "Agent" ]; then
        ok "ProjectType is 'Agent'"
    else
        error "ProjectType is '${PROJECT_TYPE}', expected 'Agent'"
    fi
fi

# 4. Check agent type and structure
echo ""
echo "Agent configuration:"

if [ -f "${AGENT_DIR}/agent.json" ]; then
    AGENT_TYPE=$(python3 -c "import json; print(json.load(open('${AGENT_DIR}/agent.json')).get('type',''))" 2>/dev/null || echo "")

    if [ "${AGENT_TYPE}" = "lowCode" ]; then
        ok "Agent type: lowCode"

        # Check low-code specific files
        if [ -d "${AGENT_DIR}/.agent-builder" ]; then
            ok ".agent-builder/ directory exists"
        else
            warn ".agent-builder/ directory missing (recommended for low-code agents)"
        fi

        # Check messages
        MSG_COUNT=$(python3 -c "import json; d=json.load(open('${AGENT_DIR}/agent.json')); print(len(d.get('messages',[])))" 2>/dev/null || echo "0")
        if [ "${MSG_COUNT}" -ge 2 ]; then
            ok "Messages defined (${MSG_COUNT} messages)"
        else
            warn "Less than 2 messages defined (system + user recommended)"
        fi

    elif [ "${AGENT_TYPE}" = "coded" ]; then
        ok "Agent type: coded"

        # Check coded-specific files
        if [ -f "${AGENT_DIR}/source_code/main.py" ]; then
            ok "source_code/main.py exists"
        else
            error "source_code/main.py missing (required for coded agents)"
        fi

        if [ -f "${AGENT_DIR}/source_code/pyproject.toml" ]; then
            ok "source_code/pyproject.toml exists"
        else
            error "source_code/pyproject.toml missing"
        fi

        if [ -f "${AGENT_DIR}/source_code/uipath.json" ]; then
            ok "source_code/uipath.json exists"
        else
            warn "source_code/uipath.json missing"
        fi
    else
        error "Unknown agent type: '${AGENT_TYPE}'"
    fi

    # Check model configuration
    MODEL=$(python3 -c "import json; d=json.load(open('${AGENT_DIR}/agent.json')); print(d.get('settings',{}).get('model',''))" 2>/dev/null || echo "")
    if [ -n "${MODEL}" ]; then
        ok "Model configured: ${MODEL}"
    else
        warn "No model configured in settings"
    fi

    # Check schemas
    HAS_INPUT=$(python3 -c "import json; d=json.load(open('${AGENT_DIR}/agent.json')); print('yes' if d.get('inputSchema') else 'no')" 2>/dev/null || echo "no")
    HAS_OUTPUT=$(python3 -c "import json; d=json.load(open('${AGENT_DIR}/agent.json')); print('yes' if d.get('outputSchema') else 'no')" 2>/dev/null || echo "no")

    if [ "${HAS_INPUT}" = "yes" ]; then
        ok "inputSchema defined"
    else
        warn "No inputSchema defined"
    fi
    if [ "${HAS_OUTPUT}" = "yes" ]; then
        ok "outputSchema defined"
    else
        warn "No outputSchema defined"
    fi
fi

# 5. Check resources/tools
echo ""
echo "Resources/Tools:"

if [ -d "${AGENT_DIR}/resources" ]; then
    RESOURCE_COUNT=$(find "${AGENT_DIR}/resources" -name "resource.json" -type f | wc -l)
    ok "${RESOURCE_COUNT} resource(s) found"

    for res in "${AGENT_DIR}/resources"/*/resource.json; do
        if [ -f "$res" ]; then
            RES_NAME=$(python3 -c "import json; print(json.load(open('$res')).get('name','unknown'))" 2>/dev/null || echo "unknown")
            RES_TYPE=$(python3 -c "import json; print(json.load(open('$res')).get('\$resourceType','unknown'))" 2>/dev/null || echo "unknown")
            ok "  ${RES_NAME} (${RES_TYPE})"
        fi
    done
else
    ok "No resources directory (agent has no tools)"
fi

# 6. Check evaluations
echo ""
echo "Evaluations:"

for eval_dir in "${AGENT_DIR}/evals" "${AGENT_DIR}/coded-evals"; do
    if [ -d "${eval_dir}" ]; then
        EVAL_SET_COUNT=$(find "${eval_dir}/eval-sets" -name "*.json" -type f 2>/dev/null | wc -l)
        EVALUATOR_COUNT=$(find "${eval_dir}/evaluators" -name "*.json" -type f 2>/dev/null | wc -l)
        ok "$(basename "${eval_dir}")/: ${EVAL_SET_COUNT} eval set(s), ${EVALUATOR_COUNT} evaluator(s)"
    fi
done

if [ ! -d "${AGENT_DIR}/evals" ] && [ ! -d "${AGENT_DIR}/coded-evals" ]; then
    warn "No evaluation sets found"
fi

# Summary
echo ""
echo "=========================================="
if [ ${ERRORS} -eq 0 ] && [ ${WARNINGS} -eq 0 ]; then
    echo "PASS: Agent structure is valid"
elif [ ${ERRORS} -eq 0 ]; then
    echo "PASS with ${WARNINGS} warning(s)"
else
    echo "FAIL: ${ERRORS} error(s), ${WARNINGS} warning(s)"
    exit 1
fi
