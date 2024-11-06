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

    if ! hash oc; then
        echo "ERROR: oc is not present on the system - please install"
        exit 1
    fi
}

download_gateway_api_manifests() {
    dest="$1"
    version="$2"

    title "downloading Gateway API manifests ${version}"
    local GATEWAY_API_STAGING="${STAGING_DIR}/gateway-api"
    rm -rf "${GATEWAY_API_STAGING}"
    mkdir -p "${GATEWAY_API_STAGING}"

    crd_file="${dest}/gateway.networking.k8s.io_crds.yaml"
    oc kustomize "github.com/kubernetes-sigs/gateway-api/config/crd?ref=v${version}" > "${crd_file}"
}

# Runs each OSSM rebase step in sequence, commiting the step's output to git
rebase_ossm_to() {
    local operator_bundle_manifest="$1"
    local gateway_api_version="$2"

    title "# Rebasing OSSM to ${operator_bundle_manifest}"

    download_ossm_operator_bundle_manifest "${operator_bundle_manifest}" "${gateway_api_version}"

    # OSSM image names may include `/` and `:`, which make messy branch names.
    rebase_branch="rebase-ossm-${operator_bundle_manifest//[:\/]/-}"
    git branch -D "${rebase_branch}" || true
    git checkout -b "${rebase_branch}"

    update_last_ossm_rebase "${operator_bundle_manifest}" "${gateway_api_version}"
  
    update_ossm_manifests "${operator_bundle_manifest}"
    update_ossm_images
    if [[ -n "$(git status -s assets)" ]]; then
        title "## Committing changes to assets and pkg/assets"
        git add assets pkg/assets
        git commit -m "update OSSM manifests and images"
    else
        echo "No changes to OSSM assets."
    fi

    title "# Removing staging directory"
    rm -rf "${STAGING_DIR}"
}

# OSSM is not integrated into the ocp release image, so the workflow does not fit with core component rebase.
# Because OSSM is used without OLM, all of the manifests in the bundle are downloaded.
download_ossm_operator_bundle_manifest() {
  bundle_manifest="$1"
  gateway_api_version="$2"

  title "downloading OSSM operator bundle manifests ${bundle_manifest}"
  local OSSM_STAGING="${STAGING_DIR}/ossm"
  rm -rf "${OSSM_STAGING}"
  mkdir -p "${OSSM_STAGING}"

  # Persist the version of the OSSM operator bundle for use in manifest steps
  local version
  version=$(echo "${bundle_manifest}" | awk -F':' '{print $2}')

  title "recognized version: ${version}"

  authentication=""
  if [ -f "${PULL_SECRET_FILE}" ]; then
    authentication="--registry-config ${PULL_SECRET_FILE}"
  else
    >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
  fi

  local -r csv="servicemeshoperator3.clusterserviceversion.yaml"
  local -r namespace=$(yq '.metadata.name' "${REPOROOT}/assets/optional/gateway-api/00-openshift_gateway_api_namespace.yaml")

  for arch in "${ARCHS[@]}"; do
    mkdir -p "${OSSM_STAGING}/${arch}"
    pushd "${OSSM_STAGING}/${arch}" || return 1
    title "extracting OSSM operator bundle for \"${arch}\" architecture"
    # shellcheck disable=SC2086  # Double quote to prevent globbing and word splitting.
    oc image extract \
      ${authentication} \
      --path /manifests/:. "${bundle_manifest}" \
      --filter-by-os "${arch}" \
      ||  {
        popd
        return 1
      }

      download_gateway_api_manifests "${PWD}" "${gateway_api_version}"
      extract_ossm_rbac_from_cluster_service_version "${PWD}" "${csv}" "${namespace}"
      extract_ossm_deploy_from_cluster_service_version "${PWD}" "${csv}" "${namespace}"

      for file in "${PWD}"/*; do
        if [[ ${file} == *.yaml || ${file} == *.yml ]]; then
            patch_namespace "Service" "${file}" "${namespace}"
            patch_namespace "Deployment" "${file}" "${namespace}"
            patch_image_values "Deployment" "${file}"
        fi
      done

      popd || return 1
    done
}

patch_namespace() {
  local kind=$1
  local file=$2
  local namespace=$3

  if [[ $(yq e ".kind == \"${kind}\"" "${file}") == "true" ]]; then
    # Check if the .metadata.namespace is not set or empty
    if [[ $(yq e ".metadata.namespace == \"${namespace}\"" "${file}") == "false" ]]; then
      echo "patching .metadata.namespace to \"${namespace}\" in ${file}"
      # Set the .metadata.namespace to the specified value
      ns=${namespace} yq e '.metadata.namespace = strenv(ns)'  -i "${file}"
    fi
  fi
}

patch_image_values() {
  local kind=$1
  local file=$2

  if [[ $(yq e ".kind == \"${kind}\"" "${file}") == "true" ]]; then
    yq -i 'with(.spec.template.spec.containers[]; .image=.name)' "${file}"
  fi
}


write_ossm_images_for_arch() {
    local arch="$1"
    title "Updating images for ${arch}"
    arch_dir="${STAGING_DIR}/ossm/${arch}"
    [ -d "${arch_dir}" ] || {
        echo "dir ${arch_dir} not found"
        return 1
    }

    local csv_manifest="${arch_dir}/servicemeshoperator3.clusterserviceversion.yaml"
    local kustomization_arch_file="${REPOROOT}/assets/optional/gateway-api/kustomization.${GOARCH_TO_UNAME_MAP[${arch}]}.yaml"
    local gateway_api_release_json="${REPOROOT}/assets/optional/gateway-api/release-gateway-api-${GOARCH_TO_UNAME_MAP[${arch}]}.json"

    local base_release
    base_release=$(yq ".spec.version" "${csv_manifest}")
    jq -n "{\"release\": {\"base\": \"${base_release}\"}, \"images\": {}}" > "${gateway_api_release_json}"

    cat <<EOF > "${kustomization_arch_file}"
images:
EOF
    images=$(yq '.spec.relatedImages[].name?' "${csv_manifest}")
    for image_name in ${images}; do
        local new_image
        new_image=$(yq ".spec.relatedImages[] | select(.name == \"${image_name}\") | .image" "${csv_manifest}")
        local new_image_name="${new_image%@*}"
        local new_image_digest="${new_image#*@}"
        cat <<EOF >> "${kustomization_arch_file}"
  - name: ${image_name}
    newName: ${new_image_name}
    digest: ${new_image_digest}
EOF
        yq -i -o json ".images += {\"${image_name}\": \"${new_image}\"}" "${gateway_api_release_json}"
    done
}

update_ossm_images() {
    title "Updating OSSM images"
    local workdir="${STAGING_DIR}/ossm"
    [ -d "${workdir}" ] || {
        >&2 echo 'ossm staging dir not found, aborting image update'
        return 1
    }
    for arch in "${ARCHS[@]}"; do
        write_ossm_images_for_arch "${arch}"
    done
}

update_ossm_manifests() {
    title "Copying OSSM manifests"
    local workdir="${STAGING_DIR}/ossm"
    [ -d "${workdir}" ] || {
        >&2 echo 'ossm staging dir not found, aborting asset update'
        return 1
    }
    "${REPOROOT}/scripts/auto-rebase/handle_assets.py" ./scripts/auto-rebase/ossm_assets.yaml
}

update_last_ossm_rebase() {
    local operator_bundle_manifest="$1"
    local gateway_api_version="$2"

    title "## Updating last_ossm_rebase.sh"

    local last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase_gateway_api.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase_gateway_api.sh to "${operator_bundle_manifest}" "${gateway_api_version}"
EOF
    chmod +x "${last_rebase_script}"

    (cd "${REPOROOT}" && \
         if test -n "$(git status -s scripts/auto-rebase/last_rebase_gateway_api.sh)"; then \
             title "## Committing changes to last_rebase_gateway_api.sh" && \
             git add scripts/auto-rebase/last_rebase_gateway_api.sh && \
             git commit -m "update last_rebase_gateway_api.sh"; \
         fi)
}

# In the ClusterServiceVersion there are encoded RBAC information for OLM deployments.
# Since microshift skips this installation and uses a custom one based on the bundle, we have to extract the RBAC
# manifests from the CSV by reading them out into separate files.
# shellcheck disable=SC2207
extract_ossm_rbac_from_cluster_service_version() {
  local dest="$1"
  local csv="$2"
  local namespace="$3"

  title "extracting OSSM clusterserviceversion.yaml into separate RBAC"

  local clusterPermissions=($(yq eval '.spec.install.spec.clusterPermissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${clusterPermissions[@]}"; do
    echo "extracting bundle .spec.install.spec.clusterPermissions by serviceAccountName ${service_account_name}"

    local clusterrole="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_clusterrole.yaml"
    echo "generating ${clusterrole}"
    extract_ossm_clusterrole_from_csv_by_service_account_name "${service_account_name}" "${csv}" "${clusterrole}"

    local clusterrolebinding="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml"
    echo "generating ${clusterrolebinding}"
    extract_ossm_clusterrolebinding_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${clusterrolebinding}"

    local service_account="${dest}/${service_account_name}_v1_serviceaccount.yaml"
    echo "generating ${service_account}"
    extract_ossm_service_account_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${service_account}"
  done

  local permissions=($(yq eval '.spec.install.spec.permissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${permissions[@]}"; do
    echo "extracting bundle .spec.install.spec.permissions by serviceAccountName ${service_account_name}"

    local role="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_role.yaml"
    echo "generating ${role}"
    extract_ossm_role_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${csv}" "${role}"

    local rolebinding="${dest}/${service_account_name}_rbac.authorization.k8s.io_v1_rolebinding.yaml"
    echo "generating ${rolebinding}"
    extract_ossm_rolebinding_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${rolebinding}"

    local service_account="${dest}/${service_account_name}_v1_serviceaccount.yaml"
    echo "generating ${service_account}"
    extract_ossm_service_account_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${service_account}"
  done
}

extract_ossm_deploy_from_cluster_service_version() {
  local dest="$1"
  local csv="$2"
  local namespace="$3"

  title "extracting OSSM clusterserviceversion.yaml into separate Deployments"

  mapfile -t deployments < <(yq eval '.spec.install.spec.deployments[].name' < "${csv}")

  for deployment in "${deployments[@]}"; do
    echo "extracting bundle .spec.install.spec.deployments by name ${deployment}"

    local deployment_file="${dest}/${deployment}_apps_v1_deployment.yaml"
    echo "generating ${deployment_file}"
    yq eval ".spec.install.spec.deployments[] | select(.name == \"${deployment}\") |
        .apiVersion = \"apps/v1\" |
        .kind = \"Deployment\" |
        .metadata.namespace = \"${namespace}\" |
        .metadata.name = .name |
        del(.name) |
        .metadata.labels = .label |
        del(.label) |
        del(.spec.template.spec.containers[].resources.limits)
        " "${csv}" > "${deployment_file}"
  done
}

extract_ossm_clusterrole_from_csv_by_service_account_name() {
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

extract_ossm_role_from_csv_by_service_account_name() {
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

extract_ossm_clusterrolebinding_from_csv_by_service_account_name() {
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

extract_ossm_rolebinding_from_csv_by_service_account_name() {
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
subjects:
- kind: ServiceAccount
  name: ${service_account_name}
  namespace: ${namespace}
EOL
)
  echo "${crb}" > "${target}"
}

extract_ossm_service_account_from_csv_by_service_account_name() {
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
    echo "$(basename "$0") to OSSM_RELEASE_BUNDLE GATEWAY_API_VERSION        Performs all the steps to rebase OSSM"
    echo "$(basename "$0") download OSSM_RELEASE_BUNSLE GATEWAY_API_VERSION  Downloads the content of a OSSM release image to disk in preparation for rebasing"
    echo "$(basename "$0") images                                            Updates OSSM images"
    echo "$(basename "$0") manifests                                         Updates OSSM manifests"
    exit 1
}

check_preconditions

command=${1:-help}
case "${command}" in
    to)
        rebase_ossm_to "$2" "$3"
        ;;
    download)
        download_ossm_operator_bundle_manifest "$2" "$3"
        ;;
    images)
        update_ossm_images
        ;;
    manifests)
        update_ossm_manifests
        ;;
    *) usage;;
esac
