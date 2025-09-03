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

export MINOR_VERSION=20
export PREVIOUS_MINOR_VERSION=$(( "${MINOR_VERSION}" - 1 ))
export YMINUS2_MINOR_VERSION=$(( "${MINOR_VERSION}" - 2 ))
export FAKE_NEXT_MINOR_VERSION=$(( "${MINOR_VERSION}" + 1 ))

# For a main branch, the current release repository usually comes from
# the OpenShift mirror site, either 'ocp-dev-preview' in the beginning of the
# development cycle or 'ocp' when release candidates are built regularly.
#
# For a release branch, the current release repository should come from the
# official 'rhocp' stream.
CURRENT_RELEASE_REPO="" # "https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/microshift/ocp/latest-4.21/el9/os"
CURRENT_RELEASE_VERSION="" # "$(get_vrel_from_beta "${CURRENT_RELEASE_REPO}")"
export CURRENT_RELEASE_REPO
export CURRENT_RELEASE_VERSION

# The previous release repository value should either point to the OpenShift
# mirror URL or the 'rhocp' repository name.
#
# For a main branch, the previous release repository may come from the
# official 'rhocp' stream or the OpenShift mirror. It is necessary to use the
# release candidate repository from the OpenShift mirror after a branch is
# created, but the previous release has not been made public yet.
#
# For a release branch, the previous release repository should come from the
# official 'rhocp' stream.
PREVIOUS_RELEASE_REPO="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/microshift/ocp/latest-4.20/el9/os"
PREVIOUS_RELEASE_VERSION="$(get_vrel_from_beta "${PREVIOUS_RELEASE_REPO}")"
export PREVIOUS_RELEASE_REPO
export PREVIOUS_RELEASE_VERSION

# The y-2 release repository value should always contain the 'rhocp' repository name.
YMINUS2_RELEASE_REPO="rhocp-4.19-for-rhel-9-$(uname -m)-rpms"
YMINUS2_RELEASE_VERSION="$(get_vrel_from_rhsm "${YMINUS2_RELEASE_REPO}")"
export YMINUS2_RELEASE_REPO
export YMINUS2_RELEASE_VERSION

RHOCP_MINOR_Y=""
RHOCP_MINOR_Y_BETA="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/4.21-el9-beta/"
export RHOCP_MINOR_Y
export RHOCP_MINOR_Y_BETA

# Define a release version and/or the OpenShift mirror beta repository URL.
# If the release version is defined, the repository should be deduced from the
# PREVIOUS_RELEASE_REPO setting.
# Beta repository URL needs to be set for CentOS images as they don't have access to the RHOCP.
RHOCP_MINOR_Y1=""
RHOCP_MINOR_Y1_BETA="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/4.20-el9-beta/"
export RHOCP_MINOR_Y1
export RHOCP_MINOR_Y1_BETA

# Define a release version as it is not expected to use the OpenShift mirror
# for the y-2 release.
export RHOCP_MINOR_Y2=19

export CNCF_SONOBUOY_VERSION=v0.57.3

# The version of systemd-logs image included in the sonobuoy release.
export CNCF_SYSTEMD_LOGS_VERSION=v0.4

# The current version of the microshift-gitops package.
export GITOPS_VERSION=1.16

# The brew release versions needed for release regression testing
BREW_Y0_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/4.${MINOR_VERSION}-zstream/${UNAME_M}/")"
BREW_Y1_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/4.${PREVIOUS_MINOR_VERSION}-zstream/${UNAME_M}/")"
BREW_Y2_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/4.${YMINUS2_MINOR_VERSION}-zstream/${UNAME_M}/")"
BREW_RC_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/4.${MINOR_VERSION}-rc/${UNAME_M}/")"
BREW_EC_RELEASE_VERSION="$(get_vrel_from_rpm "${BREW_RPM_SOURCE}/4.${MINOR_VERSION}-ec/${UNAME_M}/")"
export BREW_Y0_RELEASE_VERSION
export BREW_Y1_RELEASE_VERSION
export BREW_Y2_RELEASE_VERSION
export BREW_RC_RELEASE_VERSION
export BREW_EC_RELEASE_VERSION

LATEST_RELEASE_TYPE="ec"
export LATEST_RELEASE_TYPE

BREW_LREL_RELEASE_VERSION="${BREW_EC_RELEASE_VERSION}"
export BREW_LREL_RELEASE_VERSION

# Branch and commit for the openshift-tests-private repository
OPENSHIFT_TESTS_PRIVATE_REPO_BRANCH="release-4.${MINOR_VERSION}"
OPENSHIFT_TESTS_PRIVATE_REPO_COMMIT="61613d96c91db7b2907c24dd257075d3f2201991"
export OPENSHIFT_TESTS_PRIVATE_REPO_BRANCH
export OPENSHIFT_TESTS_PRIVATE_REPO_COMMIT
