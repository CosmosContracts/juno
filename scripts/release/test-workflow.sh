#!/bin/bash
set -e

# Get the directory where the script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"

# Run the test workflow using act
act workflow_dispatch \
  -W "${SCRIPT_DIR}/../../.github/workflows/test-release-dispatch.yml" \
  -e "${SCRIPT_DIR}/test-event.json" \
  --secret-file "${SCRIPT_DIR}/.act-secrets.env"

echo "Test workflow completed!"