#!/bin/bash
#
# This script runs on the hypervisor to execute all scenario tests.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

TEST_JOB_LOG="${IMAGEDIR}/test_jobs.txt"

cd "${TESTDIR}"

if [ ! -d "${RF_VENV}" ]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" robotframework
fi

# Tell scenario.sh to merge stderr into stdout
export SCENARIO_MERGE_OUTPUT_STREAMS=true

# Show the summary of the output of the parallel jobs.
if [ -t 0 ]; then
    progress="--progress"
else
    progress=""
fi

TEST_OK=true
if ! parallel \
    ${progress} \
    --results "${SCENARIO_INFO_DIR}/{/.}/run.log" \
    --joblog "${TEST_JOB_LOG}" \
    bash -x ./bin/scenario.sh run ::: "${SCENARIOS_TO_RUN}"/*.sh ; then
   TEST_OK=false
fi

cat "${TEST_JOB_LOG}"

echo "Test phase complete"
if ! "${TEST_OK}"; then
    echo "Some tests failed"
    exit 1
fi
exit 0
