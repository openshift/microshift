#!/bin/bash
set -euo pipefail

# Following file is auto-generated using generate_common_versions.py.
# It should not be edited manually.

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script must be sourced, not executed."
    exit 1
fi

get_vrel_from_beta() {
    local -r beta_repo="$1"
    local -r beta_vrel=$(\
        dnf repoquery microshift \
            --quiet \
            --queryformat '%{version}-%{release}' \
            --disablerepo '*' \
            --repofrompath "this,${beta_repo}" \
            --latest-limit 1 2>/dev/null \
        )
    if [ -n "${beta_vrel}" ]; then
        echo "${beta_vrel}"
        return
    fi
    echo ""
}

get_vrel_from_rhsm() {
    local -r rhsm_repo="$1"
    local -r rhsm_vrel=$(\
        dnf repoquery microshift \
            --quiet \
            --queryformat '%{version}-%{release}' \
            --repo "${rhsm_repo}" \
            --latest-limit 1 2>/dev/null \
        )
    if [ -n "${rhsm_vrel}" ]; then
        echo "${rhsm_vrel}"
        return
    fi
    echo ""
}

get_vrel_from_rpm() {
    local -r rpm_dir="$1"

    # exit if rpm_dir directory does not exist
    if [ ! -d "${rpm_dir}" ]; then
        echo ""
        return
    fi

    local -r rpm_release_info_file=$(find "${rpm_dir}" -name "microshift-release-info-*.rpm" | sort | tail -n1)
    if [ -z "${rpm_release_info_file}" ]; then
        echo ""
        return
    fi

    local -r rpm_vrel=$(\
        rpm -qp --queryformat '%{version}-%{release}' \
            -p "${rpm_release_info_file}" 2>/dev/null \
        )
    if [ -n "${rpm_vrel}" ]; then
        echo "${rpm_vrel}"
        return
    fi
    echo ""
}

# The current release version (e.g. '4.17') affects
# the definition of previous and fake next versions.
export MAJOR_VERSION=5
export MINOR_VERSION=0
export PREVIOUS_MAJOR_VERSION=4
export PREVIOUS_MINOR_VERSION=22
export YMINUS2_MAJOR_VERSION=4
export YMINUS2_MINOR_VERSION=21
# Handle cross-major version boundary (e.g. 4.22 -> 5.0)
declare -A LAST_MINOR_FOR_MAJOR=([4]=22)
if [[ -n "${LAST_MINOR_FOR_MAJOR[${MAJOR_VERSION}]:-}" && \
      "${MINOR_VERSION}" -eq "${LAST_MINOR_FOR_MAJOR[${MAJOR_VERSION}]}" ]]; then
    export FAKE_NEXT_MAJOR_VERSION=$(( MAJOR_VERSION + 1 ))
    export FAKE_NEXT_MINOR_VERSION=0
else
    export FAKE_NEXT_MAJOR_VERSION="${MAJOR_VERSION}"
    export FAKE_NEXT_MINOR_VERSION=$(( "${MINOR_VERSION}" + 1 ))
fi

# For a main branch, the current release repository usually comes from
# the OpenShift mirror site, either 'ocp-dev-preview' in the beginning of the
# development cycle or 'ocp' when release candidates are built regularly.
#
# For a release branch, the current release repository should come from the
# official 'rhocp' stream.
CURRENT_RELEASE_REPO=""
CURRENT_RELEASE_VERSION=""
export CURRENT_RELEASE_REPO
export CURRENT_RELEASE_VERSION

# The previous release repository value should either point to the OpenShift
# mirror URL or the 'rhocp' repository name.
#
# For a main branch, the previous release repository may come from the official
# 'rhocp' stream or the OpenShift mirror. It is necessary to use the release
# candidate repository from the OpenShift mirror after a branch is created, but
# the previous release has not been made public yet.
#
# For a release branch, the previous release repository should come from the
# official 'rhocp' stream.# The previous release repository value should either
# point to the OpenShift mirror URL or the 'rhocp' repository name.
PREVIOUS_RELEASE_REPO="https://mirror.openshift.com/pub/openshift-v4/${UNAME_M}/microshift/ocp/latest-4.22/el9/os"
PREVIOUS_RELEASE_VERSION="$(get_vrel_from_beta "${PREVIOUS_RELEASE_REPO}")"
export PREVIOUS_RELEASE_REPO
export PREVIOUS_RELEASE_VERSION

# The y-2 release repository value should either point to the OpenShift
# mirror URL or the 'rhocp' repository name. It should always come from
# the 'rhocp' stream.
YMINUS2_RELEASE_REPO="rhocp-4.21-for-rhel-9-${UNAME_M}-rpms"
YMINUS2_RELEASE_VERSION="$(get_vrel_from_rhsm "${YMINUS2_RELEASE_REPO}")"
export YMINUS2_RELEASE_REPO
export YMINUS2_RELEASE_VERSION

# The 'rhocp_major_y' and 'rhocp_minor_y' variables should be the major and minor
# version numbers, if the current release is available through the 'rhocp' stream,
# otherwise empty.
RHOCP_MAJOR_Y=""
RHOCP_MINOR_Y=""
# The beta repository, containing dependencies, should point to the
# OpenShift mirror URL. If the mirror for current minor is not
# available yet, it should point to an older release.
RHOCP_MINOR_Y_BETA="https://mirror.openshift.com/pub/openshift-v4/${UNAME_M}/dependencies/rpms/4.22-el9-beta"
export RHOCP_MAJOR_Y
export RHOCP_MINOR_Y
export RHOCP_MINOR_Y_BETA

# The 'rhocp_major_y1' and 'rhocp_minor_y1' variables should be the previous major
# and minor version numbers, if the previous release is available through the
# 'rhocp' stream, otherwise empty.
RHOCP_MAJOR_Y1=""
RHOCP_MINOR_Y1=""
# The beta repository, containing dependencies, should point to the
# OpenShift mirror URL. The mirror for previous release should always
# be available.
RHOCP_MINOR_Y1_BETA="https://mirror.openshift.com/pub/openshift-v4/${UNAME_M}/dependencies/rpms/4.22-el9-beta"
export RHOCP_MAJOR_Y1
export RHOCP_MINOR_Y1
export RHOCP_MINOR_Y1_BETA

# The 'rhocp_major_y2' and 'rhocp_minor_y2' should always be the y-2 version numbers.
export RHOCP_MAJOR_Y2=4
export RHOCP_MINOR_Y2=21

export CNCF_SONOBUOY_VERSION=v0.57.3

# The version of systemd-logs image included in the sonobuoy release.
export CNCF_SYSTEMD_LOGS_VERSION=v0.4

# The current version of the microshift-gitops package.
export GITOPS_VERSION=1.19

# The brew release versions needed for release regression testing
BREW_Y0_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/${MAJOR_VERSION}.${MINOR_VERSION}-zstream/${UNAME_M}/")"
BREW_Y1_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/${PREVIOUS_MAJOR_VERSION}.${PREVIOUS_MINOR_VERSION}-zstream/${UNAME_M}/")"
BREW_Y2_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/${YMINUS2_MAJOR_VERSION}.${YMINUS2_MINOR_VERSION}-zstream/${UNAME_M}/")"
BREW_RC_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/${MAJOR_VERSION}.${MINOR_VERSION}-rc/${UNAME_M}/")"
BREW_EC_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/${MAJOR_VERSION}.${MINOR_VERSION}-ec/${UNAME_M}/")"
BREW_NIGHTLY_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/${MAJOR_VERSION}.${MINOR_VERSION}-nightly/${UNAME_M}/")"
export BREW_Y0_RELEASE_VERSION
export BREW_Y1_RELEASE_VERSION
export BREW_Y2_RELEASE_VERSION
export BREW_RC_RELEASE_VERSION
export BREW_EC_RELEASE_VERSION
export BREW_NIGHTLY_RELEASE_VERSION

# Set the release type based on priority: zstream > RC > EC > nightly
if [ -n "${BREW_Y0_RELEASE_VERSION}" ]; then
    BREW_LREL_RELEASE_VERSION="${BREW_Y0_RELEASE_VERSION}"
elif [ -n "${BREW_RC_RELEASE_VERSION}" ]; then
    BREW_LREL_RELEASE_VERSION="${BREW_RC_RELEASE_VERSION}"
elif [ -n "${BREW_EC_RELEASE_VERSION}" ]; then
    BREW_LREL_RELEASE_VERSION="${BREW_EC_RELEASE_VERSION}"
else
    BREW_LREL_RELEASE_VERSION="${BREW_NIGHTLY_RELEASE_VERSION}"
fi

export BREW_LREL_RELEASE_VERSION

# Branch and commit for the openshift-tests-private repository
OPENSHIFT_TESTS_PRIVATE_REPO_BRANCH="release-${MAJOR_VERSION}.${MINOR_VERSION}"
OPENSHIFT_TESTS_PRIVATE_REPO_COMMIT="b5111e366dc8f517732c6d48219ed659497de8e0"
export OPENSHIFT_TESTS_PRIVATE_REPO_BRANCH
export OPENSHIFT_TESTS_PRIVATE_REPO_COMMIT
