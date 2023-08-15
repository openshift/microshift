#!/bin/bash
#
# This script cleans up all scenario infrastructure, reverts the
# hypervisor configuration and kills the web server process.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# Clean up all of the VMs
for scenario in "${SCENARIO_SOURCES}"/*.sh; do
    echo "Deleting $(basename "${scenario}")"
    "${TESTDIR}/bin/scenario.sh" cleanup "${scenario}" &>/dev/null || true
done

# Clean up the hypervisor configuration for the tests
"${TESTDIR}/bin/manage_hypervisor_config.sh" cleanup

sudo pkill nginx || true
