#!/usr/bin/env bash

set -xeuo pipefail

mkdir -p ~/.kube
# shellcheck disable=SC2024
sudo cat /var/lib/microshift/resources/kubeadmin/kubeconfig > ~/.kube/config
