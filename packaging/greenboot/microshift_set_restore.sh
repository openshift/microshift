#!/usr/bin/env bash

set -ex

ACTION=restore

SCRIPT_NAME=$(basename "$0")
if [ "$(id -u)" -ne 0 ]; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

if [ ! -f /run/ostree-booted ]; then
    echo "System is not booted with ostree"
    exit 0
fi

mkdir -p /var/lib/microshift-backups

boot=$(tr -d '-' </proc/sys/kernel/random/boot_id)
deploy=$(rpm-ostree status --booted --jsonpath='$.deployments[0].id' | jq -r '.[0]')
jq \
    --null-input \
    --arg action "${ACTION}" \
    --arg deploy "${deploy}" \
    --arg boot "${boot}" \
    '{ "action": $action, "deployment_id": $deploy, "boot_id": $boot }' >/var/lib/microshift-backups/system.json
