#!/bin/bash
#
# This script should be run on the image build server to
# create a local RPM repository containing the RPMs built
# from the source under test.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

create_repo() {
    local -r repodir="$1"
    createrepo "${repodir}"

    echo "Fixing permissions of RPM repo contents"
    find "${repodir}" -type f -exec sudo chmod a+r  {} \;
    find "${repodir}" -type d -exec sudo chmod a+rx {} \;
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
    # shellcheck disable=SC2086  # no quotes for command arguments to allow word splitting
    cp -R ${builddir}/{RPMS,SPECS,SRPMS} "${repodir}/"

    # Create the repository
    create_repo "${repodir}"
}

download_repo() {
    local -r repodir="$1"
    local -r repover="$2"

    local -r osversion="$(awk -F: '{print $5}' /etc/system-release-cpe)"
    local -r ocp_repo_name="rhocp-${repover}-for-rhel-${osversion}-$(uname -m)-rpms"

    title "Creating RPM repo at ${repodir}"

    # Start from scratch
    rm -rf "${repodir}" 2>/dev/null || true

    mkdir -p "${repodir}"
    pushd "${repodir}" &>/dev/null

    # Download MicroShift RPMs
    sudo dnf download --enablerepo="${ocp_repo_name}" microshift\*

    # Exit if no RPM packages were found
    if [ "$(find "${repodir}" -name '*.rpm' | wc -l)" -eq 0 ] ; then
        echo "No MicroShift RPM packages were found at the '${ocp_repo_name}' repository. Exiting..."
        exit 1
    fi

    # Create the repository
    create_repo "${repodir}"
    popd &>/dev/null
}

mkdir -p "${IMAGEDIR}"
cd "${IMAGEDIR}"

make_repo "${LOCAL_REPO}" "${RPM_SOURCE}"
make_repo "${NEXT_REPO}" "${NEXT_RPM_SOURCE}"
make_repo "${YPLUS2_REPO}" "${YPLUS2_RPM_SOURCE}"
make_repo "${BASE_REPO}" "${BASE_RPM_SOURCE}"

download_repo "${YMINUS1_REPO}" "4.$(previous_minor_version)"
