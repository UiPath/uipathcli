#!/usr/bin/env bash
# Packs a UiPath solution directory into a .uis file (ZIP archive).
#
# Usage: solution_pack.sh <solution_directory> [output_file]
#
# If output_file is not specified, creates <directory_name>.uis in the
# current directory.
#
# Example: solution_pack.sh ./MySolution
#          solution_pack.sh ./MySolution ~/Desktop/MySolution.uis

set -euo pipefail

SOLUTION_DIR="${1:?Usage: solution_pack.sh <solution_directory> [output_file]}"
SOLUTION_DIR="${SOLUTION_DIR%/}"  # Remove trailing slash

# Resolve to absolute path
SOLUTION_DIR="$(cd "${SOLUTION_DIR}" && pwd)"

# Determine output filename (absolute path)
if [ -n "${2:-}" ]; then
    OUTPUT_FILE="$(cd "$(dirname "$2")" 2>/dev/null && pwd)/$(basename "$2")" 2>/dev/null || OUTPUT_FILE="$(pwd)/$2"
else
    BASENAME=$(basename "${SOLUTION_DIR}")
    OUTPUT_FILE="$(pwd)/${BASENAME}.uis"
fi

# Validate solution structure
if [ ! -f "${SOLUTION_DIR}/SolutionStorage.json" ]; then
    echo "ERROR: ${SOLUTION_DIR}/SolutionStorage.json not found." >&2
    echo "This does not appear to be a valid UiPath solution directory." >&2
    exit 1
fi

# Find .uipx manifest
UIPX_FILE=$(find "${SOLUTION_DIR}" -maxdepth 1 -name "*.uipx" -type f | head -1)
if [ -z "${UIPX_FILE}" ]; then
    echo "ERROR: No .uipx manifest found in ${SOLUTION_DIR}/" >&2
    exit 1
fi

# Remove existing output file if present
if [ -f "${OUTPUT_FILE}" ]; then
    rm "${OUTPUT_FILE}"
fi

# Create ZIP archive from inside the solution directory
cd "${SOLUTION_DIR}"
zip -r "${OUTPUT_FILE}" . -x ".git/*" -x "__pycache__/*" -x "*.pyc"
cd - > /dev/null

echo "Packed solution: ${OUTPUT_FILE}"
echo "Size: $(du -h "${OUTPUT_FILE}" | cut -f1)"
