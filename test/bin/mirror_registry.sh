#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

QUAY_VERSION=1.3.9
REGISTRY_HOST=${REGISTRY_HOST:-$(hostname):8443}
REGISTRY_ROOT=${REGISTRY_ROOT:-${HOME}/mirror-registry}
REGISTRY_CONTAINER_DIR=${REGISTRY_CONTAINER_DIR:-${REGISTRY_ROOT}/containers}
REGISTRY_CONTAINER_LIST=${REGISTRY_CONTAINER_LIST:-${REGISTRY_ROOT}/mirror-list.txt}
PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}

get_container_images() {
    containers=""
    local -r release_info_rpm=$(find "${IMAGEDIR}/rpm-repos" -name "microshift-release-info-*.rpm" | sort)
    if [ -z "${release_info_rpm}" ] ; then
        echo "Error: missing microshift-release-info RPMs"
        exit 1
    fi
    for package in ${release_info_rpm}; do
        containers="$(rpm2cpio "${package}" | cpio  -i --to-stdout "*release-$(uname -m).json" 2> /dev/null | jq -r '[ .images[] ] | join("\n")')\n${containers}"
    done
    echo -n -e "${containers}" | sort -u
}

prereqs() {
    mkdir -p "${REGISTRY_ROOT}"
    mkdir -p "${REGISTRY_CONTAINER_DIR}"
    curl -L https://github.com/quay/mirror-registry/releases/download/v${QUAY_VERSION}/mirror-registry-offline.tar.gz -o /tmp/mirror-registry-offline.tar.gz
    tar xf /tmp/mirror-registry-offline.tar.gz -C "${REGISTRY_ROOT}"
    rm -f /tmp/mirror-registry-offline.tar.gz
    sudo dnf install -y podman skopeo jq
}

setup_registry() {
    pushd "${REGISTRY_ROOT}" &>/dev/null
    ./mirror-registry install -r "${REGISTRY_ROOT}" --initUser microshift --initPassword microshift
    sudo cp "${REGISTRY_ROOT}/quay-rootCA/rootCA.pem" /etc/pki/ca-trust/source/anchors/mirror-registry.pem
    sudo update-ca-trust
    podman login -u microshift -p microshift "${REGISTRY_HOST}" --authfile "${REGISTRY_ROOT}/local-auth.json"
    popd &>/dev/null
}

mirror_images() {
    get_container_images > "${REGISTRY_CONTAINER_LIST}"
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --reg-to-dir "${PULL_SECRET}" "${REGISTRY_CONTAINER_LIST}" "${REGISTRY_CONTAINER_DIR}"
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --dir-to-reg "${REGISTRY_ROOT}/local-auth.json" "${REGISTRY_CONTAINER_DIR}" "${REGISTRY_HOST}"
    jq -s '.[0] * .[1]' "${PULL_SECRET}" "${REGISTRY_ROOT}/local-auth.json" > "${PULL_SECRET}.tmp"
    mv "${PULL_SECRET}.tmp" "${PULL_SECRET}"
    rm -rf "${REGISTRY_CONTAINER_DIR}"
}

prereqs
setup_registry
mirror_images
