#! /usr/bin/env bash
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
set -o nounset
set -o pipefail

shopt -s expand_aliases
shopt -s extglob

# debugging options
#trap 'echo "# $BASH_COMMAND"' DEBUG
#set -x

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/..")"
STAGING_DIR="$REPOROOT/_output/staging"
PULL_SECRET_FILE="${HOME}/.docker/config.json"

EMBEDDED_COMPONENTS="openshift-apiserver openshift-controller-manager oauth-apiserver hyperkube etcd"
EMBEDDED_COMPONENT_OPERATORS="cluster-kube-apiserver-operator cluster-openshift-apiserver-operator cluster-kube-controller-manager-operator cluster-openshift-controller-manager-operator cluster-kube-scheduler-operator machine-config-operator"
LOADED_COMPONENTS="cluster-dns-operator cluster-ingress-operator service-ca-operator"


title() {
    echo -e "\E[34m$1\E[00m";
}


# Returns the list of release image names from a release_${arch}.go file
get_release_images() {
    file=$1

    awk "BEGIN {output=0} /^}/ {output=0} {if (output == 1) print substr(\$1, 2, length(\$1)-3)} /^var Image/ {output=1}" "${file}"
}


# Downloads a release's tools and manifest content into a staging directory,
# then checks out the required components for the rebase at the release's commit.
download_release() {
    local release_image=$1

    rm -rf "${STAGING_DIR}"
    mkdir -p "${STAGING_DIR}"
    pushd "${STAGING_DIR}" >/dev/null

    authentication=""
    if [ -f "${PULL_SECRET_FILE}" ]; then
        authentication="-a ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
    fi

    title "# Downloading and extracting ${release_image} tools"
    oc adm release extract ${authentication} --tools "${release_image}"

    title "# Extracing ${release_image} manifest content"
    mkdir -p release-manifests
    pushd release-manifests >/dev/null
    content=$(oc adm release info ${authentication} --contents "${release_image}")
    echo "${content}" | awk '{ if ($0 ~ /^# [A-Za-z0-9._-]+.yaml$/ || $0 ~ /^# image-references$/ || $0 ~ /^# release-metadata$/) filename = $2; else print >filename;}'
    popd >/dev/null

    title "# Cloning ${release_image} component repos"
    commits=$(oc adm release info ${authentication} --commits -o json "${release_image}")
    echo "${commits}" | jq -r '.references.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location") \(.annotations."io.openshift.build.commit.id")"' > source-commits

    git config --global advice.detachedHead false
    while IFS="" read -r line || [ -n "$line" ]
    do
        component=$(echo "${line}" | cut -d ' ' -f 1)
        repo=$(echo "${line}" | cut -d ' ' -f 2)
        commit=$(echo "${line}" | cut -d ' ' -f 3)
        if [[ "${EMBEDDED_COMPONENTS}" == *"${component}"* ]] || [[ "${LOADED_COMPONENTS}" == *"${component}"* ]] || [[ "${EMBEDDED_COMPONENT_OPERATORS}" == *"${component}"* ]]; then
            title "## Cloning ${repo} at commit ${commit}..."
            git clone "${repo}"
            pushd "${repo##*/}" >/dev/null
            git checkout "${commit}"
            popd >/dev/null
            echo
        fi
    done < source-commits

    popd >/dev/null
}

get_pseudoversion() {
    local modulepath=$1
    local component=$2

    version=$(echo "${modulepath}" | grep -o "v[0-9]$")
    if [ -z "${version}" ]; then
        version="v0.0.0"
    else
        version="${version}.0.0"
    fi
    timestamp_commit=$( cd "${STAGING_DIR}/${component}" && TZ=UTC git --no-pager show --quiet --abbrev=12 --date='format-local:%Y%m%d%H%M%S' --format="%cd-%h" )
    echo "${version}-${timestamp_commit}"
}

#
get_modulepath_version_for_release() {
    local component=$1

    repo=$( cd "${STAGING_DIR}/${component}" && git config --get remote.origin.url )
    modulepath="${repo#https://}"
    echo "${modulepath}@$(get_pseudoversion ${modulepath} ${component})"
}

# Returns the line (including trailing comment) in the #{gomod_file} containing the ReplaceDirective for ${module_path}
get_replace_directive() {
    local gomod_file=$1
    local module_path=$2

    # TODO: Handle special case of keyword "replace" being included in the line
    go mod edit -print "${gomod_file}" | grep "^[[:space:]]${module_path}[[:space:]][[:alnum:][:space:].-]*=>"
}

lookup_modulepath_version_from_component() {
    local modulepath=$1
    local component=$2

    # special-case etcd
    if [[ "${modulepath}" =~ ^go.etcd.io/etcd/ ]]; then
        modulepath=$(echo "${modulepath}" | sed 's|^go.etcd.io/etcd|github.com/openshift/etcd|')
        pseudoversion=$(get_pseudoversion ${modulepath} etcd)
        echo "${modulepath}@${pseudoversion}"
        return
    fi

    replace_directive=$(get_replace_directive "${STAGING_DIR}/${component}/go.mod" "${modulepath}")
    replace_directive=$(strip_comment "${replace_directive}")
    replacement=$(echo "${replace_directive}" | sed -E "s|.*=>[[:space:]]*(.*)[[:space:]]*|\1|")
    if [[ "${replacement}" =~ ^./staging ]]; then
        replacement=$(echo "${replacement}" | sed 's|^./staging/|github.com/openshift/kubernetes/staging/|')
        replacement="${replacement} $(get_pseudoversion ${modulepath} kubernetes)"
    fi
    echo "${replacement}" | sed 's| |@|'
}

# Returns ${line} without comment
strip_comment() {
    local line=$1

    echo "${line%%//*}"
}

# Returns the comment in ${line} if one exists or an empty string if not
get_comment() {
    local line=$1

    comment=${line##*//}
    if [ "${comment}" != "${line}" ]; then
        echo ${comment}
    else
        echo ""
    fi
}

valid_component_or_exit() {
    local component=$1
    if [[ ! " etcd kubernetes openshift-apiserver openshift-controller-manager " =~ " ${component} " ]]; then
        echo "error: release reference must be one of [etcd kubernetes openshift-apiserver openshift-controller-manager], have ${component}"
        exit 1
    fi
}

update_go_mod() {
    pushd "${STAGING_DIR}" >/dev/null

    title "# Updating go.mod"

    replaced_modulepaths=$(go mod edit -json | jq -r '.Replace // []' | jq -r '.[].Old.Path' | xargs)
    for modulepath in ${replaced_modulepaths}; do
        current_replace_directive=$(get_replace_directive "${REPOROOT}/go.mod" "${modulepath}")
        comment=$(get_comment "${current_replace_directive}")
        command=${comment%% *}
        arguments=${comment#${command} }
        case "${command}" in
        from)
            component=${arguments%% *}
            valid_component_or_exit "${component}"
            new_modulepath_version=$(lookup_modulepath_version_from_component "${modulepath}" "${component}")
            go mod edit -replace ${modulepath}=${new_modulepath_version}
            ;;
        release)
            component=${arguments%% *}
            valid_component_or_exit "${component}"
            new_modulepath_version=$(get_modulepath_version_for_release "${component}")
            go mod edit -replace ${modulepath}=${new_modulepath_version}
            ;;
        override)
            echo "skipping modulepath ${modulepath}: override [${arguments}]"
            ;;
        *)
            echo "skipping modulepath ${modulepath}: no or unknown command [${comment}]"
            ;;
        esac
    done

    popd >/dev/null
}

regenerate_openapi() {
    pushd "${STAGING_DIR}/kubernetes" >/dev/null

    title "Regenerating kube OpenAPI"
    make gen_openapi
    cp ./pkg/generated/openapi/zz_generated.openapi.go "${REPOROOT}/vendor/k8s.io/kubernetes/pkg/generated/openapi"

    popd >/dev/null
}


# Updates the image digests in pkg/release/release*.go
update_images() {
    if [ ! -f "${STAGING_DIR}/release.txt" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Rebasing release_*.go"
    base_release=$(grep -o -P "^[[:space:]]+Version:[[:space:]]+\K([[:alnum:].-]+)" "${STAGING_DIR}"/release.txt)

    images="$(get_release_images "${REPOROOT}/pkg/release/release.go" | xargs)"

    for arch in amd64; do
        w=$(awk "BEGIN {n=split(\"${images}\", images, \" \"); max=0; for (i=1;i<=n;i++) {if (length(images[i]) > max) {max=length(images[i])}}; print max+2; exit}")
        for i in ${images}; do
            digest=$(awk "/ ${i//_/-} / {print \$2}" release.txt)
            if [[ -n "${digest}" ]]; then
                awk "!/\"${i}\"/ {print \$0} /\"${i}\"/ {printf(\"\\t\\t%-${w}s  %s\n\", \"\\\"${i}\\\":\", \"\\\"${digest}\\\",\")}" \
                    "${REPOROOT}/pkg/release/release_${arch}.go" > t
                mv t "${REPOROOT}/pkg/release/release_${arch}.go"
            fi
        done
    done

    sed -i "/^var Base/c\var Base = \"${base_release}\"" "${REPOROOT}/pkg/release/release.go"

    popd >/dev/null
}


# Updates embedded component manifests
update_manifests() {
    if [ ! -f "${STAGING_DIR}/release.txt" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Rebasing manifests"
    for crd in ${REPOROOT}/assets/crd/*.yaml; do
        cp "${STAGING_DIR}"/release-manifests/$(basename ${crd}) "${REPOROOT}"/assets/crd || true
    done
    rm -f "${REPOROOT}"/assets/scc/*.yaml
    cp "${STAGING_DIR}"/release-manifests/0000_20_kube-apiserver-operator_00_scc-*.yaml "${REPOROOT}"/assets/scc || true

    rm -f "${REPOROOT}"/assets/components/openshift-dns/dns/*
    cp "${STAGING_DIR}"/cluster-dns-operator/assets/dns/* "${REPOROOT}"/assets/components/openshift-dns/dns 2>/dev/null || true 
    rm -f "${REPOROOT}"/assets/components/openshift-dns/node-resolver/*
    cp "${STAGING_DIR}/"cluster-dns-operator/assets/node-resolver/* "${REPOROOT}"/assets/components/openshift-dns/node-resolver 2>/dev/null || true
    rm -f "${REPOROOT}"/assets/components/openshift-router/*
    cp "${STAGING_DIR}"/cluster-ingress-operator/assets/router/* "${REPOROOT}"/assets/components/openshift-router 2>/dev/null || true
    rm -f "${REPOROOT}"/assets/components/service-ca/*
    cp "${STAGING_DIR}"/service-ca-operator/bindata/v4.0.0/controller/* "${REPOROOT}"/assets/components/service-ca 2>/dev/null || true

    popd >/dev/null
}


usage() {
    echo "Usage:"
    echo "$(basename "$0") download RELEASE_IMAGE"
    echo "$(basename "$0") go.mod"
    echo "$(basename "$0") generated-apis"
    echo "$(basename "$0") images"
    echo "$(basename "$0") manifests"
    exit 1
}

command=${1:-help}
case "$command" in
    download)
        [[ $# -ne 2 ]] && usage
        download_release "$2"
        ;;
    go.mod) update_go_mod;;
    generated-apis) regenerate_openapi;;
    images) update_images;;
    manifests) update_manifests;;
    *) usage;;
esac
