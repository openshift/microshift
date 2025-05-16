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
PULL_SECRET_FILE="${HOME}/.pull-secret.json"
GO_MOD_DIRS=("$REPOROOT/" "$REPOROOT/etcd")

EMBEDDED_COMPONENTS="route-controller-manager cluster-policy-controller hyperkube etcd kube-storage-version-migrator cluster-config-api"
EMBEDDED_COMPONENT_OPERATORS="cluster-kube-apiserver-operator cluster-kube-controller-manager-operator cluster-openshift-controller-manager-operator cluster-kube-scheduler-operator machine-config-operator operator-lifecycle-manager"
LOADED_COMPONENTS="cluster-dns-operator cluster-ingress-operator service-ca-operator cluster-network-operator cluster-csi-snapshot-controller-operator"
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
    clone_repo "https://github.com/openshift/cert-manager-operator" "v1.13.1" "."
}

# Updates the image digests in pkg/release/release*.go
update_images() {
    if [ ! -f "${STAGING_DIR}/release_amd64.json" ] || [ ! -f "${STAGING_DIR}/release_arm64.json" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

   
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
        # FIXME: Allow more than 3 arguments until pipelines stop passing LVMS value.
        [[ $# -lt 3 ]] && usage
        rebase_to "$2" "$3"
        ;;
    download)
        # FIXME: Allow more than 3 arguments until pipelines stop passing LVMS value.
        #[[ $# -lt 3 ]] && usage
        # download_release "$2" "$3"
        download_cert_manager
        ;;
    images)
        update_images
        ;;

    manifests)
        copy_manifests
        update_cert_manager_manifests
        ;;
    *) usage;;
esac
