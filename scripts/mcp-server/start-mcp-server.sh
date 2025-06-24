#!/bin/bash

set -euo pipefail
# Load environment variables from .env file next to this script
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

export KUBECONFIG_PATH="${KUBECONFIG_PATH:-/Users/agullon/workspace/mcp.kubeconfig}"
export SSH_IP_ADDR="${SSH_IP_ADDR:-10.1.235.14}"
export SSH_USER="${SSH_USER:-microshift}"
export SSH_CONFIG_FILE="${SSH_CONFIG_FILE:-/Users/agullon/.ssh/config}"

"${SCRIPT_DIR}"/.venv/bin/python3 "${SCRIPT_DIR}"/main.py
