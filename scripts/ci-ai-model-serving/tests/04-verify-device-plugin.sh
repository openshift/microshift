#!/usr/bin/env bash

set -xeuo pipefail

sudo microshift healthcheck \
    --namespace nvidia-device-plugin \
    --daemonsets nvidia-device-plugin-daemonset

retries=20
while [ ${retries} -gt 0 ] ; do
    ((retries-=1))

    capacity=$(oc get node -o=jsonpath='{ .items[0].status.capacity }')
    gpus=$(echo "${capacity}" | jq -r '."nvidia.com/gpu"')

    if [[ "${gpus}" != "null" ]]; then
        exit 0
    fi

    sleep 5
done

echo "Node's capacity does not include NVIDIA GPU"
exit 1
