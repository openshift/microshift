#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)

RF_VENV="${ROOTDIR}/_output/robotenv"
"${ROOTDIR}/scripts/fetch_tools.sh" robotframework

cd "${ROOTDIR}/test"

# Configured robocop rules:
# https://robocop.readthedocs.io/en/stable/rules.html#too-long-test-case-w0504
# https://robocop.readthedocs.io/en/stable/rules.html#too-many-calls-in-test-case-w0505

set -x
"${RF_VENV}/bin/robocop" check

"${RF_VENV}/bin/robocop" format --check --diff --no-overwrite
