#!/bin/bash
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

DISTRIBUTION_VERSION=2.8.3
REGISTRY_IMAGE="quay.io/microshift/distribution:${DISTRIBUTION_VERSION}"
REGISTRY_HOST=${REGISTRY_HOST:-${MIRROR_REGISTRY_URL}}
PULL_SECRET=${PULL_SECRET:-${HOME}/.pull-secret.json}
LOCAL_REGISTRY_NAME="microshift-local-registry"

retry_pull_image() {
    for attempt in $(seq 3) ; do
        if ! podman pull "$@" ; then
            echo "WARNING: Failed to pull image, retry #${attempt}"
        else
            return 0
        fi
        sleep $(( "${attempt}" * 10 ))
    done

    echo "ERROR: Failed to pull image, quitting after 3 tries"
    return 1
}

prereqs() {
    # Install packages if not yet available locally
    if ! rpm -q podman skopeo jq &>/dev/null ; then
        "${SCRIPTDIR}/../../scripts/dnf_retry.sh" "install" "podman skopeo jq"
    fi
    podman stop "${LOCAL_REGISTRY_NAME}" || true
    podman rm "${LOCAL_REGISTRY_NAME}" || true
    retry_pull_image "${REGISTRY_IMAGE}"
    mkdir -p "${MIRROR_REGISTRY_DIR}"
}

setup_registry() {
    # Docker distribution does not support TLS authentication. The mirror-images.sh helper uses skopeo without tls options
    # and it defaults to https. Since this is not supported we need to configure registries.conf so that skopeo tries http instead.
    sudo bash -c 'cat > /etc/containers/registries.conf.d/900-microshift-mirror.conf' << EOF
[[registry]]
location = "${REGISTRY_HOST}"
insecure = true
EOF
    # Create the registry configuration file.
    # See https://distribution.github.io/distribution/about/configuration.
    cat > "${MIRROR_REGISTRY_DIR}/config.yaml" <<EOF
version: 0.1
log:
  accesslog:
    disabled: true
  level: info
storage:
    delete:
      enabled: false
    cache:
        blobdescriptor: inmemory
    filesystem:
        rootdirectory: /var/lib/registry
        maxthreads: 1024
    maintenance:
        uploadpurging:
            enabled: false
http:
    addr: :${MIRROR_REGISTRY_PORT}
health:
  storagedriver:
    enabled: false
EOF
    # Start the registry container
    podman run -d -p "${MIRROR_REGISTRY_PORT}:${MIRROR_REGISTRY_PORT}" --restart always \
        -v "${MIRROR_REGISTRY_DIR}:/var/lib/registry" \
        -v "${MIRROR_REGISTRY_DIR}/config.yaml:/etc/docker/registry/config.yml" \
        --name "${LOCAL_REGISTRY_NAME}" "${REGISTRY_IMAGE}"
}

mirror_images() {
    local -r ifile=$1
    local -r ofile=$(mktemp /tmp/container-list.XXXXXXXX)

    sort -u "${ifile}" > "${ofile}"
    "${ROOTDIR}/scripts/mirror-images.sh" --mirror "${PULL_SECRET}" "${ofile}" "${REGISTRY_HOST}"
    rm -f "${ofile}"
}

usage() {
    echo ""
    echo "Usage: ${0} [-cf FILE] [-bd DIR]"
    echo "   -cf FILE    File containing the container image references to mirror."
    echo "               Defaults to '${CONTAINER_LIST}', skipped if does not exist."
    echo ""
    echo "The registry data is stored at '${MIRROR_REGISTRY_DIR}' on the host."
    exit 1
}

#
# Main
#
image_list_file="${CONTAINER_LIST}"

while [ $# -gt 0 ]; do
    case $1 in
    -cf)
        shift
        [ -z "$1" ] && usage
        image_list_file=$1
        ;;
    *)
        usage
        ;;
    esac
    shift
done

prereqs
setup_registry

if [ -f "${image_list_file}" ]; then
    mirror_images "${image_list_file}"
else
    echo "WARNING: File '${image_list_file}' does not exist, skipping"
fi
