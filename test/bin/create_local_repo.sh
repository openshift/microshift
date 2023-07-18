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
    find "${repodir}" -type f -print -exec chmod a+r  {} \;
    find "${repodir}" -type d -exec chmod a+rx {} \;
}

mkdir -p "${IMAGEDIR}"
cd "${IMAGEDIR}"

make_repo "${LOCAL_REPO}" "${RPM_SOURCE}"
make_repo "${NEXT_REPO}" "${NEXT_RPM_SOURCE}"
