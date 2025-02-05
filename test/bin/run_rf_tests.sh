#!/bin/bash

set -xeuo pipefail
IFS=$'\n\t'

# path vars
TEST_BIN_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# default config
DRYRUN=""
OUTDIR="${TEST_BIN_DIR}/_output/e2e-$(date +%Y%m%d-%H%M%S)"
RF_VENV="${TEST_BIN_DIR}/_output/robotenv"
RF_VARIABLES="${TEST_BIN_DIR}/variables.yaml"
SCENARIO="Default Name"
TEST_RANDOMIZATION="all"
TEST_EXECUTION_TIMEOUT="300m"
SCENARIO="Default Scenario"
EXITONFAILURE=""

function usage {
    local -r script_name=$(basename "$0")
    cat - <<EOF
${script_name} [-h] [-n] [-o output_dir] [-v venv_dir] [-i var_file] [-c scenario] [-b rf_binary] [-s name=value] [-k test_names] [-r randomize_value] [-t timeout_value] [-x] [test suite files]

Options:

  -h                           Print this help text.
  -n                           Dry-run, do not run the tests.
  -o DIR                       The output directory. (${OUTDIR})
  -v DIR                       The venv directory. (${RF_VENV})
  -i PATH                      The variables file. (${RF_VARIABLES})
  -c scenario                  RobotFramework scenario (${SCENARIO})
  -b RF_BINARY                 RobotFramework binary (${RF_BINARY})
  -s NAME=VALUE                To enable a stress condition.
  -k SKIP_TESTS                Comma separated list of tests to skip.
  -r TEST_RANDOMIZATION        Define RF Test order (${TEST_RANDOMIZATION})
  -r TEST_EXECUTION_TIMEOUT    RF execution timeout (${TEST_EXECUTION_TIMEOUT})
  -x                           Stops test execution if any test fails. The remaining tests are marked as failed without actually executing them.
EOF
}

while getopts "hno:v:i:b:c:s:k:r:t:x" opt; do
    case ${opt} in
        h)
            usage
            exit 0
            ;;
        n)
            DRYRUN=--dryrun
            ;;
        o)
            OUTDIR=${OPTARG}
            ;;
        v)
            RF_VENV=${OPTARG}
            ;;
        i)
            RF_VARIABLES=${OPTARG}
            ;;
        b)
            RF_BINARY=${OPTARG}
            ;;
        c)
            SCENARIO=${OPTARG}
            ;;
        s)
            STRESS_TESTING=${OPTARG}
            ;;
        k)
            SKIP_TESTS=${OPTARG}
            ;;
        r)
            TEST_RANDOMIZATION=${OPTARG}
            ;;
        t)
            TEST_EXECUTION_TIMEOUT=${OPTARG}
            ;;
        x)
            EXITONFAILURE="--exitonfailure"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done
shift $((OPTIND-1))

if [ ! -f "${RF_VARIABLES}" ]; then
    echo "Please create or provide a variables file at ${RF_VARIABLES}" 1>&2
    echo "See ${TEST_BIN_DIR}/variables.yaml.example for the expected content." 1>&2
    exit 1
fi

cd "${TEST_BIN_DIR}" || (echo "Did not find ${TEST_BIN_DIR}" 1>&2; exit 1)

TESTS="$*"
if [ -z "${TESTS}" ]; then
    echo "ERROR: missing which RF suites/tests to run"
    exit 1
fi

# enable stress condition
if [ "${STRESS_TESTING:-}" ]; then
    # DEST_DIR var is the python env dir used by fetch_tools.sh to install the tools
    export DEST_DIR="${RF_VENV}"
    "${TEST_BIN_DIR}/../../scripts/fetch_tools.sh" yq
    YQ_BINARY="${RF_VENV}/yq"

    CONDITION="${STRESS_TESTING%=*}"
    VALUE="${STRESS_TESTING#*=}"

    SSH_HOST=$("${YQ_BINARY}" '.USHIFT_HOST' "${RF_VARIABLES}")
    SSH_USER=$("${YQ_BINARY}" '.USHIFT_USER' "${RF_VARIABLES}")
    SSH_PORT=$("${YQ_BINARY}" '.SSH_PORT' "${RF_VARIABLES}")
    SSH_PKEY=$("${YQ_BINARY}" '.SSH_PRIV_KEY' "${RF_VARIABLES}")

    "${TEST_BIN_DIR}/stress_testing.sh" -e "${CONDITION}" -v "${VALUE}" -h "${SSH_HOST}" -u "${SSH_USER}" -p "${SSH_PORT}" -k "${SSH_PKEY}"
fi

# Make sure the test execution times out after a predefined period.
# The 'timeout' command sends the HUP signal and, if the test does not
# exit after 5m, it sends the KILL signal to terminate the process.
timeout_robot="timeout -v --kill-after=5m ${TEST_EXECUTION_TIMEOUT} ${RF_BINARY}"
if [ -t 0 ]; then
    # Disable timeout for interactive mode when stdin is a terminal.
    # This is necessary for proper handling of test interruption by user.
    timeout_robot="${RF_BINARY}"
fi

# shellcheck disable=SC2086,SC2068
"${timeout_robot}" \
    ${DRYRUN} \
    ${EXITONFAILURE}  \
    --name "${SCENARIO}" \
    --randomize "${TEST_RANDOMIZATION}" \
    --prerunmodifier "${TEST_BIN_DIR}/../resources/SkipTests.py:${SKIP_TESTS:-}" \
    --loglevel TRACE \
    --outputdir "${OUTDIR}" \
    --debugfile "${OUTDIR}/rf-debug.log" \
    -V "${RF_VARIABLES}" \
    -x junit.xml \
    ${TESTS[@]}

# disable stress condition
if [ "${STRESS_TESTING:-}" ]; then
    "${TEST_BIN_DIR}/stress_testing.sh" -d "${CONDITION}" -h "${SSH_HOST}" -u "${SSH_USER}" -p "${SSH_PORT}" -k "${SSH_PKEY}"
fi
