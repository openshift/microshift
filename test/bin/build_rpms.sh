#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

build_rpms() {
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

    # Show progress for interactive mode when stdin is a terminal
    if [ -t 0 ]; then
        progress="--progress"
    else
        progress=""
    fi

    # Disable the GNU Parallel citation
    echo will cite | parallel --citation &>/dev/null
    # Run the commands in parallel
    echo "Starting parallel builds:"
    printf -- "  - %s\n" "${BUILD_CMDS[@]}"
    BUILD_OK=true
    if ! parallel \
        ${progress} \
        --results "${BUILD_RPMS_LOG}" \
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
}

make_repo() {
    local repodir="$1"
    local builddir="$2"

    title "Creating RPM repo at ${repodir}"
    if [ -d "${repodir}" ]; then
        echo "Cleaning up existing repository"
        rm -rf "${repodir}"
    fi
    mkdir -p "${repodir}"

    # Create the local RPM repository for whatever was built from source.
    echo "Copying RPMs from ${builddir} to ${repodir}"
    # shellcheck disable=SC2086  # no quotes for command arguments to allow word splitting
    cp -R ${builddir}/{RPMS,SPECS,SRPMS} "${repodir}/"

    createrepo "${repodir}"

    echo "Fixing permissions of RPM repo contents"
    find "${repodir}" -type f -exec chmod a+r  {} \;
    find "${repodir}" -type d -exec chmod a+rx {} \;
}

create_local_repo() {
    mkdir -p "${IMAGEDIR}"
    cd "${IMAGEDIR}"

    make_repo "${LOCAL_REPO}" "${RPM_SOURCE}"
    make_repo "${NEXT_REPO}" "${NEXT_RPM_SOURCE}"
    make_repo "${BASE_REPO}" "${BASE_RPM_SOURCE}"
    
    # Force recreation of dnf cache after rebuilding the repositories
    sudo dnf clean all
}


build_rpms

create_local_repo