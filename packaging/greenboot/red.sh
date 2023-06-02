#!/usr/bin/env bash

set -ex

HEALTH=unhealthy

if [ ! -f /run/ostree-booted ]; then
    echo "System is not booted with ostree"
    exit 0
fi

mkdir -p /var/lib/microshift-backups

boot=$(tr -d '-' < /proc/sys/kernel/random/boot_id)
deploy=$(rpm-ostree status --booted --jsonpath='$.deployments[0].id' | jq -r '.[0]')
jq \
    --null-input \
    --arg health "${HEALTH}" \
    --arg deploy "${deploy}" \
    --arg boot "${boot}" \
    '{ "health": $health, "deployment_id": $deploy, "boot_id": $boot }' > /var/lib/microshift-backups/health.json
