#!/bin/bash
#
# This script runs on the hypervisor, from the iso-build step.

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Cannot use common.sh because virsh is not installed, but we only
# need ROOTDIR to set up logging in this script.
ROOTDIR="$(cd "${SCRIPTDIR}/../.." && pwd)"

# Log output automatically
LOGDIR="${ROOTDIR}/_output/ci-logs"
LOGFILE="${LOGDIR}/$(basename "$0" .sh).log"
if [ ! -d "${LOGDIR}" ]; then
    mkdir -p "${LOGDIR}"
fi
echo "Logging to ${LOGFILE}"
# Set fd 1 and 2 to write to the log file
exec &> >(tee >(awk '{ print strftime("%Y-%m-%d %H:%M:%S"), $0; fflush() }' >"${LOGFILE}"))

PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}

# Clean the dnf cache to avoid corruption
dnf clean all

# Show what other dnf commands have been run to try to debug why we
# sometimes see cache collisons.
dnf history --reverse

cd ~/microshift

# Get firewalld and repos in place. Use scripts to get the right repos
# for each branch.
bash -x ./scripts/devenv-builder/configure-vm.sh --no-build --force-firewall "${PULL_SECRET}"
bash -x ./scripts/image-builder/configure.sh

# Make sure libvirtd is running. We do this here, because some of the
# other scripts use virsh.
bash -x ./scripts/devenv-builder/manage-vm.sh config

# Fix up firewall so the VMs on their NAT network can talk to the
# server running on the hypervisor. We do this here, before creating
# any VMs, because the iptables rules added when the VMs are created
# are not persistent and are lost when this script reloads the
# firewall.
cd ~/microshift/test/
bash -x ./bin/configure_hypervisor_firewall.sh

# Re-build from source.
cd ~/microshift/
rm -rf ./_output/rpmbuild
make rpm

# Set up for scenario tests
cd ~/microshift/test/
timeout 20m bash -x ./bin/create_local_repo.sh
timeout 20m bash -x ./bin/start_osbuild_workers.sh 5
timeout 20m bash -x ./bin/build_images.sh
timeout 20m bash -x ./bin/download_images.sh

echo "Build phase complete"
