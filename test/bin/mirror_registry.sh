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
        sleep 10
    done

    echo "ERROR: Failed to pull image, quitting after 3 tries"
    return 1
}

prereqs() {
    "${SCRIPTDIR}/../../scripts/dnf_retry.sh" "install" "podman skopeo jq"
    podman stop "${LOCAL_REGISTRY_NAME}" || true
    podman rm "${LOCAL_REGISTRY_NAME}" || true
    retry_pull_image "${REGISTRY_IMAGE}"
    mkdir -p "${MIRROR_REGISTRY_DIR}"
    podman run -d -p 5000:5000 --restart always \
        -v "${MIRROR_REGISTRY_DIR}:/var/lib/registry" \
        --name "${LOCAL_REGISTRY_NAME}" "${REGISTRY_IMAGE}"
}

setup_registry() {
    # Docker distribution does not support TLS authentication. The mirror-images.sh helper uses skopeo without tls options
    # and it defaults to https. Since this is not supported we need to configure registries.conf so that skopeo tries http instead.
    sudo bash -c 'cat > /etc/containers/registries.conf.d/900-microshift-mirror.conf' << EOF
[[registry]]
location = "${REGISTRY_HOST}"
insecure = true
EOF
    sudo systemctl restart podman
}

mirror_images() {
    local -r ifile=$1
    local -r ofile=$(mktemp /tmp/container-list.XXXXXXXX)

    sort -u "${ifile}" > "${ofile}"
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --mirror "${PULL_SECRET}" "${ofile}" "${REGISTRY_HOST}"
    rm -f "${ofile}"
}

mirror_bootc_images() {
    local -r idir=$1
    "${ROOTDIR}/scripts/image-builder/mirror-images.sh" --dir-to-reg "${PULL_SECRET}" "${idir}" "${REGISTRY_HOST}"
}

usage() {
    echo ""
    echo "Usage: ${0} [-cf FILE] [-bd DIR]"
    echo "   -cf FILE    File containing the container image references to mirror."
    echo "               Defaults to '${CONTAINER_LIST}', skipped if does not exist."
    echo "   -bd DIR     Directory containing the bootc containers data to mirror."
    echo "               Defaults to '${BOOTC_IMAGE_DIR}', skipped if does not exist."
    echo ""
    echo "The registry data is stored at '${MIRROR_REGISTRY_DIR}' on the host."
    exit 1
}

#
# Main
#
image_list_file="${CONTAINER_LIST}"
bootc_image_dir="${BOOTC_IMAGE_DIR}"

while [ $# -gt 0 ]; do
    case $1 in
    -cf)
        shift
        [ -z "$1" ] && usage
        image_list_file=$1
        ;;
    -bd)
        shift
        [ -z "$1" ] && usage
        bootc_image_dir=$1
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

if [ -d "${bootc_image_dir}" ] ; then
    mirror_bootc_images "${bootc_image_dir}"
else
    echo "WARNING: Directory '${bootc_image_dir}' does not exist, skipping"
fi
