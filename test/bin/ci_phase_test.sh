#!/bin/bash
#
# This script runs on the CI cluster, from the metal-tests step.

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

cd "${TESTDIR}"

# Start the web server to host ostree commit repository.
bash -x ./bin/start_webserver.sh

if [ ! -d "${RF_VENV}" ]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" robotframework
fi

for scenario in ./scenarios/*.sh; do
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

# Kill the web server
pkill caddy || true

echo "Test phase complete"
if [[ ${FAIL} -ne 0 ]]; then
    exit 1
fi
