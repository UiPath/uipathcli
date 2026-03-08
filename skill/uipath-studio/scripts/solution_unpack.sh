#!/usr/bin/env bash
# Unpacks a .uis file (ZIP archive) into a directory.
#
# Usage: solution_unpack.sh <uis_file> [output_directory]
#
# If output_directory is not specified, creates a directory named after the
# .uis file (without extension).
#
# Example: solution_unpack.sh MySolution.uis
#          solution_unpack.sh MySolution.uis ./my-output-dir

set -euo pipefail

UIS_FILE="${1:?Usage: solution_unpack.sh <uis_file> [output_directory]}"

if [ ! -f "${UIS_FILE}" ]; then
    echo "ERROR: File not found: ${UIS_FILE}" >&2
    exit 1
fi

# Determine output directory
if [ -n "${2:-}" ]; then
    OUTPUT_DIR="$2"
else
    BASENAME=$(basename "${UIS_FILE}" .uis)
    OUTPUT_DIR="${BASENAME}"
fi

# Create output directory
mkdir -p "${OUTPUT_DIR}"

# Extract
unzip -o "${UIS_FILE}" -d "${OUTPUT_DIR}"

# Validate
if [ ! -f "${OUTPUT_DIR}/SolutionStorage.json" ]; then
    echo "WARNING: No SolutionStorage.json found. This may not be a valid .uis file." >&2
fi

echo ""
echo "Unpacked to: ${OUTPUT_DIR}/"

# Show solution info
if [ -f "${OUTPUT_DIR}/SolutionStorage.json" ]; then
    SOLUTION_ID=$(python3 -c "import json; d=json.load(open('${OUTPUT_DIR}/SolutionStorage.json')); print(d.get('SolutionId','unknown'))" 2>/dev/null || echo "unknown")
    PROJECT_COUNT=$(python3 -c "import json; d=json.load(open('${OUTPUT_DIR}/SolutionStorage.json')); print(len(d.get('Projects',[])))" 2>/dev/null || echo "unknown")
    echo "Solution ID: ${SOLUTION_ID}"
    echo "Projects: ${PROJECT_COUNT}"
fi

# List projects from .uipx if available
UIPX_FILE=$(find "${OUTPUT_DIR}" -maxdepth 1 -name "*.uipx" -type f | head -1)
if [ -n "${UIPX_FILE}" ]; then
    echo ""
    echo "Project types:"
    python3 -c "
import json
d = json.load(open('${UIPX_FILE}'))
for p in d.get('Projects', []):
    print(f\"  - {p.get('Type', 'Unknown'):25s} {p.get('ProjectRelativePath', '')}\")
" 2>/dev/null || true
fi
