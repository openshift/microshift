#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
export SCRIPTDIR
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"
# shellcheck source=test/bin/common_versions.sh
source "${SCRIPTDIR}/common_versions.sh"

python3 "${SCRIPTDIR}/pyutils/build_bootc_images.py" "$@"
