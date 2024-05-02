#!/bin/bash

# Following script queries RHOCP repositories to get latest available
# for version of MicroShift from current branch.
# It expect system to be registered (i.e. entitlement exists).
#
# We cannot use branch version (or sometimes even previous minor one)
# because repositories are only usable after the release.
# Accessing them before the release results in 403 error.
#
# Output is just the minor (Y) version number.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPOROOT="$(cd "${SCRIPTDIR}/.." && pwd)"

if ! sudo subscription-manager status >&/dev/null; then
    >&2 echo "System must be subscribed"
    exit 1
fi

# Get minor version of currently checked out branch.
# It's based on values stored in Makefile.version.$ARCH.var.
make_version="${REPOROOT}/Makefile.version.$(uname -m).var"
if [ ! -f "${make_version}" ] ; then
    # Attempt to locate the Makefile version file next to the current script.
    # This is necessary when bootstrapping the development environment for the first time.
    make_version=$(find "${SCRIPTDIR}" -maxdepth 1 -name "Makefile.version.$(uname -m).*.var" | tail -1)
    if [ ! -f "${make_version}" ] ; then
        >&2 echo "Cannot find the Makefile.version.$(uname -m).var file"
        exit 1
    fi
fi
current_minor=$(cut -d'.' -f2 "${make_version}")
stop=$(( current_minor - 3 ))

# Go through minor versions, starting from current_mirror counting down
# to get latest available rhocp repository.
# For example, at the time of writing this comment, current_minor is 16,
# and following code will try to access rhocp-4.15 (which is not released yet)
# and then rhocp-4.14 (which will be returned from the script because it's usable).
for ver in $(seq "${current_minor}" -1 "${stop}"); do
    repository="rhocp-4.${ver}-for-rhel-9-$(uname -m)-rpms"
    if sudo dnf -v repository-packages "${repository}" info cri-o 1>&2;
    then
        echo "${ver}"
        exit 0
    fi
done

>&2 echo "Failed to get latest rhocp repository!"
exit 1
