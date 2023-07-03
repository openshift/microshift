#!/bin/bash
#
# This script should be run on the image build server (usually the
# hypervisor) to create a local RPM repository containing the RPMs
# built from the source under test.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

mkdir -p "${IMAGEDIR}"
cd "${IMAGEDIR}"

if [ -d "${LOCAL_REPO}" ]; then
    echo "Cleaning up existing repository"
    rm -rf "${LOCAL_REPO}"
fi
mkdir -p "${LOCAL_REPO}"

# Create the local RPM repository for whatever was built from source.
echo "Copying RPMs from ${RPM_SOURCE} to ${LOCAL_REPO}"
# shellcheck disable=SC2086  # no quotes for command arguments to allow word splitting
cp -R ${RPM_SOURCE}/{RPMS,SPECS,SRPMS} "${LOCAL_REPO}/"

echo "Creating RPM repo at ${LOCAL_REPO}"
createrepo "${LOCAL_REPO}"

echo "Fixing permissions of RPM repo contents"
find "${LOCAL_REPO}" -type f -exec chmod a+r  {} \;
find "${LOCAL_REPO}" -type d -exec chmod a+rx {} \;
