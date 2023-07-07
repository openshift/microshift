#!/bin/bash

set -e

SCRIPTDIR="$(dirname "${BASH_SOURCE[0]}")"
REPOROOT="$(git rev-parse --show-toplevel)"
OUTPUT_DIR="${REPOROOT}/_output"
ENVDIR="${OUTPUT_DIR}/jira"

if [ ! -d "${ENVDIR}" ]; then
    echo "Setting up required tools..."
    mkdir -p "${OUTPUT_DIR}"
    python3 -m venv "${ENVDIR}"
    "${ENVDIR}/bin/pip3" install jira
fi

"${ENVDIR}/bin/python3" "${SCRIPTDIR}/manage_ticket.py" "$@"
