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
PULL_SECRET_FILE="${HOME}/.pull-secret.json"
declare -a ARCHS=("amd64" "arm64")
declare -A GOARCH_TO_UNAME_MAP=( ["amd64"]="x86_64" ["arm64"]="aarch64" )

VER_FILE="${STAGING_DIR}/lvms/version"

dump_version(){
    local version="${1}"
    echo "${version}" > "${VER_FILE}"
}

load_version(){
    if [ -f "${VER_FILE}" ]; then
        cat "${VER_FILE}"
    else
        >&2 echo "error: version file not found at ${VER_FILE}"
    fi
}

parse_version(){
    local image_tag="$1"
    local v
    v="${image_tag#*:}"
    if [ -z "${v}" ]; then
        >&2 echo "error: version not found in image tag ${image_tag}"
        return 1
    fi
}

title() {
    echo -e "\E[34m$1\E[00m";
}

check_preconditions() {
    if ! hash yq; then
        title "Installing yq"

        local YQ_VER=4.26.1
        # shellcheck disable=SC2034  # appears unused
        local YQ_HASH_amd64=9e35b817e7cdc358c1fcd8498f3872db169c3303b61645cc1faf972990f37582
        # shellcheck disable=SC2034  # appears unused
        local YQ_HASH_arm64=8966f9698a9bc321eae6745ffc5129b5e1b509017d3f710ee0eccec4f5568766
        local YQ_HASH
        YQ_HASH="YQ_HASH_$(go env GOARCH)"
        local YQ_URL
        YQ_URL="https://github.com/mikefarah/yq/releases/download/v${YQ_VER}/yq_linux_$(go env GOARCH)"
        local YQ_EXE
        YQ_EXE=$(mktemp /tmp/yq-exe.XXXXX)
        local YQ_SUM
        YQ_SUM=$(mktemp /tmp/yq-sum.XXXXX)
        echo -n "${!YQ_HASH} -" > "${YQ_SUM}"
        if ! (curl -Ls "${YQ_URL}" | tee "${YQ_EXE}" | sha256sum -c "${YQ_SUM}" &>/dev/null); then
            echo "ERROR: Expected file at ${YQ_URL} to have checksum ${!YQ_HASH} but instead got $(sha256sum <"${YQ_EXE}" | cut -d' ' -f1)"
            exit 1
        fi
        chmod +x "${YQ_EXE}" && sudo cp "${YQ_EXE}" /usr/bin/yq
        rm -f "${YQ_EXE}" "${YQ_SUM}"
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

# Runs each LVMS rebase step in sequence, commiting the step's output to git
rebase_lvms_to() {
    local lvms_operator_bundle_manifest="$1"

    title "# Rebasing LVMS to ${lvms_operator_bundle_manifest}"

    download_lvms_operator_bundle_manifest "${lvms_operator_bundle_manifest}"

    # LVMS image names may include `/` and `:`, which make messy branch names.
    rebase_branch="rebase-lvms-${lvms_operator_bundle_manifest//[:\/]/-}"
    git branch -D "${rebase_branch}" || true
    git checkout -b "${rebase_branch}"

    update_last_lvms_rebase "${lvms_operator_bundle_manifest}"

    update_lvms_images
    if [[ -n "$(git status -s pkg/release)" ]]; then
        title "## Committing changes to pkg/release"
        git add pkg/release
        git commit -m "update LVMS images"
    else
        echo "No changes in LVMS images."
    fi

    update_lvms_manifests "${lvms_operator_bundle_manifest}"
    if [[ -n "$(git status -s assets)" ]]; then
        title "## Committing changes to assets and pkg/assets"
        git add assets pkg/assets
        git commit -m "update LVMS manifests"
    else
        echo "No changes to LVMS assets."
    fi

    title "# Removing staging directory"
    rm -rf "${STAGING_DIR}"
}

# create a function to generate a k8s config map with a key-value pair containing a version string
# shellcheck disable=SC2034  # appears unused
generate_version_config_map() {
    local version="$1"
    local name="$2"
    local namespace="$3"

    cat <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: ${name}
  namespace: ${namespace}
data:
  version: ${version}
EOF
}

# LVMS is not integrated into the ocp release image, so the work flow does not fit with core component rebase.  LVMS'
# operator bundle is the authoritative source for manifest and image digests.
download_lvms_operator_bundle_manifest(){
    bundle_manifest="$1"

    title "downloading LVMS operator bundles ${bundle_manifest}"
    local LVMS_STAGING="${STAGING_DIR}/lvms"
    rm -rf "${LVMS_STAGING}"
    mkdir -p "${LVMS_STAGING}"

    # Persist the version of the LVMS operator bundle for use in manifest steps
    local version
    version="$(parse_version "${bundle_manifest}")"
    dump_version "${version}"

    # Push the configMap to the kube-public namespace so that it is available to all users/apps
    generate_version_config_map "${version}" "lvms-version" "kube-public"\
        > "${REPOROOT}/assets/components/lvms/topolvm-configmap_lvms-version.yaml"

    authentication=""
    if [ -f "${PULL_SECRET_FILE}" ]; then
        authentication="--registry-config ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
    fi

    for arch in "${ARCHS[@]}"; do
        mkdir -p "${LVMS_STAGING}/${arch}"
        pushd "${LVMS_STAGING}/${arch}" || return 1
        title "extracting lvms operator bundle for \"${arch}\" architecture"
        # shellcheck disable=SC2086  # Double quote to prevent globbing and word splitting.
        oc image extract \
            ${authentication} \
            --path /manifests/:. "${bundle_manifest}" \
            --filter-by-os "${arch}" \
            ||  {
                    popd
                    return 1
                }

        local csv="lvms-operator.clusterserviceversion.yaml"
        local namespace="openshift-storage"
        extract_lvms_rbac_from_cluster_service_version "${PWD}" "${csv}" "${namespace}"

        popd || return 1
    done
}

write_lvms_images_for_arch(){
    local arch="$1"
    arch_dir="${STAGING_DIR}/lvms/${arch}"
    [ -d "${arch_dir}" ] || {
        echo "dir ${arch_dir} not found"
        return 1
    }

    declare -a include_images=(
        "topolvm-csi"
        "topolvm-csi-provisioner"
        "topolvm-csi-resizer"
        "topolvm-csi-registrar"
        "topolvm-csi-livenessprobe"
    )

    local csv_manifest="${arch_dir}/lvms-operator.clusterserviceversion.yaml"
    local image_file="${arch_dir}/images"

    parse_images "${csv_manifest}" "${image_file}"

    if [ "$(wc -l "${image_file}" | cut -d' ' -f1)" -eq 0 ]; then
        >&2 echo "error: image file (${image_file}) has fewer images than expected (${#include_images})"
        exit 1
    fi
    while read -ers LINE; do
        name=${LINE%,*}
        img=${LINE#*,}
        for included in "${include_images[@]}"; do
            if [[ "${name}" == "${included}" ]]; then
                name="$(echo "${name}" | tr '-' '_')"
                yq -iP -o=json e '.images["'"${name}"'"] = "'"${img}"'"' "${REPOROOT}/assets/release/release-${GOARCH_TO_UNAME_MAP[${arch}]}.json"
                break;
            fi
        done
    done < "${image_file}"
}

update_lvms_images(){
    title "Updating LVMS images"

    local workdir="${STAGING_DIR}/lvms"
    [ -d "${workdir}" ] || {
        >&2 echo 'lvms staging dir not found, aborting image update'
        return 1
    }
    pushd "${workdir}"
    for arch in "${ARCHS[@]}"; do
        write_lvms_images_for_arch "${arch}"
    done
    popd
}

update_lvms_manifests() {
    title "Copying LVMS manifests"
    local workdir="${STAGING_DIR}/lvms"
    [ -d "${workdir}" ] || {
        >&2 echo 'lvms staging dir not found, aborting asset update'
        return 1
    }
    "${REPOROOT}/scripts/auto-rebase/handle_assets.py" ./scripts/auto-rebase/lvms_assets.yaml

    local version
    version="$(load_version)"
    generate_version_config_map "${version}" "lvms-version" "openshift-storage"\
        > "${REPOROOT}/assets/components/lvms/topolvm-configmap_lvms-version.yaml"
}

update_last_lvms_rebase() {
    local lvms_operator_bundle_manifest="$1"

    title "## Updating last_lvms_rebase.sh"

    local last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_lvms_rebase.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase-lvms.sh to "${lvms_operator_bundle_manifest}"
EOF
    chmod +x "${last_rebase_script}"

    (cd "${REPOROOT}" && \
         if test -n "$(git status -s scripts/auto-rebase/last_lvms_rebase.sh)"; then \
             title "## Committing changes to last_lvms_rebase.sh" && \
             git add scripts/auto-rebase/last_lvms_rebase.sh && \
             git commit -m "update last_lvms_rebase.sh"; \
         fi)
}


# In the ClusterServiceVersion there are encoded RBAC information for OLM deployments.
# Since microshift skips this installation and uses a custom one based on the bundle, we have to extract the RBAC
# manifests from the CSV by reading them out into separate files.
# shellcheck disable=SC2207
extract_lvms_rbac_from_cluster_service_version() {
  local dest="$1"
  local csv="$2"
  local namespace="$3"

  title "extracting lvms clusterserviceversion.yaml into separate RBAC"

  local clusterPermissions=($(yq eval '.spec.install.spec.clusterPermissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${clusterPermissions[@]}"; do
    echo "extracting bundle .spec.install.spec.clusterPermissions by serviceAccountName ${service_account_name}"

    local clusterrole="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_clusterrole.yaml"
    echo "generating ${clusterrole}"
    extract_lvms_clusterrole_from_csv_by_service_account_name "${service_account_name}" "${csv}" "${clusterrole}"

    local clusterrolebinding="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml"
    echo "generating ${clusterrolebinding}"
    extract_lvms_clusterrolebinding_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${clusterrolebinding}"

    local service_account="${dest}/${service_account_name}_v1_serviceaccount.yaml"
    echo "generating ${service_account}"
    extract_lvms_service_account_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${service_account}"
  done

  local permissions=($(yq eval '.spec.install.spec.permissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${permissions[@]}"; do
    echo "extracting bundle .spec.install.spec.permissions by serviceAccountName ${service_account_name}"

    local role="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_role.yaml"
    echo "generating ${role}"
    extract_lvms_role_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${csv}" "${role}"

    local rolebinding="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_rolebinding.yaml"
    echo "generating ${rolebinding}"
    extract_lvms_rolebinding_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${rolebinding}"

    local service_account="${dest}/${service_account_name}_v1_serviceaccount.yaml"
    echo "generating ${service_account}"
    extract_lvms_service_account_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${service_account}"
  done
}

extract_lvms_clusterrole_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local csv="$2"
  local target="$3"
  yq eval "
    .spec.install.spec.clusterPermissions[] |
    select(.serviceAccountName == \"${service_account_name}\") |
    .apiVersion = \"rbac.authorization.k8s.io/v1\" |
    .kind = \"ClusterRole\" |
    .metadata.name = \"${service_account_name}\" |
    del(.serviceAccountName)
    " "${csv}" > "${target}"
}

extract_lvms_role_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local csv="$3"
  local target="$4"
  yq eval "
    .spec.install.spec.permissions[] |
    select(.serviceAccountName == \"${service_account_name}\") |
    .apiVersion = \"rbac.authorization.k8s.io/v1\" |
    .kind = \"Role\" |
    .metadata.name = \"${service_account_name}\" |
    .metadata.namespace = \"${namespace}\" |
    del(.serviceAccountName)
    " "${csv}" > "${target}"
}

extract_lvms_clusterrolebinding_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local target="$3"

  crb=$(cat <<EOL
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: ${service_account_name}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: ${service_account_name}
subjects:
- kind: ServiceAccount
  name: ${service_account_name}
  namespace: ${namespace}
EOL
)
  echo "${crb}" > "${target}"
}

extract_lvms_rolebinding_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local target="$3"

  crb=$(cat <<EOL
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: ${service_account_name}
  namespace: ${namespace}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: Role
  name: ${service_account_name}
  namespace: ${namespace}
subjects:
- kind: ServiceAccount
  name: ${service_account_name}
  namespace: ${namespace}
EOL
)
  echo "${crb}" > "${target}"
}

extract_lvms_service_account_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local target="$3"

  serviceAccount=$(cat <<EOL
apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: null
  name: ${service_account_name}
  namespace: ${namespace}
EOL
)
    echo "${serviceAccount}" > "${target}"
}

parse_images() {
    local src="$1"
    local dest="$2"
    yq '.spec.relatedImages[]? | [.name, .image] | @csv' "${src}" > "${dest}"
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to LVMS_RELEASE_IMAGE         Performs all the steps to rebase LVMS"
    echo "$(basename "$0") download LVMS_RELEASE_IMAGE   Downloads the content of a LVMS release image to disk in preparation for rebasing"
    echo "$(basename "$0") images                        Updates LVMS images"
    echo "$(basename "$0") manifests                     Updates LVMS manifests"
    exit 1
}

check_preconditions

command=${1:-help}
case "${command}" in
    to)
        rebase_lvms_to "$2"
        ;;
    download)
        download_lvms_operator_bundle_manifest "$2"
        ;;
    images)
        update_lvms_images
        ;;
    manifests)
        update_lvms_manifests
        ;;
    *) usage;;
esac
