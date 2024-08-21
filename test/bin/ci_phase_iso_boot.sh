#!/bin/bash
#
# This script runs on the hypervisor to boot all scenario VMs.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# create empty scenario-info directory
mkdir "${SCENARIO_INFO_DIR}"

echo "This script's functionality has been moved. The script will be removed once the necessary changes are implemented in the release repository."

exit 0


