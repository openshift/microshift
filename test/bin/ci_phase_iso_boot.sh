#!/bin/bash
#
# This script runs on the hypervisor to boot all scenario VMs.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

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

LAUNCH_VMS_JOB_LOG="${IMAGEDIR}/launch_vm_jobs.txt"

# Copy the scenario definition files to a temporary location from
# which they will be read. This allows filtering the tests from a
# specific $SCENARIO_SOURCES directory.
prepare_scenario_sources() {
    rm -rf "${SCENARIOS_TO_RUN}"
    mkdir -p "${SCENARIOS_TO_RUN}"
    cp "${SCENARIO_SOURCES}"/*.sh "${SCENARIOS_TO_RUN}"/
    if ${EXCLUDE_CNCF_CONFORMANCE}; then
        find "${SCENARIOS_TO_RUN}" -name "*cncf-conformance.sh" -delete
    fi
}

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

# Show the summary of the output of the parallel jobs.
if [ -t 0 ]; then
    progress="--progress"
else
    progress=""
fi

# Tell scenario.sh to merge stderr into stdout
export SCENARIO_MERGE_OUTPUT_STREAMS=true

# Prepare all the scenarios that need to run into a special directory
prepare_scenario_sources

LAUNCH_OK=true
if ! parallel \
    ${progress} \
    --results "${SCENARIO_INFO_DIR}/{/.}/boot.log" \
    --joblog "${LAUNCH_VMS_JOB_LOG}" \
    --delay 5 \
    bash -x ./bin/scenario.sh create ::: "${SCENARIOS_TO_RUN}"/*.sh ; then
   LAUNCH_OK=false
fi

# Show the summary of the output of the parallel jobs.
cat "${LAUNCH_VMS_JOB_LOG}"

echo "===================================="
echo "System information after booting VMs"
echo "===================================="
free -h
df -h
sudo du -sk "${IMAGEDIR}"/* | sort -n
sudo virsh list --all
echo "===================================="

if ! "${LAUNCH_OK}"; then
    echo "Failed to boot all VMs"
    exit 1
fi

echo "Boot phase complete"
exit 0
