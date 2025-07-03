#!/usr/bin/env bash

set -xeuo pipefail

# https://docs.nvidia.com/datacenter/cloud-native/edge/latest/nvidia-gpu-with-device-edge.html#installing-the-nvidia-device-plugin
# https://github.com/NVIDIA/k8s-device-plugin?tab=readme-ov-file#enabling-gpu-support-in-kubernetes

DEST="/etc/microshift/manifests.d/10-nvidia-device-plugin"
VER="v0.17.0"
URL="https://raw.githubusercontent.com/NVIDIA/k8s-device-plugin/refs/tags/${VER}/deployments/static/nvidia-device-plugin-privileged-with-service-account.yml"

sudo mkdir -p "${DEST}"
sudo curl -s -L "${URL}" -o "${DEST}/nvidia-device-plugin.yml"

cat << EOF | sudo tee "${DEST}/kustomization.yaml" > /dev/null
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - nvidia-device-plugin.yml
EOF

mkdir -p "${HOME}/artifacts"
echo "${VER}" > "${HOME}/artifacts/operator.version"
