#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).
set -euo pipefail

export SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

python3 "${SCRIPTDIR}/pyutils/build_bootc_images.py" "$@"
