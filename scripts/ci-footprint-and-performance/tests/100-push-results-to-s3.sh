#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
ROOTDIR="${SCRIPTDIR}/../../.."
OUTPUTDIR="${ROOTDIR}/_output"
AWSCLI="${OUTPUTDIR}/bin/aws"

export AWS_PROFILE=microshift-ci
export AWS_BUCKET_NAME=microshift-footprint-and-performance

JOB_TYPE=${JOB_TYPE:-}
if [[ "${JOB_TYPE}" != "periodic" ]]; then
    : Non periodic job was detected - skipping pushing results to S3
    exit 0
fi

if [ ! -e "${AWSCLI}" ] ; then
    "${ROOTDIR}/scripts/fetch_tools.sh" awscli
fi

if ! "${AWSCLI}" s3 ls "${AWS_BUCKET_NAME}" &>/dev/null ; then
    echo "ERROR: Cannot access the '${AWS_BUCKET_NAME}' AWS bucket"
    exit 1
fi

BRANCH=${BRANCH:-main}
"${AWSCLI}" s3 sync "${ARTIFACTS_DIR}" "s3://${AWS_BUCKET_NAME}/${BRANCH}/"
