#!/bin/bash
set -euo pipefail

if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    echo "This script must be sourced, not executed."
    exit 1
fi

export LATEST_RELEASE_TYPE="ec"
export LATEST_RELEASE="4.20.0-ec.5"

# shellcheck disable=SC2034
declare -a PREV_RELEASES=(
    "4.20.0-ec.4"
    "4.20.0-ec.3"
)
