#!/bin/bash

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export SCRIPTDIR

usage() {
    cat - <<EOF
generate_common_versions.sh minor

    minor -- MicroShift minor version, e.g. "21"
EOF
}

if [ "$#" -ne 1 ]; then
    usage
    exit 1
fi

python3 "${SCRIPTDIR}/pyutils/generate_common_versions.py" "$@" > "${SCRIPTDIR}/common_versions.sh"