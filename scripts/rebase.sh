#! /usr/bin/env bash
#   Copyright 2021 The Microshift authors
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

set -euo pipefail
shopt -s expand_aliases

# debugging options
#trap 'echo "# $BASH_COMMAND"' DEBUG
#set -x

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../")"
STAGING_DIR="$REPOROOT/_output/staging"

EMBEDDED_COMPONENTS="etcd hyperkube openshift-apiserver openshift-controller-manager"
LOADED_COMPONENTS="cluster-dns-operator cluster-ingress-operator service-ca-operator"


title() {
    echo -e "\E[34m\n$1\E[00m";
}

# Reads go.mod file $1 and prints lines in its section $2 ("require" or "replace")
extract_section() {
    file=$1
    section=$2

    cat ${file} | awk "BEGIN {output=0} /^)/ {output=0} {if (output == 1 && !match(\$0,\"// indirect\")) print \$0} /^${section}/ {output=1}"
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

        update_line=$(cat "${update_file}" | grep "${mod}" || true)
        if [[ ! -z "${update_line}" ]]; then
            update_version=$(get_version "${update_line}")
            version=$(printf '%s\n%s\n' "${base_version}" "${update_version}" | sort --version-sort | tail -n 1)
        fi

        echo "${mod} ${version}" 
    done < ${base_file}
}

# Returns the list of release image names from a release_${arch}.go file
get_release_images() {
    file=$1

    cat ${file} | awk "BEGIN {output=0} /^}/ {output=0} {if (output == 1) print substr(\$1, 2, length(\$1)-3)} /^var Image/ {output=1}"
}


# == MAIN ==
if [[ $EUID -ne 0 ]]; then
   >&2 echo "You need to run this script as root or in a `buildah unshare` environment:" 
   >&2 echo "  buildah unshare $0" 
   exit 1
fi
if [[ -z ${1+x} ]]; then
    >&2 echo "You need to provide an OKD release name, e.g.:"
    >&2 echo "  $0 4.7.0-0.okd-2021-08-22-163618"
    exit 1
fi
OKD_RELEASE=$1


rm -rf "${STAGING_DIR}"
mkdir -p "${STAGING_DIR}"
pushd "${STAGING_DIR}" >/dev/null


title "Downloading and extracting ${OKD_RELEASE} release image..."
curl -LO "https://github.com/openshift/okd/releases/download/${OKD_RELEASE}/release.txt"

OKD_RELEASE_IMAGE=$(grep -oP 'Pull From: \K[\w.-/@:]+' release.txt)
podman pull ${OKD_RELEASE_IMAGE}
cnt=$(buildah from ${OKD_RELEASE_IMAGE})
mnt=$(buildah mount ${cnt} | cut -d ' ' -f 2)
cat ${mnt}/release-manifests/image-references \
    | jq -r '.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location") \(.annotations."io.openshift.build.commit.id")"' \
    > source_commits.txt
mkdir -p "${STAGING_DIR}/release-manifests"
cp ${mnt}/release-manifests/*.yaml ${STAGING_DIR}/release-manifests


title "Cloning git repos..."
git config --global advice.detachedHead false
while IFS="" read -r line || [ -n "$line" ]
do
    COMPONENT=$(echo "${line}" | cut -d ' ' -f 1)
    REPO=$(echo "${line}" | cut -d ' ' -f 2)
    COMMIT=$(echo "${line}" | cut -d ' ' -f 3)
    if [[ ${EMBEDDED_COMPONENTS} == *"${COMPONENT}"* ]] || [[ ${LOADED_COMPONENTS} == *"${COMPONENT}"* ]]; then
        git clone ${REPO}
        pushd ${REPO##*/} >/dev/null
        git checkout ${COMMIT}
        echo
        popd >/dev/null
    fi
done < source_commits.txt


title "Rebasing go.mod..."
extract_section ${REPOROOT}/go.mod require > latest_require
extract_section ${REPOROOT}/go.mod replace > latest_replace
while IFS="" read -r line || [ -n "$line" ]
do
    COMPONENT=$(echo "${line}" | cut -d ' ' -f 1)
    REPO=$(echo "${line}" | cut -d ' ' -f 2)
    if [[ ${EMBEDDED_COMPONENTS} == *"${COMPONENT}"* ]]; then
        extract_section ${REPO##*/}/go.mod require > require
        extract_section ${REPO##*/}/go.mod replace > replace
        update_versions latest_require require > t; mv t latest_require
        update_versions latest_replace replace > t; mv t latest_replace
    fi
done < source_commits.txt

cat << EOF > ${REPOROOT}/go.mod
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


title "Rebasing release_*.go"
for arch in amd64; do
    images="$(get_release_images ${REPOROOT}/pkg/release/release_${arch}.go | xargs)"
    w=$(awk "BEGIN {n=split(\"${images}\", images, \" \"); max=0; for (i=1;i<=n;i++) {if (length(images[i]) > max) {max=length(images[i])}}; print max+2; exit}")
    for i in ${images}; do
        digest=$(cat release.txt | awk "/ ${i//_/-} / {print \$2}")
        if [[ ! -z "${digest}" ]]; then
            awk "!/\"${i}\"/ {print \$0} /\"${i}\"/ {printf(\"\\t%-${w}s  %s\n\", \"\\\"${i}\\\":\", \"\\\"${digest}\\\",\")}" \
                ${REPOROOT}/pkg/release/release_${arch}.go > t
            mv t ${REPOROOT}/pkg/release/release_${arch}.go
        fi
    done
done


title "Rebasing manifests"
assets=$(find ${REPOROOT}/assets -name "*.yaml")
for asset in ${assets}; do
    search_path=${REPOROOT}/_output/staging/release-manifests
    search_name=$(basename ${asset})
    search_exclude=XXX

    # TODO: Rename assets and their references to obviate the need for special cases
    case $(basename ${asset}) in
    0000_60_service-ca_00_roles.yaml)
        search_path=${REPOROOT}/_output/staging/service-ca-operator/bindata/v4.0.0/controller
        search_name=role.yaml
        ;;
    0000_60_service-ca_01_namespace.yaml)
        search_path=${REPOROOT}/_output/staging/service-ca-operator/bindata/v4.0.0/controller
        search_name=ns.yaml
        ;;
    0000_60_service-ca_04_sa.yaml)
        search_path=${REPOROOT}/_output/staging/service-ca-operator/bindata/v4.0.0/controller
        search_name=sa.yaml
        ;;
    0000_60_service-ca_05_deploy.yaml)
        search_path=${REPOROOT}/_output/staging/service-ca-operator/bindata/v4.0.0/controller
        search_name=deployment.yaml
        ;;
    0000_70_dns_00-*)
        search_path=${REPOROOT}/_output/staging/cluster-dns-operator/assets/dns
        search_name=${search_name#"0000_70_dns_00-"}
        search_exclude="${REPOROOT}/_output/staging/cluster-dns-operator/assets/dns/metrics/*"
        ;;
    0000_70_dns_01-*)
        search_path=${REPOROOT}/_output/staging/cluster-dns-operator/assets/dns
        search_name=${search_name#"0000_70_dns_01-"}
        search_exclude="${REPOROOT}/_output/staging/cluster-dns-operator/assets/dns/metrics/*"
        ;;
    0000_80_openshift-router-service.yaml)
        search_path=${REPOROOT}/_output/staging/cluster-ingress-operator/assets/router
        search_name=service-internal.yaml
        search_exclude="${REPOROOT}/_output/staging/cluster-dns-operator/assets/router/metrics/*"
        ;;
    0000_80_openshift-router-*)
        search_path=${REPOROOT}/_output/staging/cluster-ingress-operator/assets/router
        search_name="${search_name#"0000_80_openshift-router-"}"
        search_exclude="${REPOROOT}/_output/staging/cluster-dns-operator/assets/router/metrics/*"
        ;;
    0000_11_imageregistry-configs.crd.yaml)
        search_path=${REPOROOT}/_output/staging/openshift-apiserver/vendor/github.com/openshift/api/imageregistry/v1
        search_name=00-crd.yaml
        ;;
    esac

    updated_asset=$(find ${search_path} -name "${search_name}" -not -path "${search_exclude}" | tail -n 1)
    if [[ ! -z "${updated_asset}" ]]; then
        echo "Updating ${asset} from ${updated_asset}"
        cp "${updated_asset}" "${asset}"
    else
        echo -e "\E[31mNo update source found for ${asset}\E[00m";
    fi
done


title "Done."
popd >/dev/null
