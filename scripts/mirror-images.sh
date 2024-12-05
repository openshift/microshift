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

function skopeo_retry() {
    for attempt in $(seq 3) ; do
        if ! skopeo "$@" ; then
            echo "WARNING: Failed to run skopeo, retry #${attempt}"
        else
            return 0
        fi
        sleep $(( "${attempt}" * 10 ))
    done

    echo "ERROR: Failed to run skopeo, quitting after 3 tries"
    return 1
}

function mirror_registry() {
    local -r img_pull_file=$1
    local -r img_file_list=$2
    local -r dest_registry=$3

    process_image_copy() {
        local -r img_pull_file=$1
        local -r dest_registry=$2
        local -r src_img=$(cut -d' ' -f2 <<< "$3")

        # Remove the source registry prefix and SHA
        local dst_img
        dst_img=$(echo "${src_img}" | cut -d '/' -f 2-)
        local dst_img_no_tag
        dst_img_no_tag="${dst_img%%[@:]*}"
        # Add the target registry prefix
        dst_img="${dest_registry}/${dst_img}"
        dst_img_no_tag="${dest_registry}/${dst_img_no_tag}"

        # Run the image mirror and tag command
        echo "Mirroring '${src_img}' to '${dst_img}'"
        skopeo_retry copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            docker://"${src_img}" docker://"${dst_img}"

        echo "Tagging '${dst_img_no_tag}' as 'latest'"
        skopeo_retry copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            docker://"${dst_img}" docker://"${dst_img_no_tag}:latest"

    }

    # Export functions for xargs to use
    export -f process_image_copy
    export -f skopeo_retry
    # Note that images are passed as one argument by replacing "{}" in xarg input.
    xargs -P 8 -I {} -a "${img_file_list}" \
        bash -c 'process_image_copy "$@"' _ "${img_pull_file}" "${dest_registry}" "{}"
}

function registry_to_dir() {
    local img_pull_file=$1
    local img_file_list=$2
    local local_dir=$3

    process_image_copy() {
        local -r img_pull_file=$1
        local -r local_dir=$2
        local -r src_img=$3

        # Remove the source registry prefix
        local dst_img
        dst_img=$(echo "${src_img}" | cut -d '/' -f 2-)

        # Run the image download command
        echo "Downloading '${src_img}' to '${local_dir}'"
        mkdir -p "${local_dir}/${dst_img}"
        skopeo_retry copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            docker://"${src_img}" dir://"${local_dir}/${dst_img}"
    }

    # Export functions for xargs to use
    export -f process_image_copy
    export -f skopeo_retry
    # Generate a list for each image and run copy in parallel.
    # Note that the image is passed by replacing "{}" in xarg input.
    xargs -P 8 -I {} -a "${img_file_list}" \
        bash -c 'process_image_copy "$@"' _ "${img_pull_file}" "${local_dir}" "{}"
}

function dir_to_registry() {
    local img_pull_file=$1
    local local_dir=$2
    local dest_registry=$3

    process_image_copy() {
        local -r img_pull_file=$1
        local -r local_dir=$2
        local -r dest_registry=$3
        local -r src_manifest=$4

        # Remove the manifest.json file name
        local src_img
        src_img=$(dirname "${src_manifest}")
        # Add the target registry prefix and remove SHA
        local -r dst_img="${dest_registry}/${src_img}"
        local -r dst_img_no_tag="${dest_registry}/${src_img%%[@:]*}"

        # Run the image upload and tag commands
        echo "Uploading '${src_img}' to '${dst_img}'"
        skopeo_retry copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            dir://"${local_dir}/${src_img}" docker://"${dst_img}"

        echo "Tagging '${dst_img}' as '${dst_img_no_tag}:latest'"
        skopeo_retry copy --all --quiet \
            --preserve-digests \
            --authfile "${img_pull_file}" \
            docker://"${dst_img}" docker://"${dst_img_no_tag}:latest"
    }

    # Export functions for xargs to use
    export -f process_image_copy
    export -f skopeo_retry

    # Generate a list with an incremental counter for each image and run copy in parallel.
    # Note that the counter and image pairs are passed as one argument by replacing "{}" in xarg input.
    pushd "${local_dir}" >/dev/null
    find . -type f -name manifest.json -printf '%P\n' | \
        xargs -P 8 -I {} \
        bash -c 'process_image_copy "$@"' _ "${img_pull_file}" "${local_dir}" "${dest_registry}" "{}"
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
