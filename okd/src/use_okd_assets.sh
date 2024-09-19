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


