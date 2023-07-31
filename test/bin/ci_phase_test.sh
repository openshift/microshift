#!/bin/bash
#
# This script runs on the CI cluster, from the metal-tests step.

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

cd "${TESTDIR}"

if [ ! -d "${RF_VENV}" ]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" robotframework
fi

# Set up the scenario sources
[[ -z "${PERIODICS}" ]] && SCENARIO_SOURCES="scenarios/*.sh" || SCENARIO_SOURCES="scenarios/periodics/*.sh"

for scenario in ${SCENARIO_SOURCES}; do
    scenario_name="$(basename "${scenario}" .sh)"
    logfile="${SCENARIO_INFO_DIR}/${scenario_name}/run.log"
    mkdir -p "$(dirname "${logfile}")"
    SSH_PRIVATE_KEY="${HOME}/.ssh/id_rsa" bash -x ./bin/scenario.sh run "${scenario}" >"${logfile}" 2>&1 &
done

FAIL=0
for job in $(jobs -p) ; do
    jobs -l
    echo "Waiting for job: ${job}"
    wait "${job}" || ((FAIL+=1))
done

echo "Test phase complete"
if [[ ${FAIL} -ne 0 ]]; then
    exit 1
fi
