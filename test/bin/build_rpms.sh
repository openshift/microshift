#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

cd "${ROOTDIR}"
rm -rf _output/rpmbuild*

BUILD_CMDS=( \
    # Normal build of current branch from source
    'make rpm-podman' \
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
mkdir -p "${IMAGEDIR}"

# Disable the GNU Parallel citation
echo will cite | parallel --citation &>/dev/null
# Run the commands in parallel
# Avoid --progress option because it requires tty and
# it slows down execution in interactive shells
BUILD_OK=true
if ! parallel --results "${BUILD_RPMS_LOG}" \
        --jobs "${NUM_BUILD_CMDS}" \
        ::: "${BUILD_CMDS[@]}" ; then
    BUILD_OK=false
fi

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
