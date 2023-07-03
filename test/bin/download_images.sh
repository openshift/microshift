#!/bin/bash
#
# This script should be run on the hypervisor to download all of the
# images from composer.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

mkdir -p "${VM_DISK_DIR}"

download_image() {
    local buildid="${1}"
    # FIXME: These logs should go to the artifacts directory instead
    # of the build web server.
    # shellcheck disable=SC2086  # pass glob args without quotes
    rm -f ${buildid}-*.tar
    sudo composer-cli compose logs "${buildid}"
    sudo composer-cli compose metadata "${buildid}"
    sudo composer-cli compose image "${buildid}"
    # shellcheck disable=SC2086  # pass glob args without quotes
    sudo chown "$(whoami)." ${buildid}-*
}

download_dir="${IMAGEDIR}/builds"
mkdir -p "${download_dir}"

echo "Downloading installer images to ${download_dir}"
cd "${download_dir}"
for blueprint_build in *.image-installer; do
    blueprint=$(basename "${blueprint_build}" .image-installer)
    buildid=$(cat "${blueprint_build}")
    download_image "${buildid}"
    iso_file="${buildid}-installer.iso"
    echo "Moving ${iso_file} to ${VM_DISK_DIR}/${blueprint}.iso"
    mv -f "${iso_file}" "${VM_DISK_DIR}/${blueprint}.iso"
done

if [ -d "${IMAGEDIR}/repo" ]; then
    echo "Cleaning up existing images in ${IMAGEDIR}/repo"
    rm -rf "${IMAGEDIR}/repo"
fi

echo "Downloading ostree commits and metadata to ${download_dir}"
cd "${download_dir}"
for blueprint_build in *.edge-commit; do
    blueprint=$(basename "${blueprint_build}" .edge-commit)
    buildid=$(cat "${blueprint_build}")
    download_image "${buildid}"
    commit_file="${buildid}-commit.tar"
    echo "Unpacking ${commit_file} ${blueprint}"
    tar -C "${IMAGEDIR}" -xf "${commit_file}"
done

echo "Updating references"
cd "${IMAGEDIR}"
ostree summary --update --repo=repo
ostree summary --view --repo=repo
