#!/bin/bash

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE}:$LINENO \011'

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

for f in "${SCRIPTDIR}"/setup/[0-9]*.sh; do
    success=false
    for _ in $(seq 1 3); do
        if "${f}"; then
            success=true
            break
        fi
        sleep 5
    done

    if ! "${success}"; then
        echo "ERROR: ${f} failed"
        exit 1
    fi
done
