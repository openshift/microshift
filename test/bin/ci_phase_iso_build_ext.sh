#!/bin/bash
#
# This script runs on the build host to create all external layer artifacts.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

# Cannot use common.sh yet because some dependencies may be missing,
# but we only need ROOTDIR at this time.
SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
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

# Allow for a dry-run option to save on testing time
BUILD_DRY_RUN=${BUILD_DRY_RUN:-false}
dry_run() {
    ${BUILD_DRY_RUN} && echo "echo"
}

# Try downloading the 'last' build cache.
# Return 0 on success or 1 otherwise.
download_build_cache() {
    local -r cache_last="$(\
        ./bin/manage_build_cache.sh getlast \
            -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}" | \
            awk '/LAST:/ {print $NF}' \
        )"

    if ./bin/manage_build_cache.sh verify -b "${SCENARIO_BUILD_BRANCH}" -t "${cache_last}" ; then
        # Download the cached images
        ./bin/manage_build_cache.sh download -b "${SCENARIO_BUILD_BRANCH}" -t "${cache_last}"
        return 0
    fi
    return 1
}

cat /etc/os-release

# Show what other dnf commands have been run to try to debug why we
# sometimes see cache collisons.
$(dry_run) sudo dnf history --reverse

cd "${ROOTDIR}"

# Get firewalld and repos in place. Use scripts to get the right repos
# for each branch.
$(dry_run) bash -x ./scripts/devenv-builder/configure-vm.sh --no-build --force-firewall "${PULL_SECRET}"
$(dry_run) bash -x ./scripts/image-builder/configure.sh

cd "${ROOTDIR}/test/"

# Source common.sh only after all dependencies are installed.
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# Check that the external directory exists and it is not empty
if [ ! -d "${EXTERNAL_RPM_SOURCE}" ] ; then
    echo "Download externally built RPMs to '${EXTERNAL_RPM_SOURCE}'"
    exit 1
fi
EXT_RPMS=$(find "${EXTERNAL_RPM_SOURCE}" -name \*.rpm | wc -l)
if [ "${EXT_RPMS}" -eq 0 ] ; then
    echo "Download externally built RPMs to '${EXTERNAL_RPM_SOURCE}'"
    exit 1
fi

# Set up for scenario tests
$(dry_run) bash -x ./bin/create_local_repo.sh

# Start the web server to host the ostree commit repository for parent images
$(dry_run) bash -x ./bin/start_webserver.sh

# Figure out an optimal number of osbuild workers
CPU_CORES="$(grep -c ^processor /proc/cpuinfo)"
MAX_WORKERS=$(find "${ROOTDIR}/test/image-blueprints" -name \*.toml | wc -l)
CUR_WORKERS="$( [ "${CPU_CORES}" -lt  $(( MAX_WORKERS * 2 )) ] && echo $(( CPU_CORES / 2 )) || echo "${MAX_WORKERS}" )"

$(dry_run) bash -x ./bin/start_osbuild_workers.sh "${CUR_WORKERS}"

# Check if cache can be used for builds
# This may fail when AWS S3 connection is not configured, or there is no cache bucket
if ! ./bin/manage_build_cache.sh getlast -b "${SCENARIO_BUILD_BRANCH}" -t "${SCENARIO_BUILD_TAG}" ; then
    echo "ERROR: Cannot access build cache"
    exit 1
fi
if ! download_build_cache ; then
    echo "ERROR: Cannot download build cache"
    exit 1
fi

# Build the external layer
$(dry_run) bash -x ./bin/build_images.sh -l ./image-blueprints/layer4-external

echo "Build phase complete"
exit 0
