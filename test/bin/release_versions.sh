#!/bin/bash
set -euo pipefail

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script must be sourced, not executed."
    exit 1
fi
export LATEST_RELEASE_TYPE="ec"

export LATEST_RELEASE="4.20.0-ec.5"
export LATEST_RPM_RELEASE="microshift-4.20.0~ec.5-202507311047.p0.g774f47d.assembly.ec.5.el9"
export LATEST_IMAGE_RELEASE="microshift-bootc-container-v4.19.0-202508070648.p2.g5726919.assembly.4.19.7.el9"

declare -a PREV_RELEASES=(
    "4.20.0-ec.4"
    "4.20.0-ec.3"
)

declare -a PREV_RPM_RELEASES=(
    "microshift-4.20.0~ec.4-202507300814.p0.gf96c363.assembly.ec.4.el9"
    "microshift-4.20.0~ec.3-202507081733.p0.g8ae81eb.assembly.ec.3.el9"
)

declare -a PREV_IMAGE_RELEASES=(
    "microshift-bootc-container-v4.19.0-202508070648.p2.g5726919.assembly.4.19.7.el9"
)
