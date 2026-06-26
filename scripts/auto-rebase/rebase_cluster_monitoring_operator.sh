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
REBASE_USE_SSH="${REBASE_USE_SSH:-false}"

declare -a ARCHS=("amd64" "arm64")
declare -A GOARCH_TO_UNAME_MAP=( ["amd64"]="x86_64" ["arm64"]="aarch64" )

# Maps kustomization image name -> OCP release tag name
declare -A IMAGE_MAP=(
    ["quay.io/openshift/kube-metrics-server"]="kube-metrics-server"
    ["quay.io/openshift/kube-state-metrics"]="kube-state-metrics"
    ["quay.io/openshift/node-exporter"]="prometheus-node-exporter"
    ["quay.io/openshift/kube-rbac-proxy"]="kube-rbac-proxy"
)

# Maps component dir -> release JSON key
declare -A COMPONENT_JSON_KEY=(
    ["metrics-server"]="metrics_server"
    ["kube-state-metrics"]="kube_state_metrics"
    ["node-exporter"]="node_exporter"
)

# Maps release JSON key -> OCP release tag name
declare -A EXPORTER_TAG_MAP=(
    ["metrics_server"]="kube-metrics-server"
    ["kube_state_metrics"]="kube-state-metrics"
    ["node_exporter"]="prometheus-node-exporter"
)

title() {
    echo -e "\E[34m$1\E[00m";
}

retry_cmd() {
    local -r max_attempts=5
    local timeout=1
    local attempt=1
    local exit_code=0

    while (( attempt <= max_attempts )); do
        if "$@"; then
            return 0
        else
            exit_code=$?
        fi
        echo "Attempt ${attempt} of ${max_attempts} failed (exit code ${exit_code}). Retrying in ${timeout}s..."
        sleep "${timeout}"
        attempt=$(( attempt + 1 ))
        timeout=$(( timeout * 2 ))
    done

    echo "Command failed after ${max_attempts} attempts: $@"
    return "${exit_code}"
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

clone_repo() {
    local repo="$1"
    local commit="$2"
    local destdir="$3"

    local repodir="${destdir}/${repo##*/}"

    if [[ -d "${repodir}" ]]; then
        return
    fi

    if "${REBASE_USE_SSH}"; then
        repo="git@github.com:${repo#https://github.com/}"
    fi

    git init "${repodir}"
    pushd "${repodir}" >/dev/null
    git remote add origin "${repo}"
    retry_cmd git fetch origin --quiet --filter=tree:0 --tags "${commit}"
    git checkout "${commit}"
    popd >/dev/null
}

download_cluster_monitoring_operator() {
    local release_image_amd64="$1"
    local release_image_arm64="$2"

    rm -rf "${STAGING_DIR}"
    mkdir -p "${STAGING_DIR}"
    pushd "${STAGING_DIR}" >/dev/null

    local authentication=""
    if [[ -f "${PULL_SECRET_FILE}" ]]; then
        authentication="-a ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
    fi

    title "# Fetching release info for ${release_image_amd64} (amd64)"
    oc adm release info ${authentication} "${release_image_amd64}" -o json > release_amd64.json
    title "# Fetching release info for ${release_image_arm64} (arm64)"
    oc adm release info ${authentication} "${release_image_arm64}" -o json > release_arm64.json

    title "# Extracting cluster-monitoring-operator source commit"
    cat release_amd64.json \
        | jq -r '.references.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location") \(.annotations."io.openshift.build.commit.id")"' > source-commits

    local cmo_line
    cmo_line=$(grep '^cluster-monitoring-operator ' source-commits) || {
        >&2 echo "ERROR: cluster-monitoring-operator not found in release payload"
        return 1
    }

    local repo commit
    repo=$(echo "${cmo_line}" | cut -d ' ' -f 2)
    commit=$(echo "${cmo_line}" | cut -d ' ' -f 3)

    title "# Cloning cluster-monitoring-operator at ${commit}"
    clone_repo "${repo}" "${commit}" "."

    popd >/dev/null
}

update_metrics_server_manifests() {
    [[ -d "${REPOROOT}/assets/optional/metrics-server" ]] || return 0

    title "Rebasing metrics-server manifests"

    local ms_crb="${REPOROOT}/assets/optional/metrics-server/01-cluster-role-binding.yaml"
    yq -i '.subjects += [{"kind": "User", "name": "system:metrics-server"}]' "$ms_crb"

    local ms_deploy="${REPOROOT}/assets/optional/metrics-server/03-deployment.yaml"
    yq -i '.spec.replicas = 1' "$ms_deploy"
    yq -i '.spec.strategy = {"type": "Recreate"}' "$ms_deploy"
    yq -i 'del(.spec.template.spec.affinity)' "$ms_deploy"
    yq -i '.spec.template.spec.containers[0].image = "quay.io/openshift/kube-metrics-server"' "$ms_deploy"
    yq -i '.spec.template.spec.containers[0].securityContext.capabilities.drop = ["ALL"]' "$ms_deploy"
}

update_kube_state_metrics_manifests() {
    [[ -d "${REPOROOT}/assets/optional/kube-state-metrics" ]] || return 0

    title "Rebasing kube-state-metrics manifests"

    local ksm_deploy="${REPOROOT}/assets/optional/kube-state-metrics/03-deployment.yaml"

    yq -i '.spec.template.spec.containers[0].image = "quay.io/openshift/kube-state-metrics"' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[1].image = "quay.io/openshift/kube-rbac-proxy"' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[2].image = "quay.io/openshift/kube-rbac-proxy"' "$ksm_deploy"

    yq -i '.spec.template.spec.containers[0].securityContext = {"allowPrivilegeEscalation": false, "readOnlyRootFilesystem": true, "runAsNonRoot": true}' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[1].securityContext = {"allowPrivilegeEscalation": false, "readOnlyRootFilesystem": true, "runAsNonRoot": true}' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[2].securityContext = {"allowPrivilegeEscalation": false, "readOnlyRootFilesystem": true, "runAsNonRoot": true}' "$ksm_deploy"
    yq -i '.spec.template.spec.securityContext = {"runAsNonRoot": true}' "$ksm_deploy"

    yq -i '.spec.template.spec.containers[0].resources.limits = {"cpu": "100m", "memory": "200Mi"}' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[1].resources.limits = {"cpu": "20m", "memory": "40Mi"}' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[2].resources.limits = {"cpu": "20m", "memory": "40Mi"}' "$ksm_deploy"

    yq -i '(.spec.template.spec.containers[1].volumeMounts[] | select(.name == "kube-state-metrics-tls")).readOnly = true' "$ksm_deploy"
    yq -i '(.spec.template.spec.containers[2].volumeMounts[] | select(.name == "kube-state-metrics-tls")).readOnly = true' "$ksm_deploy"

    yq -i '(.spec.template.spec.containers[1].args[] | select(test("--client-ca-file="))) |= "--client-ca-file=/etc/tls/client-ca/ca.crt"' "$ksm_deploy"
    yq -i '(.spec.template.spec.containers[2].args[] | select(test("--client-ca-file="))) |= "--client-ca-file=/etc/tls/client-ca/ca.crt"' "$ksm_deploy"
    yq -i 'del(.spec.template.spec.volumes[] | select(.name == "metrics-client-ca"))' "$ksm_deploy"
    yq -i '.spec.template.spec.volumes += [{"hostPath": {"path": "/var/lib/microshift/certs/admin-kubeconfig-signer/ca.crt", "type": "File"}, "name": "admin-kubeconfig-signer-ca"}]' "$ksm_deploy"
    yq -i 'del(.spec.template.spec.containers[1].volumeMounts[] | select(.name == "metrics-client-ca"))' "$ksm_deploy"
    yq -i 'del(.spec.template.spec.containers[2].volumeMounts[] | select(.name == "metrics-client-ca"))' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[1].volumeMounts += [{"mountPath": "/etc/tls/client-ca/ca.crt", "name": "admin-kubeconfig-signer-ca", "readOnly": true}]' "$ksm_deploy"
    yq -i '.spec.template.spec.containers[2].volumeMounts += [{"mountPath": "/etc/tls/client-ca/ca.crt", "name": "admin-kubeconfig-signer-ca", "readOnly": true}]' "$ksm_deploy"

    local ksm_secret="${REPOROOT}/assets/optional/kube-state-metrics/02-kube-rbac-proxy-secret.yaml"
    sed -i '/"user":/,/"name":/d' "$ksm_secret"
}

update_node_exporter_manifests() {
    [[ -d "${REPOROOT}/assets/optional/node-exporter" ]] || return 0

    title "Rebasing node-exporter manifests"

    local ne_ds="${REPOROOT}/assets/optional/node-exporter/03-daemonset.yaml"

    yq -i '.spec.template.spec.containers[0].image = "quay.io/openshift/node-exporter"' "$ne_ds"
    yq -i '.spec.template.spec.containers[1].image = "quay.io/openshift/kube-rbac-proxy"' "$ne_ds"
    yq -i '.spec.template.spec.initContainers[0].image = "quay.io/openshift/node-exporter"' "$ne_ds"

    yq -i '(.spec.template.spec.containers[1].args[] | select(test("--secure-listen-address="))) |= "--secure-listen-address=0.0.0.0:9100"' "$ne_ds"

    yq -i '(.spec.template.spec.containers[1].args[] | select(test("--client-ca-file="))) |= "--client-ca-file=/etc/tls/client-ca/ca.crt"' "$ne_ds"
    yq -i 'del(.spec.template.spec.volumes[] | select(.name == "metrics-client-ca"))' "$ne_ds"
    yq -i '.spec.template.spec.volumes += [{"hostPath": {"path": "/var/lib/microshift/certs/admin-kubeconfig-signer/ca.crt", "type": "File"}, "name": "admin-kubeconfig-signer-ca"}]' "$ne_ds"
    yq -i 'del(.spec.template.spec.containers[1].volumeMounts[] | select(.name == "metrics-client-ca"))' "$ne_ds"
    yq -i '.spec.template.spec.containers[1].volumeMounts += [{"mountPath": "/etc/tls/client-ca/ca.crt", "name": "admin-kubeconfig-signer-ca", "readOnly": true}]' "$ne_ds"

    yq -i '(.spec.template.spec.containers[1].volumeMounts[] | select(.name == "node-exporter-tls")).readOnly = true' "$ne_ds"

    local ne_secret="${REPOROOT}/assets/optional/node-exporter/02-kube-rbac-proxy-secret.yaml"
    sed -i '/"user":/,/"name":/d' "$ne_secret"
}

update_cluster_monitoring_operator_images() {
    title "Rebasing metrics component images"

    for goarch in amd64 arm64; do
        local arch=${GOARCH_TO_UNAME_MAP["${goarch}"]:-noarch}
        local release_file="${STAGING_DIR}/release_${goarch}.json"

        local base_release
        base_release=$(jq -r ".metadata.version" "${release_file}")

        for component_dir in metrics-server kube-state-metrics node-exporter; do
            [[ -d "${REPOROOT}/assets/optional/${component_dir}" ]] || continue

            local json_key="${COMPONENT_JSON_KEY[$component_dir]}"
            local release_tag="${EXPORTER_TAG_MAP[$json_key]}"
            local new_image
            new_image=$(jq -r ".references.spec.tags[] | select(.name == \"${release_tag}\") | .from.name" "${release_file}")
            if [[ -z "${new_image}" || "${new_image}" == "null" ]]; then
                >&2 echo "ERROR: Release tag '${release_tag}' not found in payload for ${component_dir}"
                return 1
            fi
            local component_release_json="${REPOROOT}/assets/optional/${component_dir}/release-${component_dir}-${arch}.json"
            jq -n --arg base "$base_release" --arg img "${new_image}" \
                "{\"release\": {\"base\": \$base}, \"images\": {\"${json_key}\": \$img}}" > "${component_release_json}"

            local kustomization_arch_file="${REPOROOT}/assets/optional/${component_dir}/kustomization.${arch}.yaml"

            cat <<EOF > "${kustomization_arch_file}"
images:
EOF

            local image_names
            image_names=$(grep -h 'image:' "${REPOROOT}/assets/optional/${component_dir}/"*.yaml 2>/dev/null \
                | sed 's/.*image: *//; s/"//g; s/:.*//; s/@.*//' | sort -u | grep -v '^$')

            for orig_image in ${image_names}; do
                local release_tag="${IMAGE_MAP[$orig_image]:-}"
                if [[ -z "${release_tag}" ]]; then
                    >&2 echo "ERROR: Unknown metrics image '${orig_image}' in ${component_dir}"
                    return 1
                fi

                local new_image
                new_image=$(jq -r ".references.spec.tags[] | select(.name == \"${release_tag}\") | .from.name" "${release_file}")
                if [[ -z "${new_image}" || "${new_image}" == "null" ]]; then
                    >&2 echo "ERROR: Image for release tag '${release_tag}' not found in payload for ${component_dir}"
                    return 1
                fi
                local new_image_name="${new_image%@*}"
                local new_image_digest="${new_image#*@}"

                cat <<EOF >> "${kustomization_arch_file}"
  - name: ${orig_image}
    newName: ${new_image_name}
    digest: ${new_image_digest}
EOF
            done
        done
    done
}

copy_manifests() {
    title "Copying manifests"
    "$REPOROOT/scripts/auto-rebase/handle_assets.py" "./scripts/auto-rebase/assets_cluster_monitoring_operator.yaml"
}

update_last_rebase() {
    local release_image_amd64="$1"
    local release_image_arm64="$2"

    title "## Updating last_rebase_cluster_monitoring_operator.sh"

    local last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase_cluster_monitoring_operator.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase_cluster_monitoring_operator.sh to "${release_image_amd64}" "${release_image_arm64}"
EOF
    chmod +x "${last_rebase_script}"

    (cd "${REPOROOT}" && \
         if test -n "$(git status -s scripts/auto-rebase/last_rebase_cluster_monitoring_operator.sh)"; then \
             title "## Committing changes to last_rebase_cluster_monitoring_operator.sh" && \
             git add scripts/auto-rebase/last_rebase_cluster_monitoring_operator.sh && \
             git commit -m "update last_rebase_cluster_monitoring_operator.sh"; \
         fi)
}

rebase_cluster_monitoring_operator_to() {
    local release_image_amd64="$1"
    local release_image_arm64="$2"
    download_cluster_monitoring_operator "${release_image_amd64}" "${release_image_arm64}"
    copy_manifests
    update_metrics_server_manifests
    update_kube_state_metrics_manifests
    update_node_exporter_manifests
    update_cluster_monitoring_operator_images
    update_last_rebase "${release_image_amd64}" "${release_image_arm64}"
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to AMD64_RELEASE_IMAGE ARM64_RELEASE_IMAGE        Performs all the steps to rebase metrics exporters."
    echo "$(basename "$0") download AMD64_RELEASE_IMAGE ARM64_RELEASE_IMAGE  Downloads the content of release images to disk in preparation for rebasing."
    echo "$(basename "$0") images                                            Rebases the component images to the downloaded release"
    echo "$(basename "$0") manifests                                         Rebases the component manifests to the downloaded release"
    exit 1
}

check_preconditions

command=${1:-help}
case "$command" in
    to)
        [[ $# -lt 3 ]] && usage
        rebase_cluster_monitoring_operator_to "$2" "$3"
        ;;
    download)
        [[ $# -lt 3 ]] && usage
        download_cluster_monitoring_operator "$2" "$3"
        ;;
    images)
        update_cluster_monitoring_operator_images
        ;;
    manifests)
        copy_manifests
        update_metrics_server_manifests
        update_kube_state_metrics_manifests
        update_node_exporter_manifests
        ;;
    *) usage;;
esac
