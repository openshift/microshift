#!/usr/bin/env bash

set -ex

HEALTH=unhealthy

SCRIPT_NAME=$(basename "$0")
DIR=/var/lib/microshift-backups
FILE="${DIR}/health.json"

if [ "$(id -u)" -ne 0 ]; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

if [ ! -f /run/ostree-booted ]; then
    echo "System is not booted with ostree"
    exit 0
fi

# Define var FORCE to force overwrite
if [[ ! -v FORCE ]] && [[ -e "${FILE}" ]]; then
    h=$(jq -r '.health' "${FILE}")
    b=$(jq -r '.boot_id' "${FILE}")
    d=$(jq -r '.deployment_id' "${FILE}")

    expected_backup="${DIR}/${d}_${b}"
    if [[ "${h}" == "healthy" ]] && [[ ! -e "${expected_backup}" ]]; then
        echo "State 'healthy' is persisted, but the backup was not created."
        echo "Not overwriting the health.json, will retry backup on next boot"
        exit 0
    fi
fi

mkdir -p /var/lib/microshift-backups

boot=$(tr -d '-' </proc/sys/kernel/random/boot_id)
deploy=$(rpm-ostree status --booted --json | jq -r '.deployments[0].id')
jq \
    --null-input \
    --arg health "${HEALTH}" \
    --arg deploy "${deploy}" \
    --arg boot "${boot}" \
    '{ "health": $health, "deployment_id": $deploy, "boot_id": $boot }' >"${FILE}"
