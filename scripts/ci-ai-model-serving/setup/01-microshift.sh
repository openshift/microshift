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
    --pull-images \
    --skip-optional-rpms low-latency \
    "${HOME}/.pull-secret.json"

# Pull kserve and helper images for ai-model-serving, skip serving runtimes
# shellcheck disable=SC2046
"${ROOTDIR}/scripts/pull_retry.sh" $(rpm -qa | grep -e  "microshift-ai-model-serving.*-release-info" | xargs rpm -ql | grep $(uname -m).json | xargs jq -r '.images | values[]' | grep -E "kserve|auth-proxy")

sudo systemctl enable microshift

mkdir -p "${HOME}/artifacts"
microshift version -o json | jq -r '.gitVersion' | cut -d'.' -f1-2 > "${HOME}/artifacts/ocp.version"
git -C "${ROOTDIR}" rev-parse HEAD > "${HOME}/artifacts/ci_artifact.git_version"
