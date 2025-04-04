#!/usr/bin/env bash

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

shopt -s expand_aliases
shopt -s extglob

export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../..")"
STAGING_DIR="${REPOROOT}/_output/staging"
FLANNEL_VERSION=v0.25.6
FLANNEL_CNI_PLUGIN_VERSION=v1.5.1-flannel2
REGISTRY="docker.io/flannel"
declare -A GOARCH_TO_UNAME_MAP=( ["amd64"]="x86_64" ["arm64"]="aarch64" )
declare -A FLANNEL_MAP=( ["flannel"]="flannel_release.json" ["flannel-plugin"]="flannel-plugin_release.json")

title() {
    echo -e "\E[34m$1\E[00m";
}

get_manifest_from_docker() {
      title "Downloading manifest for flannel"
      mkdir -p "${STAGING_DIR}"
      pushd "${STAGING_DIR}" >/dev/null
      oc image info ${REGISTRY}/flannel:${FLANNEL_VERSION} --show-multiarch -ojson > "${FLANNEL_MAP["flannel"]}"
      oc image info ${REGISTRY}/flannel-cni-plugin:${FLANNEL_CNI_PLUGIN_VERSION} --show-multiarch -ojson > "${FLANNEL_MAP["flannel-plugin"]}"
      popd >/dev/null
}


update_flannel_images() {
      title "Rebasing flannel images"

      for goarch in amd64 arm64; do
          arch=${GOARCH_TO_UNAME_MAP["${goarch}"]:-noarch}

          local kustomization_arch_file="${REPOROOT}/assets/optional/flannel/kustomization.${arch}.yaml"
          local flannel_release_json="${REPOROOT}/assets/optional/flannel/release-flannel-${arch}.json"

          jq -n "{\"release\": {\"base\": \"${REGISTRY}\"}, \"images\": {}}" > "${flannel_release_json}"

        # Create extra kustomization for each arch in separate file.
        # Right file (depending on arch) should be appended during rpmbuild to kustomization.yaml.
        cat <<EOF > "${kustomization_arch_file}"

images:
EOF

        for container in flannel flannel-plugin; do
            new_image_name=$(jq -r ".[] | select(.config.architecture == \"${goarch}\") | .name" "${STAGING_DIR}/${FLANNEL_MAP[${container}]}")
            new_image_digest=$(jq -r ".[] | select(.config.architecture == \"${goarch}\") | .digest" "${STAGING_DIR}/${FLANNEL_MAP[${container}]}")
            local new_image="${new_image_name%%:*}@${new_image_digest}"

            cat <<EOF >> "${kustomization_arch_file}"
  - name: ${container}
    newName: ${new_image_name%%:*}
    digest: ${new_image_digest}
EOF

            yq -i -o json ".images += {\"${container}\": \"${new_image}\"}" "${flannel_release_json}"
        done  # for container
    done  # for goarch
}

get_manifest_from_docker
update_flannel_images
