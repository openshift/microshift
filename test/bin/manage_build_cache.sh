#!/bin/bash
#
# This script should be run on the build host to manage cache of build artifacts
set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

AWS_BUCKET_NAME="${AWS_BUCKET_NAME:-microshift-build-cache}"
BCH_SUBDIR=
TAG_SUBDIR=

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (upload|download|verify|setlast|getlast|keep) [options]

Manage build cache at the '${AWS_BUCKET_NAME}' AWS S3 bucket. The script
assumes that the bucket exists and it is configured for read-write access
using 'aws s3 (ls|sync|rm)' operations.

The cache directory structure is '${AWS_BUCKET_NAME}/<branch>/<arch>/<tag>'.

  -h        Show this help.

  upload:   Upload build artifacts from the local disk to the specified
            '${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/<tag>' AWS S3 bucket.

  download: Download build artifacts from the specified
            '${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/<tag>' AWS S3 bucket
            to the local disk.

  verify:   Exit with 0 status if the specified
            '${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/<tag>' sub-directory
            exists and contains files, 1 otherwise.

  setlast:  Update the '${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/last' file
            contents in the AWS S3 bucket with the specified '<tag>'.

  getlast:  Retrieve the '${AWS_BUCKET_NAME}/<branch>/${UNAME_M}/last'
            file contents from the AWS S3 bucket. The output format is
            "LAST: <tag>" for easy parsing. The script returns the
            specified '<tag>' as a fallback if the bucket file does
            not exist.

  keep:     Delete all data from the '${AWS_BUCKET_NAME}/<branch>/${UNAME_M}'
            AWS S3 bucket, only keeping the 'last' and the specified
            '<tag>' sub-directories.

Options:

  -b <branch>   The branch sub-directory in the AWS S3 bucket.

  -t <tag>      The tag sub-directory in the AWS S3 bucket.

EOF
}

action_upload() {
    local -r src_base="${IMAGEDIR}"
    local -r dst_base="s3://${AWS_BUCKET_NAME}/${BCH_SUBDIR}/${UNAME_M}/${TAG_SUBDIR}"

    # Upload ISO images
    local -r iso_base="${src_base}/${VM_POOL_BASENAME}"
    local -r iso_size="$(du -csh "${iso_base}" | awk 'END{print $1}')"
    local -r iso_dest="${dst_base}/${VM_POOL_BASENAME}"

    echo "Uploading ${iso_size} of ISO images to '${iso_dest}'"
    "${AWSCLI}" s3 sync --quiet --include '*.iso' "${iso_base}" "${iso_dest}"

    # Upload ostree commits
    local -r repo_base="${src_base}/repo"
    local -r repo_size="$(du -csh "${repo_base}" | awk 'END{print $1}')"
    local -r repo_dest="${dst_base}/repo"

    # Create dummy files in empty directories
    find "${repo_base}" -type d -empty | while IFS= read -r dir; do
        touch "${dir}/.s3-sync-empty-dir"
    done

    echo "Uploading ${repo_size} of ostree commits to '${repo_dest}'"
    "${AWSCLI}" s3 sync --quiet "${repo_base}" "${repo_dest}"
}

action_download() {
    local -r src_base="s3://${AWS_BUCKET_NAME}/${BCH_SUBDIR}/${UNAME_M}/${TAG_SUBDIR}"
    local -r dst_base="${IMAGEDIR}"

    # Download ISO images
    local -r iso_base="${src_base}/${VM_POOL_BASENAME}"
    local -r iso_dest="${dst_base}/${VM_POOL_BASENAME}"

    echo "Downloading ISO images from '${iso_base}'"
    "${AWSCLI}" s3 sync --quiet --include '*.iso' "${iso_base}" "${iso_dest}"

    local -r iso_size="$(du -csh "${iso_dest}" | awk 'END{print $1}')"
    echo "Downloaded ${iso_size} of ISO images"

    # Download ostree commits
    local -r repo_base="${src_base}/repo"
    local -r repo_dest="${dst_base}/repo"

    echo "Downloading ostree commits from '${repo_base}'"
    "${AWSCLI}" s3 sync --quiet "${repo_base}" "${repo_dest}"

    local -r repo_size="$(du -csh "${repo_dest}" | awk 'END{print $1}')"
    echo "Downloaded ${repo_size} of ostree commits"
}

action_verify() {
    local -r src_dir="s3://${AWS_BUCKET_NAME}/${BCH_SUBDIR}/${UNAME_M}/${TAG_SUBDIR}/${VM_POOL_BASENAME}"

    echo "Checking contents of '${src_dir}'"
    if "${AWSCLI}" s3 ls "${src_dir}/" | awk '{print $NF}' | grep -Eq '.iso$' ; then
        echo OK
        exit 0
    fi

    echo KO
    exit 1
}

action_setlast() {
    local -r src_file="$(mktemp /tmp/setlast.XXXXXXXX)"
    local -r dst_file="s3://${AWS_BUCKET_NAME}/${BCH_SUBDIR}/${UNAME_M}/last"

    if [ "${TAG_SUBDIR}" = "last" ] ; then
        echo "ERROR: Cannot set 'last' tag to itself"
        exit 1
    fi

    echo "Updating '${dst_file}' with the '${TAG_SUBDIR}' tag"
    echo -n "${TAG_SUBDIR}" > "${src_file}"
    "${AWSCLI}" s3 cp --quiet "${src_file}" "${dst_file}"
    rm -f "${src_file}"
}

action_getlast() {
    local -r src_file="s3://${AWS_BUCKET_NAME}/${BCH_SUBDIR}/${UNAME_M}/last"
    local -r dst_file="$(mktemp /tmp/getlast.XXXXXXXX)"

    echo "Reading '${src_file}' tag contents"
    "${AWSCLI}" s3 cp --quiet "${src_file}" "${dst_file}" || true
    if [ -s "${dst_file}" ] ; then
        echo "LAST: $(cat "${dst_file}")"
    else
        echo "LAST: ${TAG_SUBDIR}"
    fi
    rm -f "${dst_file}"
}

action_keep() {
    local -r top_dir="s3://${AWS_BUCKET_NAME}/${BCH_SUBDIR}/${UNAME_M}"
    # Get the last contents with the ${TAG_SUBDIR} default
    local -r last_dir="$(action_getlast | awk '/LAST:/ {print $NF}')"

    for sub_dir in $("${AWSCLI}" s3 ls "${top_dir}/" | awk '{print $NF}'); do
        if [ "${sub_dir}" = "last" ] ; then
            continue
        fi
        if [ "${sub_dir}" = "${TAG_SUBDIR}/" ] || [ "${sub_dir}" = "${last_dir}/" ] ; then
            echo "Keeping '${sub_dir}' sub-directory"
            continue
        fi

        echo "Deleting '${sub_dir}' sub-directory"
        "${AWSCLI}" s3 rm --recursive "${top_dir}/${sub_dir}"
    done
}

#
# Main function
#
if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

# Parse options
while getopts "b:t:h" opt; do
    case "${opt}" in
        h)
            usage
            exit 0
            ;;
        b)
            BCH_SUBDIR="${OPTARG}"
            ;;
        t)
            TAG_SUBDIR="${OPTARG}"
            ;;
        *)
            usage
            exit 1
            ;;
    esac
done

if [ -z "${BCH_SUBDIR}" ] || [ -z "${TAG_SUBDIR}" ] ; then
    echo "ERROR: The <branch> and <tag> sub-directory values are mandatory"
    echo
    usage
    exit 1
fi

# Install AWS CLI tools
if [ ! -e "${AWSCLI}" ] ; then
    "${ROOTDIR}/scripts/fetch_tools.sh" awscli
fi

# Verify the bucket can be accessed
if ! "${AWSCLI}" s3 ls "${AWS_BUCKET_NAME}" &>/dev/null ; then
    echo "ERROR: Cannot access the '${AWS_BUCKET_NAME}' AWS bucket"
    exit 1
fi

# Run actions
case "${action}" in
    upload|download|verify|setlast|getlast|keep)
        "action_${action}" "$@"
        ;;
    -h|help)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
