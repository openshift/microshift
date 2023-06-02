#!/usr/bin/env bash

set -ex

HEALTH=healthy

if [ ! -f /run/ostree-booted ]; then
    echo "System is not booted with ostree"
    exit 0
fi

mkdir -p /var/lib/microshift-backups

ID=$(rpm-ostree status --booted --jsonpath='$.deployments[0].id' | jq -r '.[0]')
jq \
    --null-input \
    --arg health "${HEALTH}" \
    --arg id "${ID}" \
    '{ "health": $health, "deployment_id": $id }' > /var/lib/microshift-backups/health.json
