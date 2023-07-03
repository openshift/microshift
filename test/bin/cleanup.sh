#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

"${SCRIPTDIR}/composer_cleanup.sh"

for scenario in scenarios/*.sh; do
    ./bin/scenario.sh cleanup "${scenario}"
done

pkill caddy || true

rm -rf "${IMAGEDIR}"
