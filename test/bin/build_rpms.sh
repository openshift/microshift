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

# Normal build of current branch from source
build_from_current_branch() {
    echo "Building from current branch"
    make rpm
}

# Build RPMs with the version number of the next minor release,
# but using the same source code as the normal build.
build_fake_next_minor() {
    echo "Building fake next minor version"
    make -C test/ fake-next-minor-rpm
}

# Build RPMs with the version number of the y+2 minor release,
# but using the same source code as the normal build.
build_fake_yplus2_minor() {
    echo "Building fake next next minor version"
    make -C test/ fake-yplus2-minor-rpm
}

# Build RPMs from release-$MAJOR.$MINOR of this repo.
# These RPMs are useful in providing a layer to upgrade from.
build_base_release_branch() {
    echo "Building base release branch"
    make -C test/ build-base-branch
}

# Build microshift-test-agent helping with creating unhealthy system scenarios
# such as: microshift being unable to make a backup or greenboot checks failing
build_test_agent() {
    echo "Building test agent"
    ./test/agent/build.sh
}

BUILD_CMDS=( \
    build_from_current_branch \
    build_fake_next_minor \
    build_fake_yplus2_minor \
    build_base_release_branch \
    build_test_agent \
)
NUM_BUILD_CMDS="${#BUILD_CMDS[@]}"
BUILD_RPMS_LOG="${IMAGEDIR}/build_rpms.json"
mkdir -p "${IMAGEDIR}"

# Export all the functions to be used by 'parallel'
for ((i=0; i<"${NUM_BUILD_CMDS}"; i++)); do
    export -f "${BUILD_CMDS[${i}]}"
done

# Disable the GNU Parallel citation
echo will cite | parallel --citation &>/dev/null
# Run the commands in parallel
# Avoid --progress option because it requires tty and
# it slows down execution in interactive shells
if ! parallel --results "${BUILD_RPMS_LOG}" \
        --jobs "${NUM_BUILD_CMDS}" \
        bash -c '{1}' ::: "${BUILD_CMDS[@]}" ; then
    echo "Failed to run RPM build in parallel"
    exit 1
fi

if [ ! -f "${BUILD_RPMS_LOG}" ] ; then
    echo "The RPM build log file does not exist"
    exit 1
fi
jq < "${BUILD_RPMS_LOG}"

# Determine the script exit code
if [ "$(jq '.Exitval==0' "${BUILD_RPMS_LOG}" | grep -wc false)" != 0 ] ; then
    echo "Failed to build all RPMs"
    exit 1
fi

MAX_RUNTIME=$(jq -s 'max_by(.JobRuntime) | .JobRuntime | floor' "${BUILD_RPMS_LOG}")
echo "RPM build complete in ${MAX_RUNTIME}s"
