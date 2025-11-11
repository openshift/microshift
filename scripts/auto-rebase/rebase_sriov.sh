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

NAMESPACE="sriov-network-operator"
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

# download_sriov_manifests() fetches the SR-IOV manifests.
# First, it downloads the SR-IOV Operator bundle CSV and extracts image ref to the SR-IOV Operator image.
# Then, it extracts the manifests from the SR-IOV Operator image to the staging dir.
# No processing is done in this function.
download_sriov_manifests() {
    local -r bundle_ref="${1}"

    rm -rf "${STAGING_SRIOV}" && mkdir -p "${STAGING_DOWNLOAD}"
    local -r authentication="$(get_auth)"

    title "Fetching SR-IOV manifests"
    # shellcheck disable=SC2086
    oc image extract \
        ${authentication} \
        --path "/manifests/:${STAGING_DOWNLOAD}" \
        "${bundle_ref}" || return 1
}

# extract_sriov_rbac_from_cluster_service_version() extract the RBAC from the
# CSV and saves the manifests in the STAGING_EXTRACTED directory. It is called
# by extract_sriov_manifests().
extract_sriov_rbac_from_cluster_service_version() {
  local dest="$1"
  local csv="$2"
  local namespace="$3"

  title "extracting sriov clusterserviceversion.yaml into separate RBAC"

  local clusterPermissions=($(yq eval '.spec.install.spec.clusterPermissions[].serviceAccountName' < "${csv}"))
  for service_account_name in "${clusterPermissions[@]}"; do
    echo "extracting bundle .spec.install.spec.clusterPermissions by serviceAccountName ${service_account_name}"

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
  done
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
  echo "${crb}" >> "${target}"
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
  echo "${crb}" >> "${target}"
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
  echo "${serviceAccount}" >> "${target}"
}

# extract_operator_from_csv() extracts the operator manifest from cluster
# service version.
extract_operator_from_csv() {
  local csv="$1"
  local namespace="$2"
  local target="$3"

  yq eval "
    .spec.install.spec.deployments[]
    | select(.name == \"sriov-network-operator\")
    | .apiVersion = \"apps/v1\"
    | .kind = \"Deployment\"
    | .metadata.name = .name
    | .metadata.namespace = \"${namespace}\"
    | del(.name)
    | del(.label)
  " "${csv}" > "${target}"
}

# create_namespace_yaml() creates a namespace manifest.
create_namespace_yaml() {
  local namespace="$1"
  local target="$2"

  namespace=$(cat <<EOL
apiVersion: v1
kind: Namespace
metadata: 
  name: ${namespace}
  labels:
    name: ${namespace}
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged
EOL
)
  echo "${namespace}" > "${target}"
}

# create_default_sriov_operator_config() creates the SriovOperatorConfig CR that
# the operator expects. This is not included in the csv, so I'm having to do it
# this way.
create_default_sriov_operator_config() {
  local namespace="$1"
  local target="$2"

  sriovoperatorconfig=$(cat <<EOL
apiVersion: sriovnetwork.openshift.io/v1
kind: SriovOperatorConfig
metadata:
  name: default
  namespace: ${namespace}
spec:
  configDaemonNodeSelector: {}
  logLevel: 2
  disableDrain: false
  configurationMode: daemon
EOL
)
  echo "${sriovoperatorconfig}" > "${target}"
}
# extract_sriov_manifests() extracts the RBAC, operator and configmap from
# cluster service version and saves the manifests in the STAGING_EXTRACTED
# directory.
extract_sriov_manifests() {
  rm -rf "${STAGING_EXTRACTED}" && mkdir -p "${STAGING_EXTRACTED}"

  # extract service_account, role, rolebinding, clusterrole, clusterrolebinding
  extract_sriov_rbac_from_cluster_service_version "${STAGING_EXTRACTED}" "${STAGING_DOWNLOAD}/${CSV_FILENAME}" "${NAMESPACE}"

  # extract supported nics configmap
  local configmap="${STAGING_DOWNLOAD}/${CONFIGMAP_FILENAME}"
  echo "generating ${configmap}"
  cp "${configmap}" "${STAGING_EXTRACTED}/"

  # extract operator
  local operator="${STAGING_EXTRACTED}/${OPERATOR_FILENAME}"
  echo "generating ${operator}"
  extract_operator_from_csv "${STAGING_DOWNLOAD}/${CSV_FILENAME}" "${NAMESPACE}" "${operator}"

  # create namespace
  local namespace="${STAGING_EXTRACTED}/namespace.yaml"
  echo "generating ${namespace}"
  create_namespace_yaml "${NAMESPACE}" "${namespace}"

  # create sriovoperatorconfig
  local sriovoperatorconfig="${STAGING_EXTRACTED}/sriovoperatorconfig.yaml"
  echo "generating ${sriovoperatorconfig}"
  create_default_sriov_operator_config "${NAMESPACE}" "${sriovoperatorconfig}"
}

# process_sriov_manifests() copies the extracted manifests and CRDs to their
# corresponding directory in assets/ and performs additional processing to align
# with MicroShift needs.
process_sriov_manifests() {
  rm -rf "${FINAL_ROOT}" && mkdir -p "${FINAL_DEPLOY}" "${FINAL_CRD}"
  local configmap="${STAGING_EXTRACTED}/${CONFIGMAP_FILENAME}"
  local operator="${STAGING_EXTRACTED}/${OPERATOR_FILENAME}"

  # copy extracted manifests to final destination
  cp -a "${STAGING_EXTRACTED}/." "${FINAL_DEPLOY}/" 

  # copy CRDs to final destination
  cp -a "${STAGING_DOWNLOAD}/sriovnetwork.openshift.io_"* "${FINAL_CRD}/"

  # add the nic we use in testing to supported nics list
  yq eval "
  .data.Intel_ixgbe_82576 = \"8086 10c9 10ca\"
  | .metadata.namespace = \"${NAMESPACE}\"
  " "${configmap}" > "${FINAL_DEPLOY}/${CONFIGMAP_FILENAME}"

  # turn off webhook and metrics exporter, change cluster type to k8s
  yq eval "
  (
    .spec.template.spec.containers[0].env[] |
    select(.name == \"ADMISSION_CONTROLLERS_ENABLED\")
  ).value = \"false\"
  |
  (
    .spec.template.spec.containers[0].env[] |
    select(.name == \"METRICS_EXPORTER_PROMETHEUS_OPERATOR_ENABLED\")
  ).value = \"false\"
  |
  .spec.template.spec.containers[0].env += [
    {\"name\": \"CLUSTER_TYPE\", \"value\": \"kubernetes\"}
  ]
  " "${operator}" > "${FINAL_DEPLOY}/${OPERATOR_FILENAME}"
}

get_sriov_bundle_version() {
    yq '.spec.version' "${STAGING_DOWNLOAD}/${CSV_FILENAME}"
}

update_last_rebase_sriov_sh() {
    local -r operator_bundle="${1}"

    title "Updating last_rebase_sriov.sh"
    local -r last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase_sriov.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase_sriov.sh to "${operator_bundle}"
EOF
    chmod +x "${last_rebase_script}"
}

update_rebase_job_entrypoint_sh() {
    local -r operator_bundle="${1}"

    title "Updating rebase_job_entrypoint.sh"
    sed -i \
        "s,^sriov_release=.*\$,sriov_release=\"${operator_bundle}\",g" \
        "${REPOROOT}/scripts/auto-rebase/rebase_job_entrypoint.sh"
}

rebase_sriov_to() {
    local -r operator_bundle="${1}"

    title "Rebasing SR-IOV for MicroShift to ${operator_bundle}"

    download_sriov_manifests "${operator_bundle}"
    local -r version=$(get_sriov_bundle_version)

    extract_sriov_manifests

    update_last_rebase_sriov_sh "${operator_bundle}"
    update_rebase_job_entrypoint_sh "${operator_bundle}"

    process_sriov_manifests

    if [[ -n "$(git status -s ./assets ./scripts/auto-rebase/rebase_job_entrypoint.sh ./scripts/auto-rebase/last_rebase_sriov.sh)" ]]; then
        title "Detected changes to assets/ or last_rebase_sriov.sh"

        if ! "${NO_BRANCH}"; then
            branch="rebase-sriov-${version}"
            title "Creating branch ${branch}"
            git branch -D "${branch}" 2>/dev/null || true && git checkout -b "${branch}"
        fi

        title "Committing changes"
        git add ./assets ./scripts/auto-rebase/rebase_job_entrypoint.sh ./scripts/auto-rebase/last_rebase_sriov.sh
        git commit -m "Update SR-IOV for MicroShift"
    else
        title "No changes to assets/ or last_rebase_sriov.sh"
    fi

    if ! "${KEEP_STAGING}"; then
        title "Removing staging directory"
        rm -rf "${STAGING_SRIOV}"
    fi
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to SRIOV_BUNDLE                    Performs all the steps to rebase SR-IOV for MicroShift"
    echo "$(basename "$0") download SRIOV_BUNDLE              Downloads the contents of the SR-IOV Operator (Bundle) to disk in preparation for rebasing"
    echo "$(basename "$0") process                            Process already downloaded SR-IOV Operator (Bundle) artifacts to update SR-IOV for MicroShift"
    exit 1
}

check_preconditions

command=${1:-help}
case "${command}" in
    to)
        rebase_sriov_to "$2"
        ;;
    download)
        download_sriov_manifests "$2"
        ;;
    process)
        extract_sriov_manifests
        process_sriov_manifests
        ;;
    *) usage;;
esac