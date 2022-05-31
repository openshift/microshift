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

# Reads go.mod file $1 and prints lines in its section $2 ("require" or "replace")
extract_section() {
    file=$1
    section=$2

    awk "BEGIN {output=0} /^)/ {output=0} {if (output == 1 && !match(\$0,\"// indirect\")) print \$0} /^${section}/ {output=1}" "${file}"
}

# Returns everything but the version of a require or replace line
get_mod() {
    line=$1

    re="^(.+) ([a-z0-9.+-]+)$"
    if [[ "$line" =~ $re ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        echo ""
    fi
}

# Returns the version of a require or replace line
get_version() {
    line=$1

    re="^(.+) ([a-z0-9.+-]+)$"
    if [[ "$line" =~ $re ]]; then
        echo "${BASH_REMATCH[2]}"
    else
        echo ""
    fi
}

# For every line in $base_file (require or replace), checks wether a corresponding module in
# $update_file exists and is of newer version. If so, prints a module line with the newer version,
# else prints the line in the $base_file.
update_versions() {
    base_file=$1
    update_file=$2

    re="^(.+) ([a-z0-9.-]+)$"
    while IFS="" read -r line || [ -n "$line" ]
    do
        if [[ "${line}" =~ ^//.* ]]; then
            continue
        fi

        mod="$(get_mod "${line}")"
        base_version=$(get_version "${line}")
        version=${base_version}

        update_line=$(grep "${mod} " "${update_file}" || true)
        if [[ -n "${update_line}" ]]; then
            update_version=$(get_version "${update_line}")
            version=$(printf '%s\n%s\n' "${base_version}" "${update_version}" | sort --version-sort | tail -n 1)
        fi

        echo "${mod} ${version}"
    done < "${base_file}"
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
    echo "${content}" | awk '/^# 0000_[A-Za-z0-9._-]*.yaml/{filename = $2;}{if (filename != "") print >filename;}'
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


update_go_mod() {
    if [ ! -f "${STAGING_DIR}/release.txt" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Rebasing go.mod..."
    extract_section "${REPOROOT}/go.mod" require > latest_require
    extract_section "${REPOROOT}/go.mod" replace > latest_replace
    while IFS="" read -r line || [ -n "$line" ]
    do
        COMPONENT=$(echo "${line}" | cut -d ' ' -f 1)
        REPO=$(echo "${line}" | cut -d ' ' -f 2)
        if [[ "${EMBEDDED_COMPONENTS}" == *"${COMPONENT}"* ]]; then
            extract_section "${REPO##*/}/go.mod" require > require
            extract_section "${REPO##*/}/go.mod" replace > replace
            update_versions latest_require require > t; mv t latest_require
            update_versions latest_replace replace > t; mv t latest_replace
        fi
    done < source-commits

    cat << EOF > "${REPOROOT}/go.mod"
module github.com/openshift/microshift

go 1.16

replace (
$(cat latest_replace)
)

require (
$(cat latest_require)
)
EOF

    go mod tidy
    go mod vendor

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
