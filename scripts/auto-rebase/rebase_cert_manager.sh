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
BIN_DIR="$REPOROOT/_output/bin"
export REGISTRY_AUTH_FILE="${HOME}/.pull-secret.json"
OPERATOR_INDEX="${STAGING_DIR}/redhat-operator-index.yaml"
OPERATOR_CERT_MANAGER_INDEX="${STAGING_DIR}/redhat-operator-cert-manager-index.yaml"
GO_MOD_DIRS=("$REPOROOT/" "$REPOROOT/etcd")

PULL_SECRET_FILE="${HOME}/.pull-secret.json"

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
        DEST_DIR="${BIN_DIR}" "${REPOROOT}/scripts/fetch_tools.sh" opm
    fi

    if [[ ! -f "${HOME}/.config/containers/policy.json" && ! -f /etc/containers/policy.json ]]; then
        echo "Could not find ${HOME}/.config/containers/policy.json or /etc/containers/policy.json - creating"
        # If these two files do not exist, we assume the script is running within the CI container.
        # These are required for the `opm` command to work correctly.
        #
        mkdir -p "${HOME}/.config/containers/"
        cat << EOF > "${HOME}/.config/containers/policy.json"
{
    "default": [
        {
            "type": "insecureAcceptAnything"
        }
    ],
    "transports":
        {
            "docker-daemon":
                {
                    "": [{"type":"insecureAcceptAnything"}]
                }
        }
}
EOF
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

    operator_manifest="$1"

    # opm needs credentials to pull the image, but there's no flag like -a or --authfile to provide our ~/.pull-secret.json
    clear_docker_config=false
    if [[ -f "${PULL_SECRET_FILE}" && ! -f "${HOME}/.docker/config.json" ]]; then
        mkdir -p "${HOME}/.docker/"
        ln -s "${PULL_SECRET_FILE}" "${HOME}/.docker/config.json"
        clear_docker_config=true
    fi

    # get the whole operator yaml
    ${BIN_DIR}/opm render "${operator_manifest}" -o yaml  >${OPERATOR_INDEX}

    if "${clear_docker_config}"; then
        rm "${HOME}/.docker/config.json"
    fi

    # find the latest published cert-manager-operator ie: cert-manager-operator.v1.16.0
    export operator=$(yq 'select(.package == "openshift-cert-manager-operator" and .name == "stable-v1") | .entries[-1].name' ${OPERATOR_INDEX})
    yq 'select (.name==env(operator))' ${OPERATOR_INDEX} >"${OPERATOR_CERT_MANAGER_INDEX}"

    echo  "found operator version ${operator}"

    # convert from cert-manager-operator.v1.16.0 to cert-manager-x.y
    branch_name=$(echo ${operator} | awk -F'[^0-9]*' '{print "cert-manager-"$2"."$3}')
    clone_repo "https://github.com/openshift/cert-manager-operator" "$branch_name" "."
    popd

}

# helper to update ConfigMap with image references in images.yaml
update_configmap_image() {
    local component_name="$1"
    local full_image_ref="$2"
    local configmap_file="$3"
    yq -i ".data.\"${component_name}\" = \"${full_image_ref}\"" "${configmap_file}"
}

# helper to update environment variables with image references in manager.yaml
update_deploymentenv_manager() {
    local component_name="$1"
    local full_image_ref="$2"
    local manager_file="$3"

    # convert component names to uppercase and replace hyphens with underscores
    component_name="${component_name^^}"
    component_name="${component_name//-/_}"
    sed -i "/name: RELATED_IMAGE_${component_name}/ {n;s|value: .*|value: ${full_image_ref}|;}" "${manager_file}"
}

write_cert_manager_images_for_arch() {
    local arch="$1"
    local arch_digest

    title "Updating images for ${arch}"
    local cert_manager_release_json="${REPOROOT}/assets/optional/cert-manager/release-cert-manager-${GOARCH_TO_UNAME_MAP[${arch}]}.json"
    local cert_manager_operator_yaml="${REPOROOT}/assets/optional/cert-manager/manager/manager-${GOARCH_TO_UNAME_MAP[${arch}]}.yaml"
    local cert_manager_images_yaml="${REPOROOT}/assets/optional/cert-manager/manager/images-${GOARCH_TO_UNAME_MAP[${arch}]}.yaml"

    local operatorVersion=$(yq '.properties[] | select(.type == "olm.package").value.version' "${OPERATOR_CERT_MANAGER_INDEX}")

    jq -n "{\"release\": {\"base\": \"${operatorVersion}\"}, \"images\": {}}" > "${cert_manager_release_json}"

    # controllerImage
    local operatorImageFull=$(yq '.properties[] | select(.type == "olm.csv.metadata").value.annotations.containerImage' "${OPERATOR_CERT_MANAGER_INDEX}")
    arch_digest=$(oc -a "${PULL_SECRET_FILE}" image info -o json --filter-by-os "linux/${arch}" "${operatorImageFull}" | jq -r '.digest')
    operatorImageFull="${operatorImageFull%@*}@${arch_digest}"

    local operatorImage="${operatorImageFull%:*}"
    local operatorTag="${operatorImageFull#*:}"

    yq -i -o json ".images += {\"cert-manager-operator\": \"${operatorImageFull}\"}" "${cert_manager_release_json}"

    # update controller image in ConfigMap
    update_configmap_image "controller" "${operatorImageFull}" "${cert_manager_images_yaml}"

    # relatedImages
    for index in $(yq '.relatedImages.[] | path | .[-1] ' "${OPERATOR_CERT_MANAGER_INDEX}"); do
        local image=$(yq ".relatedImages.${index}.image" "${OPERATOR_CERT_MANAGER_INDEX}" )
        arch_digest=$(oc -a "${PULL_SECRET_FILE}" image info -o json --filter-by-os "linux/${arch}" "${image}" | jq -r '.digest')
        image="${image%@*}@${arch_digest}"

        local component=$(yq ".relatedImages.${index}.name" "${OPERATOR_CERT_MANAGER_INDEX}")
        if [[  -n "${component}" && "${OPERATOR_COMPONENTS}" == *"${component}"* ]]; then
            yq -i -o json ".images += {\"${component}\": \"${image}\"}" "${cert_manager_release_json}"

            # update component image in images.yaml
            update_configmap_image "${component}" "${image}" "${cert_manager_images_yaml}"

            # update component image in manager.yaml
            update_deploymentenv_manager "${component}" "${image}" "${cert_manager_operator_yaml}"
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
    title "Copying manifests"
    "$REPOROOT/scripts/auto-rebase/handle_assets.py" "./scripts/auto-rebase/assets_cert_manager.yaml"

    # Copy manager.yaml per each architecture to be updated during the rebase
    for arch in "${ARCHS[@]}"; do
        cp "${REPOROOT}/assets/optional/cert-manager/manager/manager.yaml" \
            "${REPOROOT}/assets/optional/cert-manager/manager/manager-${GOARCH_TO_UNAME_MAP[${arch}]}.yaml"
    done
}

update_last_cert_manager_rebase() {
    local cert_manager_operator_bundle="$1"

    title "## Updating last_rebase_cert_manager.sh"

    local last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase_cert_manager.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase_cert_manager.sh to "${cert_manager_operator_bundle}"
EOF
    chmod +x "${last_rebase_script}"

    (cd "${REPOROOT}" && \
         if test -n "$(git status -s scripts/auto-rebase/last_rebase_cert_manager.sh)"; then \
             title "## Committing changes to last_rebase_cert_manager.sh" && \
             git add scripts/auto-rebase/last_rebase_cert_manager.sh && \
             git commit -m "update last_rebase_cert_manager.sh"; \
         fi)
}

rebase_cert_manager_to(){
    local -r operator_bundle="${1}"
    download_cert_manager "${operator_bundle}"
    copy_manifests
    update_cert_manager_images
    update_last_cert_manager_rebase "${operator_bundle}"
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to OPM_RELEASE_IMAGE                             Performs all the steps to rebase to a release image."
    echo "$(basename "$0") download OPM_RELEASE_IMAGE                       Downloads the content of a release image to disk in preparation for rebasing."
    echo "$(basename "$0") images                                           Rebases the component images to the downloaded release"
    echo "$(basename "$0") manifests                                        Rebases the component manifests to the downloaded release"
    exit 1
}

check_preconditions

command=${1:-help}
case "$command" in
    to)
        [[ $# -lt 2 ]] && usage
        rebase_cert_manager_to "$2"
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
        ;;
    *) usage;;
esac
