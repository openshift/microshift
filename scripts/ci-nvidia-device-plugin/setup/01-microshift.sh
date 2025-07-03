#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOTDIR="${SCRIPTDIR}/../../.."

function configure_vm_has_arg() {
    local arg="$1"
    grep -q -- "${arg}" "${ROOTDIR}/scripts/devenv-builder/configure-vm.sh"
}

# Setup microshift before NVIDIA because configure-vm.sh sets fixed release
# which prevents `dnf module install` from updating the kernel to
# version from newer minor release of RHEL.

configure_vm_args=""
needs_cleanup=false

if configure_vm_has_arg '--no-start'; then
    configure_vm_args="--no-start"
else
    needs_cleanup=true
fi

if configure_vm_has_arg '--pull-images'; then
    configure_vm_args="${configure_vm_args} --pull-images"
fi

# shellcheck disable=SC2086
"${ROOTDIR}/scripts/devenv-builder/configure-vm.sh" \
    --force-firewall ${configure_vm_args} \
    "${HOME}/.pull-secret.json"

# NVIDIA Device Plugin requires reconfiguration of the CRIO.
# Let's keep MicroShift clean until the reboot
# when all pieces should be in place (driver, device plugin, etc.)
if "${needs_cleanup}"; then
    sudo microshift-cleanup-data --all --keep-images <<< 1
fi

sudo systemctl enable microshift

mkdir -p "${HOME}/artifacts"
