#!/usr/bin/bash

set -euo pipefail

SCRIPTDIR="$(dirname "${BASH_SOURCE[0]}")"
REPOROOT="$(git rev-parse --show-toplevel)"
ENVDIR="${REPOROOT}/_output/release_testing"

if [[ ! -d "${ENVDIR}" ]]; then
    echo "Setting up required tools..." >&2
    mkdir -p "${REPOROOT}/_output"
    python3 -m venv "${ENVDIR}"
fi
"${ENVDIR}/bin/python3" -m pip install -r "${SCRIPTDIR}/requirements.txt" >&2

cd "${SCRIPTDIR}"
"${ENVDIR}/bin/python3" -m unittest unit_tests.test_logic -v
