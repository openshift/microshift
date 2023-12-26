#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

DISTRIBUTION_VERSION=2.8.3
REGISTRY_HOST=${REGISTRY_HOST:-$(hostname):5000}
REGISTRY_ROOT=${REGISTRY_ROOT:-${HOME}/mirror-registry}
REGISTRY_CONTAINER_DIR=${REGISTRY_CONTAINER_DIR:-${REGISTRY_ROOT}/containers}
REGISTRY_CONTAINER_LIST=${REGISTRY_CONTAINER_LIST:-${REGISTRY_ROOT}/mirror-list.txt}
PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}
LOCAL_REGISTRY_NAME="microshift-local-registry"

get_container_images() {
    containers=""
    local -r release_info_rpm=$(find "${IMAGEDIR}/rpm-repos" -name "microshift-release-info-*.rpm" | sort)
    if [ -z "${release_info_rpm}" ] ; then
        echo "Error: missing microshift-release-info RPMs"
        exit 1
    fi
    for package in ${release_info_rpm}; do
        echo "Getting image references from RPM ${package}..."
        containers="$(rpm2cpio "${package}" | cpio  -i --to-stdout "*release-$(uname -m).json" 2> /dev/null | jq -r '[ .images[] ] | join("\n")')\n${containers}"
    done
    echo -n -e "${containers}" | sort -u > "${REGISTRY_CONTAINER_LIST}"
}

prereqs() {
    mkdir -p "${REGISTRY_ROOT}"
    mkdir -p "${REGISTRY_CONTAINER_DIR}"
    "${SCRIPTDIR}/../../scripts/dnf_retry.sh" "install" "podman skopeo jq"
    podman stop "${LOCAL_REGISTRY_NAME}" || true
    podman rm "${LOCAL_REGISTRY_NAME}" || true
    podman run -d -p 5000:5000 --restart always --name "${LOCAL_REGISTRY_NAME}" "quay.io/microshift/distribution:${DISTRIBUTION_VERSION}"
}

setup_registry() {
    # Docker distribution does not support TLS authentication. The mirror-images.sh helper uses skopeo without tls options
    # and it defaults to https. Since this is not supported we need to configure registries.conf so that skopeo tries http instead.
    sudo bash -c 'cat > /etc/containers/registries.conf.d/900-microshift-mirror.conf' << EOF
[[registry]]
location = "$(hostname)"
insecure = true
EOF
    sudo systemctl restart podman
}

mirror_images() {
    get_container_images
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --reg-to-dir "${PULL_SECRET}" "${REGISTRY_CONTAINER_LIST}" "${REGISTRY_CONTAINER_DIR}"
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --dir-to-reg "${REGISTRY_ROOT}/local-auth.json" "${REGISTRY_CONTAINER_DIR}" "${REGISTRY_HOST}"
    rm -rf "${REGISTRY_CONTAINER_DIR}"
}

prereqs
setup_registry
mirror_images
