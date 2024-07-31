#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOTDIR="${SCRIPTDIR}/../../.."

"${ROOTDIR}/scripts/devenv-builder/configure-vm.sh" --force-firewall --optional-rpms "${HOME}/.pull-secret.json"
sudo systemctl enable microshift
