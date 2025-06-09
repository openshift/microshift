#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Build the flannel RPM unless overridden explicitly
WITH_FLANNEL=${WITH_FLANNEL:-1}
# Build the upstream TopoLVM RPM unless overridden explicitly
WITH_TOPOLVM=${WITH_TOPOLVM:-1}

# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

build_rpms() {
    cd "${ROOTDIR}"
    echo "Cleaning up any old builds"
    rm -rf _output/rpmbuild*

    # Normal build of current branch from source
    local build_cmds=("make WITH_FLANNEL=${WITH_FLANNEL} WITH_TOPOLVM=${WITH_TOPOLVM} rpm")

    # In CI, build the current branch from source with the build tools using used by OCP
    if [ -v CI_JOB_NAME ]; then
        build_cmds=("make WITH_FLANNEL=${WITH_FLANNEL} WITH_TOPOLVM=${WITH_TOPOLVM} rpm-podman")
    fi

    build_cmds+=(
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
    local -r num_build_cmds="${#build_cmds[@]}"
    local -r build_rpms_log="${IMAGEDIR}/build_rpms.json"
    local -r build_rpms_jobs_log="${IMAGEDIR}/build_rpms_jobs.txt"
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
    printf -- "  - %s\n" "${build_cmds[@]}"
    local build_ok=true
    if ! parallel \
        ${progress} \
        --results "${build_rpms_log}" \
        --joblog "${build_rpms_jobs_log}" \
        --jobs "${num_build_cmds}" \
        ::: "${build_cmds[@]}" ; then
        build_ok=false
    fi

    # Show the summary of the output of the parallel jobs.
    cat "${build_rpms_jobs_log}"

    if [ -f "${build_rpms_log}" ] ; then
        jq < "${build_rpms_log}"
    else
        echo "The RPM build log file does not exist"
        build_ok=false
    fi

    if ! ${build_ok} ; then
        echo "RPM build failed"
        exit 1
    fi

    local -r max_runtime=$(jq -s 'max_by(.JobRuntime) | .JobRuntime | floor' "${build_rpms_log}")
    echo "RPM build complete in ${max_runtime}s"
}

make_repo() {
    local -r repodir="$1"
    local -r builddir="$2"

    title "Creating RPM repo at ${repodir}"
    if [ -d "${repodir}" ]; then
        echo "Cleaning up existing repository"
        rm -rf "${repodir}"
    fi
    mkdir -p "${repodir}"

    # Create the local RPM repository for whatever was built from source.
    echo "Copying RPMs from ${builddir} to ${repodir}"
    if [ -d "${builddir}" ] ; then
        find "${builddir}" -name \*.rpm -exec cp -R "{}" "${repodir}/" \;
    else
        echo "WARNING: The '${builddir}' directory does not exist, skipping"
    fi

    createrepo "${repodir}"

    echo "Fixing permissions of RPM repo contents"
    find "${repodir}" -type f -exec chmod a+r  {} \;
    find "${repodir}" -type d -exec chmod a+rx {} \;
}

# Note: Brew RPM download requires a VPN connection.
#
# The brew RPMs are part of the CI cache. If access to VPN is not available and
# the cached artifacts are not present locally, an empty brew source directory
# is created so that other procedures do not fail.
download_brew_rpms() {
    mkdir -p "${BREW_RPM_SOURCE}"
    if "${SCRIPTDIR}/manage_brew_rpms.sh" access ; then
        # shellcheck source=test/bin/common_versions.sh
        source "${SCRIPTDIR}/common_versions.sh"
        # Delete the old RPMs before the download
        echo "Cleaning up old brew downloads"
        rm -rf "${BREW_RPM_SOURCE}"
        # Run the download procedure
        bash -x "${SCRIPTDIR}/../../scripts/fetch_tools.sh" brew
        bash -x "${SCRIPTDIR}/manage_brew_rpms.sh" download "4.${MINOR_VERSION}" "${BREW_RPM_SOURCE}"
    else
        echo "WARNING: The Brew Hub site is not accessible, skipping the download"
    fi
}

create_local_repo() {
    mkdir -p "${IMAGEDIR}"
    cd "${IMAGEDIR}"

    make_repo "${LOCAL_REPO}" "${RPM_SOURCE}"
    make_repo "${NEXT_REPO}" "${NEXT_RPM_SOURCE}"
    make_repo "${BASE_REPO}" "${BASE_RPM_SOURCE}"
    make_repo "${BREW_REPO}" "${BREW_RPM_SOURCE}/${UNAME_M}"

    # Force recreation of dnf caches after rebuilding the repositories
    sudo dnf clean all
         dnf clean all
}

#
# Main
#
build_rpms
download_brew_rpms
create_local_repo
