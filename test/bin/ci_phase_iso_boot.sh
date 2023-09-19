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

BOOT_LOG="${IMAGEDIR}/boot_vms.json"

BOOT_OK=true
if ! parallel \
        --results "${BOOT_LOG}" \
        --jobs 3 \
        bash -x ./bin/scenario.sh create {} \
        ::: "${SCENARIO_SOURCES}"/*.sh ; then
    BOOT_OK=false
fi

echo "===================================="
echo "System information after booting VMs"
echo "===================================="
free -h
df -h
sudo du -sk "${IMAGEDIR}"/* | sort -n
sudo virsh list --all
echo "===================================="

if ! ${BOOT_OK}; then
    echo "Failed to boot all VMs"
    exit 1
fi

echo "Boot phase complete"
