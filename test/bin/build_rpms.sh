#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

cd "${ROOTDIR}"
echo "Cleaning up any old builds"
rm -rf _output/rpmbuild*

# Normal build of current branch from source
BUILD_CMDS=('make rpm')

# In CI, build the current branch from source with the build tools using used by OCP
if [ -v CI_JOB_NAME ]; then
    BUILD_CMDS=('make rpm-podman')
fi

BUILD_CMDS+=(
    # Build RPMs with the version number of the next minor release,
    # but using the same source code as the normal build.
    'make -C test/ fake-next-minor-rpm' \
    # Build RPMs with the version number of the y+2 minor release,
    # but using the same source code as the normal build.
    'make -C test/ fake-yplus2-minor-rpm' \
    # Build RPMs from release-$MAJOR.$MINOR of this repo.
    # These RPMs are useful in providing a layer to upgrade from.
    'make -C test/ build-base-branch' \
    # Build microshift-test-agent helping with creating unhealthy system scenarios
    # such as: microshift being unable to make a backup or greenboot checks failing
    './test/agent/build.sh' \
)
NUM_BUILD_CMDS="${#BUILD_CMDS[@]}"
BUILD_RPMS_LOG="${IMAGEDIR}/build_rpms.json"
BUILD_RPMS_JOB_LOG="${IMAGEDIR}/build_rpms_jobs.txt"
mkdir -p "${IMAGEDIR}"

# Disable the GNU Parallel citation
echo will cite | parallel --citation &>/dev/null
# Run the commands in parallel
# Avoid --progress option because it requires tty and
# it slows down execution in interactive shells
echo "Starting parallel builds:"
printf -- "  - %s\n" "${BUILD_CMDS[@]}"
BUILD_OK=true
if ! parallel --results "${BUILD_RPMS_LOG}" \
        --joblog "${BUILD_RPMS_JOB_LOG}" \
        --jobs "${NUM_BUILD_CMDS}" \
        ::: "${BUILD_CMDS[@]}" ; then
    BUILD_OK=false
fi

# Show the summary of the output of the parallel jobs.
cat "${BUILD_RPMS_JOB_LOG}"

if [ -f "${BUILD_RPMS_LOG}" ] ; then
    jq < "${BUILD_RPMS_LOG}"
else
    echo "The RPM build log file does not exist"
    BUILD_OK=false
fi

if ! ${BUILD_OK} ; then
    echo "RPM build failed"
    exit 1
fi

MAX_RUNTIME=$(jq -s 'max_by(.JobRuntime) | .JobRuntime | floor' "${BUILD_RPMS_LOG}")
echo "RPM build complete in ${MAX_RUNTIME}s"
