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
STAGING_RHOAI="${STAGING_ROOT}/rhoai"
STAGING_BUNDLE="${STAGING_RHOAI}/bundle"
STAGING_OPERATOR="${STAGING_RHOAI}/operator"

CSV_FILENAME="rhods-operator.clusterserviceversion.yaml"
PULL_SECRET_FILE="${HOME}/.pull-secret.json"

RELEASE_JSON="${REPOROOT}/assets/optional/ai-model-serving/release-ai-model-serving-x86_64.json"

KEEP_STAGING="${KEEP_STAGING:-false}"

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

get_auth() {
    if [ -f "${PULL_SECRET_FILE}" ]; then
        echo "--registry-config ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
        echo ""
    fi
}

# download_rhoai_manifests() fetches the RHOAI's kserve and runtime manifests.
# First, it downloads the RHOAI Operator bundle CSV and extracts image ref to the RHOAI Operator image.
# Then, it extracts the manifests from the RHOAI Operator image to the staging dir.
# No processing is done in this functions.
download_rhoai_manifests() {
    local -r bundle_ref="${1}"

    # Jan/Feb 2025: Only x86_64 is supported (https://access.redhat.com/articles/rhoai-supported-configs)
    # therefore there's no loop over architectures.

    rm -rf "${STAGING_RHOAI}" && mkdir -p "${STAGING_BUNDLE}" "${STAGING_OPERATOR}"
    local -r authentication="$(get_auth)"

    title "Fetching RHOAI CSV"
    # shellcheck disable=SC2086
    oc image extract \
        ${authentication} \
        --path "/manifests/${CSV_FILENAME}:${STAGING_BUNDLE}" \
        "${bundle_ref}" \
        --filter-by-os amd64 || return 1

    local -r operator_ref=$(yq '.spec.relatedImages[] | select(.name == "odh-rhel8-operator-*") | .image' "${STAGING_BUNDLE}/${CSV_FILENAME}")
    title "Fetching RHOAI manifests"
    # shellcheck disable=SC2086
    oc image extract \
        ${authentication} \
        --path "/opt/manifests/:${STAGING_OPERATOR}" \
        "${operator_ref}" \
        --filter-by-os amd64 || return 1
}

process_rhoai_manifests() {
    title "Copying manifests from staging dir to assets/"
    "${REPOROOT}/scripts/auto-rebase/handle_assets.py" ./scripts/auto-rebase/assets_ai_model_serving.yaml

    title "Initializing release.json file"
    local -r version=$(get_rhoai_bundle_version)
    echo "{ \"release\": {\"base\": \"${version}\"}, \"images\": {}}" | yq -o json > "${RELEASE_JSON}"

    update_runtimes
    update_kserve
}

update_kserve() {
    local -r kserve_images=$(cat "${STAGING_OPERATOR}/kserve/overlays/odh/params.env")
    for image in ${kserve_images}; do
        local image_name="${image%=*}"
        local image_ref="${image#*=}"
        yq -i ".images.${image_name} = \"${image_ref}\"" "${RELEASE_JSON}"
    done
}

update_runtimes() {
    title "Dropping template containers from ServingRuntimes and changing them to ClusterServingRuntimes"
    shopt -s globstar nullglob
    for runtime in "${REPOROOT}/assets/optional/ai-model-serving/runtimes/"*.yaml; do
        if [[ $(basename "${runtime}") == "kustomization.yaml" ]]; then
            continue
        fi
        yq --inplace '.objects[0] | .kind = "ClusterServingRuntime"' "${runtime}"
        containers_amount=$(yq '.spec.containers | length' "${runtime}")
        for ((i=0; i<containers_amount; i++)); do
            # shellcheck disable=SC2016
            idx="${i}" yq --inplace --string-interpolation=false \
                '.spec.containers[env(idx)].image |= sub("\$\((.*)\)", "${1}")' \
                "${runtime}"
        done
    done

    title "Creating ClusterServingRuntimes images kustomization"

    local -r kustomization_images="${REPOROOT}/assets/optional/ai-model-serving/runtimes/kustomization.x86_64.yaml"
    cat <<EOF > "${kustomization_images}"

images:
EOF

    local -r images=$(cat "${STAGING_OPERATOR}"/odh-model-controller/base/*.env | grep "\-image")
    for image in ${images}; do
        local image_name="${image%=*}"
        local image_ref="${image#*=}"
        local image_ref_repo="${image_ref%@*}"
        local image_ref_digest="${image_ref#*@}"

        cat <<EOF >> "${kustomization_images}"
  - name: ${image_name}
    newName: ${image_ref_repo}
    digest: ${image_ref_digest}
EOF

        yq -i ".images.${image_name} = \"${image_ref}\"" "${RELEASE_JSON}"
    done
}

get_rhoai_bundle_version() {
    yq '.spec.version' "${STAGING_BUNDLE}/${CSV_FILENAME}"
}

update_last_rebase_ai_model_serving_sh() {
    local -r operator_bundle="${1}"

    title "Updating last_rebase_ai_model_serving.sh"
    local -r last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase_ai_model_serving.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase_ai_model_serving.sh to "${operator_bundle}"
EOF
    chmod +x "${last_rebase_script}"
}

rebase_ai_model_serving_to() {
    local -r operator_bundle="${1}"

    title "Rebasing AI Model Serving for MicroShift to ${operator_bundle}"

    download_rhoai_manifests "${operator_bundle}"
    local -r version=$(get_rhoai_bundle_version)

    update_last_rebase_ai_model_serving_sh "${operator_bundle}"

    process_rhoai_manifests

    if [[ -n "$(git status -s assets ./scripts/auto-rebase/last_rebase_ai_model_serving.sh)" ]]; then
        branch="rebase-ai-model_serving-${version}"
        title "Detected changes to assets/ or last_rebase_ai_model_serving.sh - creating branch ${branch}"
        git branch -D "${branch}" 2>/dev/null || true && git checkout -b "${branch}"

        title "Committing changes"
        git add assets ./scripts/auto-rebase/last_rebase_ai_model_serving.sh
        git commit -m "Update AI Model Serving for MicroShift"
    else
        title "No changes to assets/ or last_rebase_ai_model_serving.sh"
    fi

    if ! "${KEEP_STAGING}"; then
        title "Removing staging directory"
        rm -rf "${STAGING_RHOAI}"
    fi
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to RHOAI_OPERATOR_BUNDLE           Performs all the steps to rebase RHOAI Model Serving for MicroShift"
    echo "$(basename "$0") download RHOAI_OPERATOR_BUNDLE     Downloads the contents of the RHOAI Operator (Bundle) to disk in preparation for rebasing"
    echo "$(basename "$0") process                            Process already downloaded RHOAI Operator (Bundle) artifacts to update RHOAI Model Serving for MicroShift"
    exit 1
}

check_preconditions

command=${1:-help}
case "${command}" in
    to)
        rebase_ai_model_serving_to "$2"
        ;;
    download)
        download_rhoai_manifests "$2"
        ;;
    process)
        process_rhoai_manifests
        ;;
    *) usage;;
esac
