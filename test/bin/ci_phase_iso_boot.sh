#!/bin/bash
#
# This script runs on the hypervisor to boot all scenario VMs.

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

cd "${ROOTDIR}"

# Make sure libvirtd is running. We do this here, because some of the
# other scripts use virsh.
bash -x ./scripts/devenv-builder/manage-vm.sh config

# Clean up the image builder cache to free disk for virtual machines
bash -x ./scripts/image-builder/cleanup.sh -full

cd "${ROOTDIR}/test"

# Set up the hypervisor configuration for the tests
bash -x ./bin/manage_hypervisor_config.sh create

# Start the web server to host the kickstart files and ostree commit
# repository.
bash -x ./bin/start_webserver.sh

declare -A pidToScenario

# Build all of the needed VMs
for scenario in "${SCENARIO_SOURCES}"/*.sh; do
    scenario_name=$(basename "${scenario}" .sh)
    logfile="${SCENARIO_INFO_DIR}/${scenario_name}/boot.log"
    mkdir -p "$(dirname "${logfile}")"
    bash -x ./bin/scenario.sh create "${scenario}" >"${logfile}" 2>&1 &
    pidToScenario["$!"]="${scenario}"
    sleep 5
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

echo "===================================="
echo "System information after booting VMs"
echo "===================================="
free -h
df -h
sudo du -sk "${IMAGEDIR}"/* | sort -n
sudo virsh list --all
echo "===================================="

if [ ${FAIL} -ne 0 ]; then
    echo "Failed to boot all VMs"
    exit 1
fi

echo "Boot phase complete"
