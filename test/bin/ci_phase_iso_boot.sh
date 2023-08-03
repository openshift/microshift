#!/bin/bash
#
# This script runs on the hypervisor, from the iso-build step.

set -xeuo pipefail
PS4='+ $(date "+%T.%N")\011 '

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# Log output automatically
LOGDIR="${ROOTDIR}/_output/ci-logs"
LOGFILE="${LOGDIR}/$(basename "$0" .sh).log"
if [ ! -d "${LOGDIR}" ]; then
    mkdir -p "${LOGDIR}"
fi
echo "Logging to ${LOGFILE}"
# Set fd 1 and 2 to write to the log file
exec &> >(tee >(awk '{ print strftime("%Y-%m-%d %H:%M:%S"), $0; fflush() }' >"${LOGFILE}"))

cd ~/microshift/test

# Start the web server to host the kickstart files and ostree commit
# repository.
bash -x ./bin/start_webserver.sh

# Set up the storage pool for VMs
bash -x ./bin/manage_vm_storage_pool.sh create

declare -A pidToScenario

# Build all of the needed VMs
for scenario in scenarios/*.sh; do
    scenario_name=$(basename "${scenario}" .sh)
    logfile="${SCENARIO_INFO_DIR}/${scenario_name}/boot.log"
    mkdir -p "$(dirname "${logfile}")"
    bash -x ./bin/scenario.sh create "${scenario}" >"${logfile}" 2>&1 &
    pidToScenario["$!"]="${scenario}"
done

set +x
for pid in "${!pidToScenario[@]}"; do echo "${pid} - ${pidToScenario[${pid}]}"; done
set -x

FAIL=0
for job in $(jobs -p); do
    jobs -l
    echo "Waiting for job: ${job}"
    if ! wait "${job}"; then
        ((FAIL += 1))
        echo "Failed to boot VMs for scenario: ${pidToScenario[${job}]}"
    fi
done

if [ ${FAIL} -ne 0 ]; then
    echo "Failed to boot all VMs"
    exit 1
fi

echo "Boot phase complete"
