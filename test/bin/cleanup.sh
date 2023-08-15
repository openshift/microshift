#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

"${SCRIPTDIR}/composer_cleanup.sh"

for scenario in scenarios/*.sh; do
    ./bin/scenario.sh cleanup "${scenario}"
done

sudo pkill nginx || true

rm -rf "${IMAGEDIR}"
