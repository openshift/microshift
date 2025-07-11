#!/usr/bin/env bash

set -xeuo pipefail

# ci-nvidia-device-plugin scripts are used by many branches, not only the main branch.
# Therefore the scripts include workarounds for older MicroShift versions.
# For the versions that do not have the healthcheck command, the script uses oc rollout status.

if microshift --help | grep -q "healthcheck"; then
    sudo microshift healthcheck \
        --namespace nvidia-device-plugin \
        --daemonsets nvidia-device-plugin-daemonset
else
    oc rollout status \
        --timeout=180s \
        -n nvidia-device-plugin daemonset/nvidia-device-plugin-daemonset
fi

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
