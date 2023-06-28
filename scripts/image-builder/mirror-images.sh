#!/bin/bash
set -euo pipefail

function usage() {
    echo "Usage: $(basename "$0") <--mirror | --reg-to-dir | --dir-to-reg> <OPTIONS>"
    echo ""
    echo "  --mirror <pull_secret_file> <image_list_file> <target_registry_host_port>"
    echo "          Mirror images from the specified image list file to the target registry."
    echo "          The pull secret file should contain credentials both for the source and"
    echo "          target registries."
    echo "  --reg-to-dir <pull_secret_file> <image_list_file> <local_directory>"
    echo "          Download images from the specified image list file to the local directory."
    echo "          The pull secret file should contain credentials for the source registry."
    echo "  --dir-to-reg <pull_secret_file> <source_directory> <target_registry_host_port>"
    echo "          Upload images from the local directory to the target registry."
    echo "          The pull secret file should contain credentials for the target registry."

    exit 1
}

function mirror_registry() {
    local img_pull_file=$1
    local img_file_list=$2
    local dest_registry=$3

    # Use timestamp and counter as a tag on the target images to avoid
    # their overwrite by the 'latest' automatic tagging
    local -r image_tag=mirror-$(date +%y%m%d%H%M%S)
    local image_cnt=1

    while read -r src_img ; do
        # Remove the source registry prefix and SHA
        local dst_img
        dst_img=$(echo "${src_img}" | cut -d '/' -f 2-)
        dst_img=$(echo "${dst_img}" | awk -F'@' '{print $1}')
        # Add the target registry prefix
        dst_img="${dest_registry}/${dst_img}"

        # Run the image copy command
        echo "Mirroring '${src_img}' to '${dst_img}'"
        skopeo copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            docker://"${src_img}" docker://"${dst_img}:${image_tag}-${image_cnt}"
        # Increment the counter
        (( image_cnt += 1 ))

    done < "${img_file_list}"
}

function registry_to_dir() {
    local img_pull_file=$1
    local img_file_list=$2
    local local_dir=$3

    while read -r src_img ; do
        # Remove the source registry prefix
        local dst_img
        dst_img=$(echo "${src_img}" | cut -d '/' -f 2-)

        # Run the image download command
        echo "Downloading '${src_img}' to '${local_dir}'"
        mkdir -p "${local_dir}/${dst_img}"
        skopeo copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            docker://"${src_img}" dir://"${local_dir}/${dst_img}"

    done < "${img_file_list}"
}

function dir_to_registry() {
    local img_pull_file=$1
    local local_dir=$2
    local dest_registry=$3

    # Use timestamp and counter as a tag on the target images to avoid
    # their overwrite by the 'latest' automatic tagging
    local -r image_tag=mirror-$(date +%y%m%d%H%M%S)
    local image_cnt=1

    pushd "${local_dir}" >/dev/null
    while read -r src_manifest ; do
        # Remove the manifest.json file name
        local src_img
        src_img=$(dirname "${src_manifest}")
        # Add the target registry prefix and remove SHA
        local dst_img
        dst_img="${dest_registry}/${src_img}"
        dst_img=$(echo "${dst_img}" | awk -F'@' '{print $1}')

        # Run the image upload command
        echo "Uploading '${src_img}' to '${dst_img}'"
        skopeo copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            dir://"${local_dir}/${src_img}" docker://"${dst_img}:${image_tag}-${image_cnt}"
        # Increment the counter
        (( image_cnt += 1 ))

    done < <(find . -type f -name manifest.json -printf '%P\n')
    popd >/dev/null
}

#
# Main
#
if [ $# -ne 4 ] ; then
    usage
fi

case "$1" in
--mirror)
    mirror_registry "$2" "$3" "$4"
    ;;
--reg-to-dir)
    registry_to_dir "$2" "$3" "$4"
    ;;
--dir-to-reg)
    dir_to_registry "$2" "$3" "$4"
    ;;
*)
    usage
    ;;
esac
