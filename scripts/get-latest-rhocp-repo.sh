#!/bin/bash

# Following script queries RHOCP repositories to get latest available
# for version of MicroShift from current branch.
# It expect system to be registered (i.e. entitlement exists).
#
# We cannot use branch version (or sometimes even previous minor one)
# because repositories are only usable after the release.
# Accessing them before the release results in 403 error.
#
# Script can accept a minor version (e.g. 16) to check for RHOCP of specific version only.
#
# Output is:
# - just a minor version in case of subscription RHOCP repository, e.g.: 15
# - or an URL to beta mirror followed by comma and minor version, e.g.:
#   https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rpms/4.16-el9-beta/,16

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPOROOT="$(cd "${SCRIPTDIR}/.." && pwd)"

if ! sudo subscription-manager status >&/dev/null; then
    >&2 echo "System must be subscribed"
    exit 1
fi

if [[ "$#" -eq 1 ]]; then
    current_minor="${1}"
    stop="${current_minor}"
else
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
fi

# Go through minor versions, starting from current_mirror counting down
# to get latest available rhocp repository.
# For example, at the time of writing this comment, current_minor is 16,
# and following code will try to access rhocp-4.15 (which is not released yet)
# and then rhocp-4.14 (which will be returned from the script because it's usable).
for ver in $(seq "${current_minor}" -1 "${stop}"); do
    repository="rhocp-4.${ver}-for-rhel-9-$(uname -m)-rpms"
    if sudo dnf repository-packages --showduplicates "${repository}" info cri-o 1>&2; then
        echo "${ver}"
        exit 0
    fi

    rhocp_beta_url="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/4.${ver}-el9-beta/"
    if sudo dnf repository-packages --showduplicates --disablerepo '*' --repofrompath "this,${rhocp_beta_url}" this info cri-o 1>&2; then
        echo "${rhocp_beta_url},${ver}"
        exit 0
    fi
done

>&2 echo "Failed to get latest rhocp repository!"
exit 1
