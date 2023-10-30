#!/bin/bash
#
# This script runs on the hypervisor, from the iso-build step.

set -xeuo pipefail

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
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
sudo dnf clean all

# Show what other dnf commands have been run to try to debug why we
# sometimes see cache collisons.
sudo dnf history --reverse

cd "${ROOTDIR}"

# Get firewalld and repos in place. Use scripts to get the right repos
# for each branch.
bash -x ./scripts/devenv-builder/configure-vm.sh --no-build --force-firewall "${PULL_SECRET}"
bash -x ./scripts/image-builder/configure.sh

cd "${ROOTDIR}/test/"

# Re-build from source.
bash -x ./bin/build_rpms.sh

# Set up for scenario tests
bash -x ./bin/create_local_repo.sh

# Start the web server to host the ostree commit repository for parent images
bash -x ./bin/start_webserver.sh

# Build all of the images
CPU_CORES="$(grep -c ^processor /proc/cpuinfo)"
MAX_WORKERS=$(find "${ROOTDIR}/test/image-blueprints" -name \*.toml | wc -l)
CUR_WORKERS="$( [ "${CPU_CORES}" -lt  $(( MAX_WORKERS * 2 )) ] && echo $(( CPU_CORES / 2 )) || echo "${MAX_WORKERS}" )"

bash -x ./bin/start_osbuild_workers.sh "${CUR_WORKERS}"

# Image build can be optimized in CI based on the job type
if [ -v CI_JOB_NAME ] ; then
    bash -x ./bin/build_images.sh -g ./image-blueprints/group1
    bash -x ./bin/build_images.sh -g ./image-blueprints/group2
    # Group 3 only contains images used in periodic CI jobs
    if [[ "${CI_JOB_NAME}" =~ .*periodic.* ]]; then
        bash -x ./bin/build_images.sh -g ./image-blueprints/group3
    fi
else
    # Fall back to full build when not running in CI
    bash -x ./bin/build_images.sh
fi

echo "Build phase complete"
