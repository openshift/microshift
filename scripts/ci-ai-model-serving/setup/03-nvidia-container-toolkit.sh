#!/usr/bin/env bash

set -xeuo pipefail

### CONTAINER TOOLKIT
# https://docs.nvidia.com/datacenter/cloud-native/container-toolkit/latest/install-guide.html#with-dnf-rhel-centos-fedora-amazon-linux
# https://github.com/NVIDIA/k8s-device-plugin/blob/67223246a979c9d35eea372ea6a3fdd0a8c28e90/README.md?plain=1#L104-L142

curl -s -L https://nvidia.github.io/libnvidia-container/stable/rpm/nvidia-container-toolkit.repo | \
  sudo tee /etc/yum.repos.d/nvidia-container-toolkit.repo

sudo dnf config-manager --enable nvidia-container-toolkit-experimental

sudo dnf install nvidia-container-toolkit -y
sudo setsebool -P container_use_devices on

sudo nvidia-ctk runtime configure --runtime=crio --set-as-default --config=/etc/crio/crio.conf.d/99-nvidia.conf

# Update runtimes from ["docker-runc", "runc", "crun"] to ["crun", "docker-runc", "runc"]
sudo sed -i 's/^runtimes =.*$/runtimes = ["crun", "docker-runc", "runc"]/g' /etc/nvidia-container-runtime/config.toml

sudo systemctl restart crio
