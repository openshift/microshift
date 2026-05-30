#!/usr/bin/env bash

set -xeuo pipefail

retries=30
while [ ${retries} -gt 0 ] ; do
    ((retries-=1))
    if oc get daemonset/nvidia-device-plugin-daemonset -n nvidia-device-plugin &>/dev/null; then
        break
    fi
    echo "Waiting for nvidia-device-plugin daemonset to be created... (${retries} retries remaining)"
    sleep 10
done

oc rollout status -n nvidia-device-plugin daemonset/nvidia-device-plugin-daemonset

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
