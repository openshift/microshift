#!/bin/bash
#
# This script should be run on the hypervisor to download all of the
# images from composer.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

mkdir -p "${VM_DISK_DIR}"

download_image() {
    local buildid="${1}"
    local status

    status=$(sudo composer-cli compose status | grep "${buildid}" | awk '{print $2}')
    if [ "${status}" != "FINISHED" ]; then
        sudo composer-cli compose info "${buildid}"
        return 1
    fi

    # shellcheck disable=SC2086  # pass glob args without quotes
    rm -f ${buildid}-*.tar
    sudo composer-cli compose metadata "${buildid}"
    sudo composer-cli compose image "${buildid}"
    # shellcheck disable=SC2086  # pass glob args without quotes
    sudo chown "$(whoami)." ${buildid}-*

    return 0
}

download_dir="${IMAGEDIR}/builds"
mkdir -p "${download_dir}"

title "Downloading installer images to ${download_dir}"
cd "${download_dir}"
for blueprint_build in *.image-installer; do
    blueprint=$(basename "${blueprint_build}" .image-installer)
    buildid=$(cat "${blueprint_build}")
    if ! download_image "${buildid}"; then
        echo "Did not download image"
        continue
    fi
    iso_file="${buildid}-installer.iso"
    echo "Moving ${iso_file} to ${VM_DISK_DIR}/${blueprint}.iso"
    mv -f "${iso_file}" "${VM_DISK_DIR}/${blueprint}.iso"
done

if [ -d "${IMAGEDIR}/repo" ]; then
    echo "Cleaning up existing images in ${IMAGEDIR}/repo"
    rm -rf "${IMAGEDIR}/repo"
fi

title "Downloading ostree commits and metadata to ${download_dir}"
cd "${download_dir}"
for blueprint_build in *.edge-commit; do
    blueprint=$(basename "${blueprint_build}" .edge-commit)
    buildid=$(cat "${blueprint_build}")
    if ! download_image "${buildid}"; then
        echo "Did not download image"
        continue
    fi
    commit_file="${buildid}-commit.tar"
    echo "Unpacking ${commit_file} ${blueprint}"
    tar -C "${IMAGEDIR}" -xf "${commit_file}"
done

title "Updating ostree references in ${IMAGEDIR}/repo"
cd "${IMAGEDIR}"
ostree refs --repo=repo --force \
       --create "rhel-9.2-microshift-source-aux" "rhel-9.2-microshift-source"
ostree summary --update --repo=repo
ostree summary --view --repo=repo
