#!/bin/bash

set -xeuo pipefail
export PS4='\n+ $(date "+%T.%N") ${BASH_SOURCE}:$LINENO \011'

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

for f in "${SCRIPTDIR}"/setup/[0-9]*.sh; do
    if ! "${f}"; then
        echo "ERROR: ${f} failed"
        exit 1
    fi
done
