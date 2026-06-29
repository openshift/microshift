#!/bin/bash
set -euo pipefail

# Run a single Robot Framework test against a remote MicroShift cluster.
#
# Usage:
#   ./test/bin/run-single-test.sh <variables.yaml> <robot-file> [test-name]
#
# Examples:
#   ./test/bin/run-single-test.sh test/variables.yaml suites/router/router-config-infra.robot
#   ./test/bin/run-single-test.sh test/variables.yaml suites/router/router-config-infra.robot 'Custom Listening IPs And Ports'

VARIABLES_FILE="$(cd "$(dirname "${1:?Usage: $0 <variables.yaml> <robot-file> [test-name]}")" && pwd)/$(basename "$1")"
ROBOT_FILE="${2:?Usage: $0 <variables.yaml> <robot-file> [test-name]}"
TEST_NAME="${3:-}"

ROOTDIR="$(cd "$(dirname "$0")/../.." && pwd)"
TESTDIR="${ROOTDIR}/test"
RF_VENV="${ROOTDIR}/_output/robotenv"
OUTDIR="${ROOTDIR}/_output/rf-single-$(date +%Y%m%d-%H%M%S)"

if [[ ! -f "${VARIABLES_FILE}" ]]; then
    echo "ERROR: Variables file not found: ${VARIABLES_FILE}"
    echo ""
    echo "Create one from the example:"
    echo "  cp test/variables.yaml.example test/variables.yaml"
    echo ""
    echo "Required fields:"
    echo "  USHIFT_HOST: <host-ip>"
    echo "  USHIFT_USER: <ssh-user>"
    echo "  SSH_PRIV_KEY: <path-to-key-or-empty-for-agent>"
    echo "  SSH_PORT: 22"
    echo "  API_PORT: 6443"
    exit 1
fi

if [[ ! -d "${RF_VENV}" ]]; then
    echo "Setting up Robot Framework venv..."
    "${ROOTDIR}/scripts/fetch_tools.sh" robotframework
fi

RF_BINARY="${RF_VENV}/bin/robot"
if [[ ! -f "${RF_BINARY}" ]]; then
    echo "ERROR: robot binary not found at ${RF_BINARY}"
    exit 1
fi

mkdir -p "${OUTDIR}"

TEST_FLAG=()
if [[ -n "${TEST_NAME}" ]]; then
    TEST_FLAG=(-t "${TEST_NAME}")
    echo "=== Running test: ${TEST_NAME} ==="
else
    echo "=== Running all tests ==="
fi
echo "Variables: ${VARIABLES_FILE}"
echo "Robot file: ${ROBOT_FILE}"
echo "Output dir: ${OUTDIR}"
echo ""

cd "${TESTDIR}"

exec "${RF_BINARY}" \
    --loglevel TRACE \
    -V "${VARIABLES_FILE}" \
    --outputdir "${OUTDIR}" \
    --pythonpath resources \
    ${TEST_FLAG+"${TEST_FLAG[@]}"} \
    "${ROBOT_FILE}"
