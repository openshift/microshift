#!/bin/bash
#
# This script cleans up osbuild-composer. It cancels any running
# builds, deletes failed and completed builds, and removes package
# sources other than the defaults.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# Clean up the composer cache
"${ROOTDIR}/scripts/image-builder/cleanup.sh" -full
