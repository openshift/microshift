#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENVDIR="${SCRIPT_DIR}/.venv"

if [ ! -d "${ENVDIR}" ]; then
    echo "Setting up required tools..."
    python3.11 -m venv "${ENVDIR}"
    "${ENVDIR}/bin/pip3" install --upgrade pip
    "${ENVDIR}/bin/pip3" install --quiet -r "${SCRIPT_DIR}/requirements.txt"
fi

# shellcheck source=/dev/null
source "${SCRIPT_DIR}/mcp-server-vars.env"

"${SCRIPT_DIR}"/.venv/bin/python3 "${SCRIPT_DIR}"/mcp-server.py
