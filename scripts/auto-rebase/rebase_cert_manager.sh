#!/usr/bin/env bash
# shellcheck disable=all
#   Copyright 2022 The MicroShift authors
#
#   Licensed under the Apache License, Version 2.0 (the "License");
#   you may not use this file except in compliance with the License.
#   You may obtain a copy of the License at
#
#       http://www.apache.org/licenses/LICENSE-2.0
#
#   Unless required by applicable law or agreed to in writing, software
#   distributed under the License is distributed on an "AS IS" BASIS,
#   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#   See the License for the specific language governing permissions and
#   limitations under the License.
#

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

shopt -s expand_aliases
shopt -s extglob

#debugging options
#trap 'echo "#L$LINENO: $BASH_COMMAND" >&2' DEBUG
#set -xo functrace
#PS4='+ $LINENO  '
REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../..")"
STAGING_DIR="$REPOROOT/_output/staging"
export REGISTRY_AUTH_FILE="${HOME}/.pull-secret.json"
OPERATOR_INDEX="${STAGING_DIR}/redhat-operator-index.yaml"
OPERATOR_CERT_MANAGER_INDEX="${STAGING_DIR}/redhat-operator-cert-manager-index.yaml"
GO_MOD_DIRS=("$REPOROOT/" "$REPOROOT/etcd")

OPERATOR_COMPONENTS="cert-manager-controller cert-manager-ca-injector cert-manager-webhook cert-manager-acmesolver cert-manager-istiocsr"

declare -a ARCHS=("amd64" "arm64")
declare -A GOARCH_TO_UNAME_MAP=( ["amd64"]="x86_64" ["arm64"]="aarch64" )

title() {
    echo -e "\E[34m$1\E[00m";
}

check_preconditions() {
    if ! hash yq; then
        title "Installing yq"
        sudo DEST_DIR=/usr/bin/ "${REPOROOT}/scripts/fetch_tools.sh" yq
    fi

    if ! hash opm; then
        title "Installing opm"
        sudo DEST_DIR=/usr/bin/ "${REPOROOT}/scripts/fetch_tools.sh" opm
    fi


    if ! hash python3; then
        echo "ERROR: python3 is not present on the system - please install"
        exit 1
    fi

    if ! python3 -c "import yaml"; then
        echo "ERROR: missing python's yaml library - please install"
        exit 1
    fi
}

# Clone a repo at a commit
clone_repo() {
    local repo="$1"
    local commit="$2"
    local destdir="$3"

    local repodir="${destdir}/${repo##*/}"

    if [[ -d "${repodir}" ]]
    then
        return
    fi

    git init "${repodir}"
    pushd "${repodir}" >/dev/null
    git remote add origin "${repo}"
    git fetch origin --quiet  --filter=tree:0 --tags "${commit}"
    git checkout "${commit}"
    popd >/dev/null
}

download_cert_manager(){
    rm -rf "${STAGING_DIR}"
    mkdir -p "${STAGING_DIR}"
    pushd "${STAGING_DIR}" >/dev/null

   #  export REGISTRY_AUTH_FILE=${PULL_SECRET_FILE}

   operator_manifest="$1"

    # get the whole operator yaml for 4.19
    opm render "${operator_manifest}" -o yaml  >${OPERATOR_INDEX}

    # find the latest published cert-manager-operator ie: cert-manager-operator.v1.16.0
    export operator=$(yq 'select(.package == "openshift-cert-manager-operator" and .name == "stable-v1") | .entries[-1].name' ${OPERATOR_INDEX})
    yq 'select (.name==env(operator))' ${OPERATOR_INDEX} >"${OPERATOR_CERT_MANAGER_INDEX}"
 
    echo  "found operator version ${operator}"

    # convert from cert-manager-operator.v1.16.0 to cert-manager-x.y
    branch_name=$(echo ${operator} | awk -F'[^0-9]*' '{print "cert-manager-"$2"."$3}')
    clone_repo "https://github.com/openshift/cert-manager-operator" "$branch_name" "."

}

# Updates the image digests in pkg/release/release*.go
# update_images() {
#     if [ ! -f "${STAGING_DIR}/release_amd64.json" ] || [ ! -f "${STAGING_DIR}/release_arm64.json" ]; then
#         >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
#         exit 1
#     fi
#     pushd "${STAGING_DIR}" >/dev/null

   
# }



write_cert_manager_images_for_arch() {
    local arch="$1"
    title "Updating images for ${arch}"
    #local csv_manifest="${arch_dir}/servicemeshoperator3.clusterserviceversion.yaml"
    #local kustomization_arch_file="${REPOROOT}/assets/optional/gateway-api/kustomization.${GOARCH_TO_UNAME_MAP[${arch}]}.yaml"
    local cert_manager_release_json="${REPOROOT}/assets/optional/cert-manager/release-cert-manager-${GOARCH_TO_UNAME_MAP[${arch}]}.json"
    local cert_manager_operator_yaml="${REPOROOT}/assets/optional/cert-manager/manager/manager.yaml"
    local cert_manager_kustomization_yaml="${REPOROOT}/assets/optional/cert-manager/manager/kustomization.yaml"

    local base_release=4.20
    jq -n "{\"release\": {\"base\": \"${base_release}\"}, \"images\": {}}" > "${cert_manager_release_json}"
    
    #containerImage
    local operatorImage=$(yq '.properties[] | select(.type == "olm.csv.metadata").value.annotations.containerImage' "${OPERATOR_CERT_MANAGER_INDEX}")
    
    yq -i -o json ".images += {\"cert-manager-operator\": \"${operatorImage}\"}" "${cert_manager_release_json}"
    sed -i "s#newName:.*openshift.io\/cert-manager-operator.*#newName: ${operatorImage}#g" "${cert_manager_kustomization_yaml}"

    #relatedImages
    for index in $(yq '.relatedImages.[] | path | .[-1] ' "${OPERATOR_CERT_MANAGER_INDEX}"); do
     local image=$(yq ".relatedImages.${index}.image" "${OPERATOR_CERT_MANAGER_INDEX}" )
     local component=$(yq ".relatedImages.${index}.name" "${OPERATOR_CERT_MANAGER_INDEX}")
    if [[  -n "${component}" && "${OPERATOR_COMPONENTS}" == *"${component}"* ]]; then
        yq -i -o json ".images += {\"${component}\": \"${image}\"}" "${cert_manager_release_json}"
        sed -i "s#value:.*${component}.*#value: ${image}#g" "${cert_manager_operator_yaml}"

        # handle special case istiocsr v istio-csr mismatch 
        if [[ "${component}" == "cert-manager-istiocsr" ]]; then
            sed -i "s#value:.*cert-manager-istio-csr.*#value: ${image}#g" "${cert_manager_operator_yaml}"
        fi
    fi
       

    done

}

update_cert_manager_images() {
    title "Updating cert_manager images"
    local workdir="${STAGING_DIR}/cert-manager-operator"
    [ -d "${workdir}" ] || {
        >&2 echo 'cert_manager staging dir not found, aborting image update'
        return 1
    }
    for arch in "${ARCHS[@]}"; do
        write_cert_manager_images_for_arch "${arch}"
    done
}


copy_manifests() {
   # if [ ! -f "${STAGING_DIR}/release_amd64.json" ]; then
   #     >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
   #     exit 1
   # fi
    title "Copying manifests"
    "$REPOROOT/scripts/auto-rebase/handle_assets.py" "./scripts/auto-rebase/assets_cert_manager.yaml"
}


# Updates embedded component manifests by gathering these from various places
# in the staged repos and copying them into the asset directorcay.
update_cert_manager_manifests() {
    pushd "${STAGING_DIR}" >/dev/null

    title "Modifying OpenShift manifests"
    
    for index in $(yq '.[] | path | .[-1] ' "${OPERATOR_CERT_MANAGER_INDEX}")
    do
     image=$(yq ".${index}.image" "${OPERATOR_CERT_MANAGER_INDEX}")
     component=$(yq ".${index}.name" "${OPERATOR_CERT_MANAGER_INDEX}")

    if [[  -n "${component}" && "${OPERATOR_COMPONENTS}" == *"${component}"* ]]; then
        #clone_repo "${repo}" "${commit}" "."
        #echo "${repo} embedded-component ${commit}" >> "${new_commits_file}"
        echo "${image} ${component}"
    fi
    done


    popd >/dev/null
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM         Performs all the steps to rebase to a release image. Specify both amd64 and arm64 OCP releases."
    echo "$(basename "$0") download RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM   Downloads the content of a release image to disk in preparation for rebasing. Specify both amd64 and arm64 OCP releases."
    echo "$(basename "$0") images                                           Rebases the component images to the downloaded release"
    echo "$(basename "$0") manifests                                        Rebases the component manifests to the downloaded release"
    exit 1
}

check_preconditions

command=${1:-help}
case "$command" in
    to)
        [[ $# -lt 3 ]] && usage
        rebase_to "$2" "$3"
        ;;
    download)
        #[[ $# -lt 3 ]] && usage
        # download_release "$2" "$3"
        download_cert_manager "$2"
        ;;
    images)
        update_cert_manager_images
        ;;

    manifests)
        copy_manifests
        update_cert_manager_manifests
        ;;
    *) usage;;
esac
