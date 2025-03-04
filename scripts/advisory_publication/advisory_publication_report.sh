#!/bin/bash

set -e

SCRIPTDIR="$(dirname "${BASH_SOURCE[0]}")"
REPOROOT="$(git rev-parse --show-toplevel)"
OUTPUT_DIR="${REPOROOT}/_output"
ENVDIR="${OUTPUT_DIR}/advisory_publication"

if [ ! -d "${ENVDIR}" ]; then
    echo "Setting up required tools..."
    mkdir -p "${OUTPUT_DIR}"
    python3 -m venv "${ENVDIR}"
    "${ENVDIR}/bin/pip3" install -r "${SCRIPTDIR}/requirements.txt"
fi

"${ENVDIR}/bin/python3" "${SCRIPTDIR}/advisory_publication_report.py" "$@"
