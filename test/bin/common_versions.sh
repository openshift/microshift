#!/bin/bash
set -euo pipefail

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

export MINOR_VERSION=20
export PREVIOUS_MINOR_VERSION=$(( "${MINOR_VERSION}" - 1 ))
export YMINUS2_MINOR_VERSION=$(( "${MINOR_VERSION}" - 2 ))
export FAKE_NEXT_MINOR_VERSION=$(( "${MINOR_VERSION}" + 1 ))

CURRENT_RELEASE_REPO=""
CURRENT_RELEASE_VERSION=""
export CURRENT_RELEASE_REPO
export CURRENT_RELEASE_VERSION

PREVIOUS_RELEASE_REPO="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/microshift/ocp/latest-4.19/el9/os"
PREVIOUS_RELEASE_VERSION="$(get_vrel_from_beta "${PREVIOUS_RELEASE_REPO}")"
export PREVIOUS_RELEASE_REPO
export PREVIOUS_RELEASE_VERSION

YMINUS2_RELEASE_REPO="rhocp-4.18-for-rhel-9-$(uname -m)-rpms"
YMINUS2_RELEASE_VERSION="$(get_vrel_from_rhsm "${YMINUS2_RELEASE_REPO}")"
export YMINUS2_RELEASE_REPO
export YMINUS2_RELEASE_VERSION

RHOCP_MINOR_Y=""
RHOCP_MINOR_Y_BETA="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/4.20-el9-beta"
export RHOCP_MINOR_Y
export RHOCP_MINOR_Y_BETA

RHOCP_MINOR_Y1=""
RHOCP_MINOR_Y1_BETA="https://mirror.openshift.com/pub/openshift-v4/$(uname -m)/dependencies/rpms/4.19-el9-beta"
export RHOCP_MINOR_Y1
export RHOCP_MINOR_Y1_BETA

export RHOCP_MINOR_Y2=18

export CNCF_SONOBUOY_VERSION=v0.57.3

