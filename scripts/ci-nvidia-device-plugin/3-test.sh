#!/bin/bash

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE}:$LINENO \011'

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TEST_TIME=$(date +'%Y-%m-%d_%T.%N')
export TEST_TIME

for f in "${SCRIPTDIR}"/tests/[0-9]*.sh; do
    if ! "${f}"; then
        echo "ERROR: ${f} failed"
        exit 1
    fi
done
