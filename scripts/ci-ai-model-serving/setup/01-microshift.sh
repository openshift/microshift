#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOTDIR="${SCRIPTDIR}/../../.."

# Setup microshift before NVIDIA because configure-vm.sh sets fixed release
# which prevents `dnf module install` from updating the kernel.

"${ROOTDIR}/scripts/devenv-builder/configure-vm.sh" \
    --no-start \
    --force-firewall \
    --optional-rpms \
    --skip-optional-rpms low-latency \
    "${HOME}/.pull-secret.json"

sudo systemctl enable microshift
