#!/usr/bin/env bash

set -o errexit
set -o errtrace
set -o nounset
set -o pipefail

shopt -s expand_aliases
shopt -s extglob

export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../..")"
STAGING_ROOT="${REPOROOT}/_output/staging"
STAGING_SRIOV="${STAGING_ROOT}/sriov"
STAGING_DOWNLOAD="${STAGING_SRIOV}/download"
STAGING_EXTRACTED="${STAGING_SRIOV}/extracted"

FINAL_ROOT="${REPOROOT}/assets/optional/sriov"
FINAL_DEPLOY="${FINAL_ROOT}/deploy"
FINAL_CRD="${FINAL_ROOT}/crd"

CSV_FILENAME="sriov-network-operator.clusterserviceversion.yaml"
CONFIGMAP_FILENAME="supported-nic-ids_v1_configmap.yaml"
OPERATOR_FILENAME="operator.yaml"
PULL_SECRET_FILE="${HOME}/.pull-secret.json"

RELEASE_JSON="${REPOROOT}/assets/optional/sriov/release-sriov-x86_64.json" # ??? TODO fix later when figured out what this is for

KEEP_STAGING="${KEEP_STAGING:-false}"
NO_BRANCH=${NO_BRANCH:-false}

title() {
    echo -e "\E[34m$1\E[00m";
}

check_preconditions() {
    if [[ "${REPOROOT}" != "$(pwd)" ]]; then
        echo "Script must be executed from root of the MicroShift repository"
        exit 1
    fi

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

    if ! hash oc; then
        echo "ERROR: oc is not present on the system - please install"
        exit 1
    fi
}

get_auth() {
    if [ -f "${PULL_SECRET_FILE}" ]; then
        echo "--registry-config ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
        echo ""
    fi
}

# append_if_exists() checks if the target file is non-empty, and if so, appends
# "---" to it, so another manifest can be included in the same file
append_if_exists() {
  local target="$1"
  
  if [[ -s "${target}" ]]; then
    echo "---" >> "${target}"
  fi
}

# download_rhoai_manifests() fetches the RHOAI's kserve and runtime manifests.
# First, it downloads the RHOAI Operator bundle CSV and extracts image ref to the RHOAI Operator image.
# Then, it extracts the manifests from the RHOAI Operator image to the staging dir.
# No processing is done in this functions.
download_sriov_manifests() {
    local -r bundle_ref="${1}"

    rm -rf "${STAGING_SRIOV}" && mkdir -p "${STAGING_DOWNLOAD}" "${STAGING_EXTRACTED}"
    local -r authentication="$(get_auth)"

    title "Fetching SR-IOV manifests"
    # shellcheck disable=SC2086
    oc image extract \
        ${authentication} \
        --path "/manifests/:${STAGING_DOWNLOAD}" \
        "${bundle_ref}" || return 1
}

extract_sriov_rbac_from_cluster_service_version() {
  local dest="$1"
  local csv="$2"
  local namespace="$3"

  title "extracting sriov clusterserviceversion.yaml into separate RBAC"

  local clusterPermissions=($(yq eval '.spec.install.spec.clusterPermissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${clusterPermissions[@]}"; do
    echo "extracting bundle .spec.install.spec.clusterPermissions by serviceAccountName ${service_account_name}"

    # local clusterrole="${dest}/${service_account_name}_clusterrole.yaml"
    local clusterrole="${dest}/clusterrole.yaml"

    echo "generating ${clusterrole}"
    extract_sriov_clusterrole_from_csv_by_service_account_name "${service_account_name}" "${csv}" "${clusterrole}"

    local clusterrolebinding="${dest}/clusterrolebinding.yaml"
    echo "generating ${clusterrolebinding}"
    extract_sriov_clusterrolebinding_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${clusterrolebinding}"

    local service_account="${dest}/serviceaccount.yaml"
    echo "generating ${service_account}"
    extract_sriov_service_account_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${service_account}"
  done

  local permissions=($(yq eval '.spec.install.spec.permissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${permissions[@]}"; do
    echo "extracting bundle .spec.install.spec.permissions by serviceAccountName ${service_account_name}"

    local role="${dest}/role.yaml"
    echo "generating ${role}"
    extract_sriov_role_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${csv}" "${role}"

    local rolebinding="${dest}/rolebinding.yaml"
    echo "generating ${rolebinding}"
    extract_sriov_rolebinding_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${rolebinding}"

    local service_account="${dest}/serviceaccount.yaml"
    echo "generating ${service_account}"
    extract_sriov_service_account_from_csv_by_service_account_name "${service_account_name}" "${namespace}" "${service_account}"
    
  done

  local operator="${dest}/${OPERATOR_FILENAME}"
  echo "generating ${operator}"
  extract_operator_from_csv "${csv}" "${operator}"

}

extract_sriov_clusterrole_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local csv="$2"
  local target="$3"

  append_if_exists "${target}"

  yq eval "
    .spec.install.spec.clusterPermissions[] |
    select(.serviceAccountName == \"${service_account_name}\") |
    .apiVersion = \"rbac.authorization.k8s.io/v1\" |
    .kind = \"ClusterRole\" |
    .metadata.name = \"${service_account_name}\" |
    del(.serviceAccountName)
    " "${csv}" >> "${target}"
}

extract_sriov_role_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local csv="$3"
  local target="$4"

  append_if_exists "${target}"

  yq eval "
    .spec.install.spec.permissions[] |
    select(.serviceAccountName == \"${service_account_name}\") |
    .apiVersion = \"rbac.authorization.k8s.io/v1\" |
    .kind = \"Role\" |
    .metadata.name = \"${service_account_name}\" |
    .metadata.namespace = \"${namespace}\" |
    del(.serviceAccountName)
    " "${csv}" >> "${target}"
}
extract_sriov_clusterrolebinding_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local target="$3"

  append_if_exists "${target}"

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

extract_sriov_rolebinding_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local target="$3"
 
  append_if_exists "${target}"

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

extract_sriov_service_account_from_csv_by_service_account_name() {
  local service_account_name="$1"
  local namespace="$2"
  local target="$3"

  append_if_exists "${target}"

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

extract_configmap_from_bundle() {
  local configmap="${STAGING_DOWNLOAD}/${CONFIGMAP_FILENAME}"

  cp "${configmap}" "${STAGING_EXTRACTED}/"
}

extract_operator_from_csv() {
  local csv="$1"
  local target="$2"

  yq eval '.spec.install.spec.deployments[] | select(.name == "sriov-network-operator")' "${csv}" > "${target}"
}

process_manifests() {
  rm -rf "${FINAL_DEPLOY}" && mkdir -p "${FINAL_DEPLOY}"
  local namespace="$1"
  local operatorfile="${STAGING_EXTRACTED}/${OPERATOR_FILENAME}"
  local configmap="${STAGING_EXTRACTED}/${CONFIGMAP_FILENAME}"

  cp -a "${STAGING_EXTRACTED}/." "${FINAL_DEPLOY}/" 

  # add apiVersion, Kind and namespace to operator.yaml
  yq eval ".apiVersion = \"v1\" | .kind = \"Deployment\" | .namespace = \"${namespace}\"" ${operatorfile} > "${FINAL_DEPLOY}/operator.yaml"

  # add the nic we use in testing to supported nics list
  yq eval ".data.Intel_ixgbe_82576 = \"8086 10c9 10ca\"" "${configmap}" > "${FINAL_DEPLOY}/configmap.yaml"
}

process_crds() {
  rm -rf "${FINAL_CRD}" && mkdir -p "${FINAL_CRD}"

  cp -a "${STAGING_DOWNLOAD}/sriovnetwork.openshift.io_"* "${FINAL_CRD}/"
}


check_preconditions

download_sriov_manifests "registry.redhat.io/openshift4/ose-sriov-network-operator-bundle:latest"

extract_sriov_rbac_from_cluster_service_version "${STAGING_EXTRACTED}" "${STAGING_DOWNLOAD}/${CSV_FILENAME}" "sriov-network-operator"

extract_configmap_from_bundle

process_manifests "sriov-network-operator"

process_crds
