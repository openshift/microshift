#!/bin/bash

set -eo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
MICROSHIFT_ROOT="${SCRIPTDIR}/../.."

declare -A UNAME_TO_GOARCH_MAP=( ["x86_64"]="amd64" ["aarch64"]="arm64" )


verify(){
    local -r okd_url=$1
    local -r okd_releaseTag=$2

    #stdout=$(oc adm release info "${okd_url}:${okd_releaseTag}" 2>&1)
    if ! stdout=$(oc adm release info "${okd_url}:${okd_releaseTag}" 2>&1)  ; then
        echo -e "error verifying okd release (URL: ${okd_url} , TAG: ${okd_releaseTag}) \nERROR: ${stdout}"
        exit 1
    fi
}

replace_assets(){
    local -r okd_url=$1
    local -r okd_releaseTag=$2
    local -r arch=$(uname -m)
    local -r temp_release_json=$(mktemp "/tmp/release-${arch}.XXXXX.json")

    oc adm release info --image-for="${op}" "${okd_url}:${okd_releaseTag}"

    # replace Microshift images with upstream (from OKD release)
    for op in $(jq -e -r  '.images | keys []' "${MICROSHIFT_ROOT}/assets/release/release-${arch}.json") 
    do
        local image
        image=$(oc adm release info --image-for="${op}" "${okd_url}:${okd_releaseTag}" || true) 
        if [ -n "${image}" ] ; then
            echo "${op} ${image}"
            jq --arg a "${op}" --arg b "${image}"  '.images[$a] = $b' "${MICROSHIFT_ROOT}/assets/release/release-${arch}.json" >"${temp_release_json}"
            mv "${temp_release_json}" "${MICROSHIFT_ROOT}/assets/release/release-${arch}.json"
        fi
    done

    pod_image=$(oc adm release info --image-for=pod "${okd_url}:${okd_releaseTag}" || true) 
    # update the infra pods for crio
    sed -i 's,pause_image .*,pause_image = '"\"${pod_image}\""',' "packaging/crio.conf.d/10-microshift_${UNAME_TO_GOARCH_MAP[${arch}]}.conf"

    # kube proxy is required for flannel
    kube_proxy_okd_image_with_hash=$(oc adm release info --image-for="kube-proxy" "${okd_url}:${okd_releaseTag}")
    echo "kube-proxy ${kube_proxy_okd_image_with_hash}"
    # The OKD image we retrieve is in the format quay.io/okd/scos-content@sha256:<hash>,
    # where the image name and digest (hash) are combined in a single string.
    # However, in the kustomization.${arch}.yaml file, we need the image name (newName) and
    # the digest in separate fields. To achieve this, we first extract the image name and digest
    # using parameter expansion, then use the yq command to insert these values into the
    # appropriate places within the YAML file.
    kube_proxy_okd_image_name="${kube_proxy_okd_image_with_hash%%@*}"
    kube_proxy_okd_image_hash="${kube_proxy_okd_image_with_hash##*@}"
    # install yq tool to update the image and hash
    "${MICROSHIFT_ROOT}"/scripts/fetch_tools.sh yq
    "${MICROSHIFT_ROOT}"/_output/bin/yq eval ".images[] |= select(.name == \"kube-proxy\") |= (.newName = \"${kube_proxy_okd_image_name}\" | .digest = \"${kube_proxy_okd_image_hash}\")" -i "${MICROSHIFT_ROOT}/assets/optional/kube-proxy/kustomization.${arch}.yaml"
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") --verify OKD_URL RELEASE_TAG         verify upstream release"
    echo "$(basename "$0") --replace OKD_URL RELEASE_TAG         replace microshift assets with upstream images"
    exit 1
}

if [ $# -eq 3 ] ; then
    case "$1" in
    --replace)
        verify "$2" "$3"
        replace_assets "$2" "$3"
        ;;
    --verify)
        verify "$2" "$3"
        ;;        
    *)
        usage
        ;;
    esac
else
    usage
fi


