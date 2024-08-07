#!/bin/bash

set -xeuo pipefail
export PS4='\n+ $(date "+%T.%N") ${BASH_SOURCE}:$LINENO \011'

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TEST_TIME=$(date +'%Y-%m-%d_%T.%N')
export TEST_TIME

# Main directory to sync with root of s3 bucket: /tmp/TMP_DIR -> s3://microshift-footprint-and-performance/
ARTIFACTS_DIR="$(mktemp -d)"
export ARTIFACTS_DIR

# Subdir for low-latency artifacts. Ends up in s3 as s3://microshift-footprint-and-performance/${BRANCH}/low-latency/${TEST_TIME}
LOW_LAT_ARTIFACTS="${ARTIFACTS_DIR}/low-latency/${TEST_TIME}"
export LOW_LAT_ARTIFACTS
mkdir -p "${LOW_LAT_ARTIFACTS}"

for f in "${SCRIPTDIR}"/tests/[0-9]*.sh; do
    if ! "${f}"; then
        echo "ERROR: ${f} failed"
        exit 1
    fi
done
