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

# Map of last minor version for each major (for cross-major transitions)
declare -A LAST_MINOR_FOR_MAJOR=([4]=22)

# Calculate previous version handling cross-major boundaries.
# Sets prev_major and prev_minor variables in the caller's scope.
get_prev_version() {
    local major=$1
    local minor=$2
    if (( minor > 0 )); then
        prev_major="${major}"
        prev_minor=$(( minor - 1 ))
    else
        prev_major=$(( major - 1 ))
        prev_minor="${LAST_MINOR_FOR_MAJOR[${prev_major}]:-}"
    fi
}

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
    max_steps=1
elif [[ "$#" -eq 1 ]]; then
    # Only minor provided, get major from Makefile
    current_minor="${1}"
    current_major=$(grep '^OCP_VERSION' "${make_version}" | cut -d'=' -f2 | tr -d ' ' | cut -d'.' -f1)
    max_steps=1
else
    # No arguments, get both from Makefile
    current_major=$(grep '^OCP_VERSION' "${make_version}" | cut -d'=' -f2 | tr -d ' ' | cut -d'.' -f1)
    current_minor=$(grep '^OCP_VERSION' "${make_version}" | cut -d'=' -f2 | tr -d ' ' | cut -d'.' -f2)
    max_steps=4
fi

# Go through versions, starting from current version counting down
# to get latest available rhocp repository. Handles cross-major
# boundaries (e.g. from 5.0 back to 4.22).
check_major="${current_major}"
check_minor="${current_minor}"
for (( step=0; step < max_steps; step++ )); do
    repository="rhocp-${check_major}.${check_minor}-for-rhel-9-$(uname -m)-rpms"
    if sudo dnf repository-packages --showduplicates "${repository}" info cri-o 1>&2; then
        echo "${check_minor}"
        exit 0
    fi

    rhocp_beta_url="https://mirror.openshift.com/pub/openshift-v${check_major}/$(uname -m)/dependencies/rpms/${check_major}.${check_minor}-el9-beta/"
    if sudo dnf repository-packages --showduplicates --disablerepo '*' --repofrompath "this,${rhocp_beta_url}" this info cri-o 1>&2; then
        echo "${rhocp_beta_url},${check_minor}"
        exit 0
    fi

    prev_major=""
    prev_minor=""
    get_prev_version "${check_major}" "${check_minor}"
    if [[ -z "${prev_minor}" ]]; then
        break
    fi
    check_major="${prev_major}"
    check_minor="${prev_minor}"
done

>&2 echo "Failed to get latest rhocp repository!"
exit 1
