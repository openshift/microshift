#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="${SCRIPTDIR}/../../.."
OUTPUTDIR="${ROOTDIR}/_output"
AWSCLI="${OUTPUTDIR}/bin/aws"

export AWS_PROFILE=microshift-ci
export AWS_BUCKET_NAME=microshift-footprint-and-performance

if [ ! -e "${AWSCLI}" ] ; then
    "${ROOTDIR}/scripts/fetch_tools.sh" awscli
fi

if ! "${AWSCLI}" s3 ls "${AWS_BUCKET_NAME}" &>/dev/null ; then
    echo "ERROR: Cannot access the '${AWS_BUCKET_NAME}' AWS bucket"
    exit 1
fi

get_build_branch() {
    local -r ocp_ver="$(grep ^OCP_VERSION "${ROOTDIR}/Makefile.version.$(uname -m).var"  | awk '{print $NF}' | awk -F. '{print $1"."$2}')"
    local -r cur_branch="$(git branch --show-current 2>/dev/null)"

    # Check if the current branch is derived from "main"
    local -r main_top=$(git rev-parse main 2>/dev/null)
    local -r main_base="$(git merge-base "${cur_branch}" main 2>/dev/null)"
    if [ "${main_top}" = "${main_base}" ] ; then
        echo "main"
        return
    fi

    # Check if the current branch is derived from "release-${ocp-ver}"
    local -r rel_top=$(git rev-parse "release-${ocp_ver}" 2>/dev/null)
    local -r rel_base="$(git merge-base "${cur_branch}" "release-${ocp_ver}" 2>/dev/null)"
    if [ "${rel_top}" = "${rel_base}" ] ; then
        echo "release-${ocp_ver}"
        return
    fi

    # Fallback to main if none of the above works
    echo "main"
}

SCENARIO_BUILD_BRANCH="$(get_build_branch)"

"${AWSCLI}" s3 sync "${ARTIFACTS_DIR}" "s3://${AWS_BUCKET_NAME}/${SCENARIO_BUILD_BRANCH}/"
