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

ROBOT_OUTPUT_DIR="${ROOTDIR}/_output/robot-unit"
mkdir -p "${ROBOT_OUTPUT_DIR}"

# Capture the robot process console (stdout+stderr) alongside the Robot log
# files. Prow publishes this console output too, so the sentinel scan below must
# cover it, not just output.xml/log.html. Keep running on robot failure so a
# failing test that also leaks is still reported, then propagate robot's status.
set +e
"${RF_VENV}/bin/robot" \
    --loglevel TRACE \
    --pythonpath "${ROOTDIR}/test/resources" \
    --pythonpath "${ROOTDIR}/test/unit" \
    --outputdir "${ROBOT_OUTPUT_DIR}" \
    "${ROOTDIR}/test/unit" 2>&1 | tee "${ROBOT_OUTPUT_DIR}/robot-console.log"
ROBOT_PIPE_STATUS=("${PIPESTATUS[@]}")
set -e

set +x
if grep --recursive --fixed-strings --quiet \
    "${ROBOT_PRIVACY_SENTINEL}" \
    "${ROBOT_OUTPUT_DIR}"; then
    echo "Sensitive test sentinel found in Robot logs or artifacts" >&2
    exit 1
fi

# Fail closed if tee could not capture the console: an unscanned console could
# hide a leak that prow still publishes.
if [[ "${ROBOT_PIPE_STATUS[1]}" -ne 0 ]]; then
    echo "Failed to capture ${ROBOT_OUTPUT_DIR}/robot-console.log for scanning" >&2
    exit 1
fi

if [[ "${ROBOT_PIPE_STATUS[0]}" -ne 0 ]]; then
    exit "${ROBOT_PIPE_STATUS[0]}"
fi
