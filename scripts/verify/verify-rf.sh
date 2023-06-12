#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)

RF_VENV="${ROOTDIR}/_output/robotenv"
"${ROOTDIR}/scripts/fetch_tools.sh" robotframework

cd "${ROOTDIR}/test"
"${RF_VENV}/bin/robocop"

"${RF_VENV}/bin/robotidy" --check --diff .
