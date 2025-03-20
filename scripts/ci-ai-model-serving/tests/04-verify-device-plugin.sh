#!/usr/bin/env bash

set -xeuo pipefail

oc rollout status -n nvidia-device-plugin daemonset/nvidia-device-plugin-daemonset

capacity=$(oc get node -o=jsonpath='{ .items[0].status.capacity }')
gpus=$(echo "${capacity}" | jq -r '."nvidia.com/gpu"')

if [[ "${gpus}" == "null" ]]; then
    echo "Node's capacity does not include NVIDIA GPU"
    exit 1
fi 
