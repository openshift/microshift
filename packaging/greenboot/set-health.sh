#!/usr/bin/env bash

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

if [[ "${SCRIPT_DIR}" == *"green.d"* ]]; then
    HEALTH="healthy"
else
    HEALTH="unhealthy"
fi

set -x
microshift admin health set-current "--${HEALTH}"

