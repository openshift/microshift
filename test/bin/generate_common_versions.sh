#!/bin/bash

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export SCRIPTDIR

python3 "${SCRIPTDIR}/pyutils/generate_common_versions.py" "$@" > "${SCRIPTDIR}/common_versions.sh"