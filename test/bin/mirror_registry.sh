#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

QUAY_VERSION=1.3.9
REGISTRY_HOST=${REGISTRY_HOST:-$(hostname):8443}
REGISTRY_ROOT=${REGISTRY_ROOT:-${HOME}/mirror-registry}
REGISTRY_CONTAINERS=${REGISTRY_CONTAINERS:-${HOME}/containers}
PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}

prereqs() {
    mkdir -p "${REGISTRY_ROOT}"
    mkdir -p "${REGISTRY_CONTAINERS}"
    curl -L https://github.com/quay/mirror-registry/releases/download/v${QUAY_VERSION}/mirror-registry-offline.tar.gz -o /tmp/mirror-registry-offline.tar.gz
    tar xf /tmp/mirror-registry-offline.tar.gz -C "${REGISTRY_ROOT}"
    rm -f /tmp/mirror-registry-offline.tar.gz
    sudo dnf install -y podman skopeo jq
}

setup_registry() {
    cd "${REGISTRY_ROOT}"
    ./mirror-registry install -r "${REGISTRY_ROOT}" --initUser microshift --initPassword microshift
    sudo cp "${REGISTRY_ROOT}/quay-rootCA/rootCA.pem" /etc/pki/ca-trust/source/anchors/mirror-registry.pem
    sudo update-ca-trust
    podman login -u microshift -p microshift "${REGISTRY_HOST}" --authfile "${REGISTRY_ROOT}/local-auth.json"
    cd -
}

mirror_images() {
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --reg-to-dir "${PULL_SECRET}" "${MIRROR_CONTAINERS_LIST}" "${REGISTRY_CONTAINERS}"
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --dir-to-reg "${REGISTRY_ROOT}/local-auth.json" "${REGISTRY_CONTAINERS}" "${REGISTRY_HOST}"
    jq -s '.[0] * .[1]' "${PULL_SECRET}" "${REGISTRY_ROOT}/local-auth.json" > "${PULL_SECRET}.tmp"
    mv "${PULL_SECRET}.tmp" "${PULL_SECRET}"
    rm -r "${REGISTRY_CONTAINERS}"
}

prereqs
setup_registry
mirror_images
