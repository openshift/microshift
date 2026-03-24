#!/bin/bash

# Following script queries RHOCP repositories to get latest available
# for version of MicroShift from current branch.
# It expect system to be registered (i.e. entitlement exists).
#
# We cannot use branch version (or sometimes even previous minor one)
# because repositories are only usable after the release.
# Accessing them before the release results in 403 error.
#
# Script can accept:
# - A minor version (e.g. 16) to check for RHOCP of specific version only.
# - A minor and major version (e.g. 22 4) to check for RHOCP of a specific major.minor.
#
# Output is:
# - just a minor version in case of subscription RHOCP repository, e.g.: 15
# - or an URL to beta mirror followed by comma and minor version, e.g.:
#   https://mirror.openshift.com/pub/openshift-v4/x86_64/dependencies/rpms/4.16-el9-beta/,16
#   https://mirror.openshift.com/pub/openshift-v5/x86_64/dependencies/rpms/5.0-el9-beta/,0

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
REPOROOT="$(cd "${SCRIPTDIR}/.." && pwd)"

if ! sudo subscription-manager status >&/dev/null; then
    >&2 echo "System must be subscribed"
    exit 1
fi

# Get version of currently checked out branch.
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
if [[ "$#" -ge 2 ]]; then
    # Both minor and major provided as arguments
    current_minor="${1}"
    current_major="${2}"
    stop="${current_minor}"
elif [[ "$#" -eq 1 ]]; then
    # Only minor provided, get major from Makefile
    current_minor="${1}"
    current_major=$(grep '^OCP_VERSION' "${make_version}" | cut -d'=' -f2 | tr -d ' ' | cut -d'.' -f1)
    stop="${current_minor}"
else
    # No arguments, get both from Makefile
    current_major=$(grep '^OCP_VERSION' "${make_version}" | cut -d'=' -f2 | tr -d ' ' | cut -d'.' -f1)
    current_minor=$(grep '^OCP_VERSION' "${make_version}" | cut -d'=' -f2 | tr -d ' ' | cut -d'.' -f2)
    stop=$(( current_minor - 3 ))
    if (( stop < 0 )); then
        stop=0
    fi
fi

# Go through minor versions, starting from current_minor counting down
# to get latest available rhocp repository.
# For example, if current version is 4.16, the code will try to access
# rhocp-4.15 (which may not be released yet) and then rhocp-4.14 (which
# will be returned if it's usable). Works similarly for version 5.x.
for ver in $(seq "${current_minor}" -1 "${stop}"); do
    repository="rhocp-${current_major}.${ver}-for-rhel-9-$(uname -m)-rpms"
    if sudo dnf repository-packages --showduplicates "${repository}" info cri-o 1>&2; then
        echo "${ver}"
        exit 0
    fi

    rhocp_beta_url="https://mirror.openshift.com/pub/openshift-v${current_major}/$(uname -m)/dependencies/rpms/${current_major}.${ver}-el9-beta/"
    if sudo dnf repository-packages --showduplicates --disablerepo '*' --repofrompath "this,${rhocp_beta_url}" this info cri-o 1>&2; then
        echo "${rhocp_beta_url},${ver}"
        exit 0
    fi
done

>&2 echo "Failed to get latest rhocp repository!"
exit 1
