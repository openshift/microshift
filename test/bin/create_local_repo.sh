#!/bin/bash
#
# This script should be run on the image build server (usually the
# hypervisor) to create a local RPM repository containing the RPMs
# built from the source under test.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

make_repo() {
    local repodir="$1"
    local builddir="$2"

    # Skip repo creation for non-existent builds
    if [ ! -d "${builddir}" ] ; then
        echo "Skipping repo creation for '${builddir}'"
        return
    fi

    title "Creating RPM repo at ${repodir}"
    if [ -d "${repodir}" ]; then
        echo "Cleaning up existing repository"
        rm -rf "${repodir}"
    fi
    mkdir -p "${repodir}"

    # Create the local RPM repository for whatever was built from source or downloaded
    echo "Copying RPMs from ${builddir} to ${repodir}"
    if [ -d "${builddir}/RPMS" ] && [ -d "${builddir}/SPECS" ] && [ -d "${builddir}/SRPMS" ] ; then
        cp -R "${builddir}"/{RPMS,SPECS,SRPMS} "${repodir}/"
    else
        mkdir -p "${repodir}"
        find "${builddir}" -name \*.rpm -exec cp -f {} "${repodir}/" \;
    fi

    createrepo "${repodir}"

    echo "Fixing permissions of RPM repo contents"
    find "${repodir}" -type f -exec chmod a+r  {} \;
    find "${repodir}" -type d -exec chmod a+rx {} \;
}

mkdir -p "${IMAGEDIR}"
cd "${IMAGEDIR}"

make_repo "${LOCAL_REPO}" "${RPM_SOURCE}"
make_repo "${NEXT_REPO}" "${NEXT_RPM_SOURCE}"
make_repo "${YPLUS2_REPO}" "${YPLUS2_RPM_SOURCE}"
make_repo "${BASE_REPO}" "${BASE_RPM_SOURCE}"
make_repo "${EXTERNAL_REPO}" "${EXTERNAL_RPM_SOURCE}"
