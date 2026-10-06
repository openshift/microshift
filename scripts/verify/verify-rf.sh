#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)

RF_VENV="${ROOTDIR}/_output/robotenv"
ROBOT_PRIVACY_SENTINEL="synthetic-sensitive-output-must-not-be-published"
export ROBOT_PRIVACY_SENTINEL
"${ROOTDIR}/scripts/fetch_tools.sh" robotframework

cd "${ROOTDIR}/test"

# Configured robocop rules:
# https://robocop.readthedocs.io/en/stable/rules.html#too-long-test-case-w0504
# https://robocop.readthedocs.io/en/stable/rules.html#too-many-calls-in-test-case-w0505

set -x
"${RF_VENV}/bin/robocop" check

"${RF_VENV}/bin/robocop" format --check --diff --no-overwrite

"${RF_VENV}/bin/robot" \
    --loglevel TRACE \
    --pythonpath "${ROOTDIR}/test/resources" \
    --pythonpath "${ROOTDIR}/test/unit" \
    --outputdir "${ROOTDIR}/_output/robot-unit" \
    "${ROOTDIR}/test/unit"

set +x
if grep --recursive --fixed-strings --quiet \
    "${ROBOT_PRIVACY_SENTINEL}" \
    "${ROOTDIR}/_output/robot-unit"; then
    echo "Sensitive test sentinel found in Robot logs or artifacts" >&2
    exit 1
fi
