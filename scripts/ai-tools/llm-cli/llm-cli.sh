#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENVDIR="${SCRIPT_DIR}/.venv"

if [ ! -d "${ENVDIR}" ]; then
    echo "Setting up required tools..."
    python3 -m venv "${ENVDIR}"
    "${ENVDIR}/bin/pip3" install --quiet -r "${SCRIPT_DIR}/requirements.txt"
fi

"${ENVDIR}/bin/python3" "${SCRIPT_DIR}/llm-cli.py" "${1}" "${@:2}"
