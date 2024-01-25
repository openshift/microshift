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
GO_MOD_DIRS=("$REPOROOT/" "$REPOROOT/etcd")

EMBEDDED_COMPONENTS="route-controller-manager cluster-policy-controller hyperkube etcd kube-storage-version-migrator"
EMBEDDED_COMPONENT_OPERATORS="cluster-kube-apiserver-operator cluster-kube-controller-manager-operator cluster-openshift-controller-manager-operator cluster-kube-scheduler-operator machine-config-operator operator-lifecycle-manager"
LOADED_COMPONENTS="cluster-dns-operator cluster-ingress-operator service-ca-operator cluster-network-operator cluster-csi-snapshot-controller-operator"
declare -a ARCHS=("amd64" "arm64")
declare -A GOARCH_TO_UNAME_MAP=( ["amd64"]="x86_64" ["arm64"]="aarch64" )

title() {
    echo -e "\E[34m$1\E[00m";
}

check_preconditions() {
    if ! hash yq; then
        title "Installing yq"

        local YQ_VER=4.26.1
        local YQ_HASH_amd64=9e35b817e7cdc358c1fcd8498f3872db169c3303b61645cc1faf972990f37582
        local YQ_HASH_arm64=8966f9698a9bc321eae6745ffc5129b5e1b509017d3f710ee0eccec4f5568766
        local YQ_HASH="YQ_HASH_$(go env GOARCH)"
        local YQ_URL=https://github.com/mikefarah/yq/releases/download/v${YQ_VER}/yq_linux_$(go env GOARCH)
        local YQ_EXE=$(mktemp /tmp/yq-exe.XXXXX)
        local YQ_SUM=$(mktemp /tmp/yq-sum.XXXXX)
        echo -n "${!YQ_HASH} -" > ${YQ_SUM}
        if ! (curl -Ls "${YQ_URL}" | tee ${YQ_EXE} | sha256sum -c ${YQ_SUM} &>/dev/null); then
            echo "ERROR: Expected file at ${YQ_URL} to have checksum ${!YQ_HASH} but instead got $(sha256sum <${YQ_EXE} | cut -d' ' -f1)"
            exit 1
        fi
        chmod +x ${YQ_EXE} && sudo cp ${YQ_EXE} /usr/bin/yq
        rm -f ${YQ_EXE} ${YQ_SUM}
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

# Clone a repo at a commit
clone_repo() {
    local repo="$1"
    local commit="$2"
    local destdir="$3"

    local repodir="${destdir}/${repo##*/}"

    if [[ -d "${repodir}" ]]
    then
        return
    fi

    git init "${repodir}"
    pushd "${repodir}" >/dev/null
    git remote add origin "${repo}"
    git fetch origin --quiet  --filter=tree:0 --tags "${commit}"
    git checkout "${commit}"
    popd >/dev/null
}

# Determine the image info for one architecture
download_image_state() {
    local release_image="$1"
    local release_image_arch="$2"

    local release_info_file="release_${release_image_arch}.json"
    local commits_file="image-repos-commits-${release_image_arch}"
    local new_commits_file="new-commits.txt"

    # Determine the repos and commits for the repos that build the images
    cat "${release_info_file}" \
        | jq -j '.references.spec.tags[] | if .annotations["io.openshift.build.source-location"] != "" then .name," ",.annotations["io.openshift.build.source-location"]," ",.annotations["io.openshift.build.commit.id"] else "" end,"\n"' \
             | sort -u \
             | grep -v '^$' \
                    > "${commits_file}"

    # Get list of MicroShift's container images. The names are not
    # arch-specific, so we just use the x86_64 list.
    local images=$(jq -r '.images | keys[]' "${REPOROOT}/assets/release/release-x86_64.json" | xargs)

    # Clone the repos. We clone a copy of each repo for each arch in
    # case they're on different branches or would otherwise not have
    # the history for both images if we only cloned one.
    #
    # TODO: This is probably more wasteful than just cloning the
    # entire git repo.
    mkdir -p "${release_image_arch}"
    local image=""
    for image in $images
    do
        if ! grep -q "^${image} " "${commits_file}"
        then
            # some of the images we use do not come from the release payload
            echo "${image} not from release payload, skipping"
            echo
            continue
        fi
        local line=$(grep "^${image} " "${commits_file}")
        local repo=$(echo "$line" | cut -f2 -d' ')
        local commit=$(echo "$line" | cut -f3 -d' ')
        clone_repo "${repo}" "${commit}" "${release_image_arch}"
        echo "${repo} image-${release_image_arch} ${commit}" >> "${new_commits_file}"
        echo
    done
}

# Downloads a release's tools and manifest content into a staging directory,
# then checks out the required components for the rebase at the release's commit.
download_release() {
    local release_image_amd64=$1
    local release_image_arm64=$2

    rm -rf "${STAGING_DIR}"
    mkdir -p "${STAGING_DIR}"
    pushd "${STAGING_DIR}" >/dev/null

    authentication=""
    if [ -f "${PULL_SECRET_FILE}" ]; then
        authentication="-a ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
    fi

    title "# Fetching release info for ${release_image_amd64} (amd64)"
    oc adm release info ${authentication} "${release_image_amd64}" -o json > release_amd64.json
    title "# Fetching release info for ${release_image_arm64} (arm64)"
    oc adm release info ${authentication} "${release_image_arm64}" -o json > release_arm64.json

    title "# Extracting ${release_image_amd64} manifest content"
    mkdir -p release-manifests
    pushd release-manifests >/dev/null
    content=$(oc adm release info ${authentication} --contents "${release_image_amd64}")
    echo "${content}" | awk '{ if ($0 ~ /^# [A-Za-z0-9._-]+.yaml$/ || $0 ~ /^# image-references$/ || $0 ~ /^# release-metadata$/) filename = $2; else print >filename;}'
    popd >/dev/null

    title "# Cloning ${release_image_amd64} component repos"
    cat release_amd64.json \
       | jq -r '.references.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location") \(.annotations."io.openshift.build.commit.id")"' > source-commits

    local new_commits_file="new-commits.txt"
    touch "${new_commits_file}"

    git config --global advice.detachedHead false
    git config --global init.defaultBranch main
    while IFS="" read -r line || [ -n "$line" ]
    do
        component=$(echo "${line}" | cut -d ' ' -f 1)
        repo=$(echo "${line}" | cut -d ' ' -f 2)
        commit=$(echo "${line}" | cut -d ' ' -f 3)
        if [[ "${EMBEDDED_COMPONENTS}" == *"${component}"* ]] || [[ "${LOADED_COMPONENTS}" == *"${component}"* ]] || [[ "${EMBEDDED_COMPONENT_OPERATORS}" == *"${component}"* ]]; then
            clone_repo "${repo}" "${commit}" "."
            echo "${repo} embedded-component ${commit}" >> "${new_commits_file}"
            echo
        fi
    done < source-commits

    title "# Cloning ${release_image_amd64} image repos"
    download_image_state "${release_image_amd64}" "amd64"
    download_image_state "${release_image_arm64}" "arm64"

    popd >/dev/null
}


# Greps a Golang pseudoversion from input.
grep_pseudoversion() {
    local line=$1

    echo "${line}" | grep -Po "v[0-9]+\.(0\.0-|\d+\.\d+-([^+]*\.)?0\.)\d{14}-[A-Za-z0-9]+(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?"
}

# Updates a require directive using an embedded component's commit.
require_using_component_commit() {
    local modulepath=$1
    local component=$2

    commit=$( cd "${STAGING_DIR}/${component}" && git rev-parse HEAD )
    echo "go mod edit -require ${modulepath}@${commit}"
    go mod edit -require "${modulepath}@${commit}"
    go mod tidy # needed to replace commit with pseudoversion before next invocation of go mod edit
}

# Updates a replace directive using an embedded component's commit.
# Caches component pseudoversions for faster processing.
declare -A pseudoversions
replace_using_component_commit() {
    local modulepath=$1
    local new_modulepath=$2
    local component=$3
    local reponame=$4

    if [[ ${pseudoversions[${component}]+foo} ]]; then
        echo "go mod edit -replace ${modulepath}=${new_modulepath}@${pseudoversions[${component}]}"
        go mod edit -replace "${modulepath}=${new_modulepath}@${pseudoversions[${component}]}"
    else
        commit=$( cd "${STAGING_DIR}/${reponame}" && git rev-parse HEAD )
        echo "go mod edit -replace ${modulepath}=${new_modulepath}@${commit}"
        go mod edit -replace "${modulepath}=${new_modulepath}@${commit}"
        go mod tidy # needed to replace commit with pseudoversion before next invocation of go mod edit
        pseudoversion=$(grep_pseudoversion "$(get_replace_directive "$(pwd)/go.mod" "${modulepath}")")
        pseudoversions["${component}"]="${pseudoversion}"
    fi
}

# Updates a script to record the last rebase that was run to make it
# easier to reproduce issues and to test changes to the rebase script
# against the same set of images.
update_last_rebase() {
    local release_image_amd64=$1
    local release_image_arm64=$2

    title "## Updating last_rebase.sh"

    local last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase.sh"

    rm -f "${last_rebase_script}"
    cat - >"${last_rebase_script}" <<EOF
#!/bin/bash -x
./scripts/auto-rebase/rebase.sh to "${release_image_amd64}" "${release_image_arm64}"
EOF
    chmod +x "${last_rebase_script}"

    (cd "${REPOROOT}" && \
         if test -n "$(git status -s scripts/auto-rebase/last_rebase.sh)"; then \
             title "## Committing changes to last_rebase.sh" && \
             git add scripts/auto-rebase/last_rebase.sh && \
             git commit -m "update last_rebase.sh"; \
         fi)
}

# Updates the ReplaceDirective for an old ${modulepath} with the new modulepath
# and version as per the staged checkout of ${component}.
update_modulepath_version_from_release() {
    local modulepath=$1
    local component=$2
    local reponame=$3

    local new_modulepath
    local path
    local repo

    path=""
    if [ "${component}" = "etcd" ]; then
        path="${modulepath#go.etcd.io/etcd}"
    fi
    repo=$( cd "${STAGING_DIR}/${reponame}" && git config --get remote.origin.url )
    new_modulepath="${repo#https://}${path}"
    replace_using_component_commit "${modulepath}" "${new_modulepath}" "${component}" "${reponame}"
}

# Updates the ReplaceDirective for an old ${modulepath} with the new modulepath
# in the staging directory of openshift/kubernetes at the released version.
update_modulepath_to_kubernetes_staging() {
    local modulepath=$1

    new_modulepath="github.com/openshift/kubernetes/staging/src/${modulepath}"
    replace_using_component_commit "${modulepath}" "${new_modulepath}" "kubernetes" "kubernetes"
}

# Returns the line (including trailing comment) in the #{gomod_file} containing the ReplaceDirective for ${module_path}
get_replace_directive() {
    local gomod_file=$1
    local module_path=$2

    replace=$(go mod edit -print "${gomod_file}" | grep "[[:space:]]${module_path}[[:space:]][[:alnum:][:space:].-]*=>")
    echo -e "${replace/replace /}"
}

# Updates a ReplaceDirective for an old ${modulepath} with the new modulepath
# and version as specified in the go.mod file of ${component}, taking care of
# necessary substitutions of local modulepaths.
update_modulepath_version_from_component() {
    local modulepath=$1
    local component=$2

    # Special-case etcd to use OpenShift's repo
    if [[ "${modulepath}" =~ ^go.etcd.io/etcd/ ]]; then
        update_modulepath_version_from_release "${modulepath}" "${component}" "${component}"
        return
    fi

    replace_directive=$(get_replace_directive "${STAGING_DIR}/${component}/go.mod" "${modulepath}")
    replace_directive=$(strip_comment "${replace_directive}")
    replacement=$(echo "${replace_directive}" | sed -E "s|.*=>[[:space:]]*(.*)[[:space:]]*|\1|")
    if [[ "${replacement}" =~ ^./staging ]]; then
        new_modulepath=$(echo "${replacement}" | sed 's|^./staging/|github.com/openshift/kubernetes/staging/|')
        replace_using_component_commit "${modulepath}" "${new_modulepath}" "${component}" "${component}"
    else
        echo "go mod edit -replace ${modulepath}=${replacement/ /@}"
        go mod edit -replace "${modulepath}=${replacement/ /@}"
    fi
}

# Trim trailing whitespace
rtrim() {
    local line=$1

    echo "${line}" | sed 's|[[:space:]]*$||'
}

# Trim leading whitespace
ltrim() {
    local line=$1

    echo "${line}" | sed 's|^[[:space:]]*||'
}

# Trim leading and trailing whitespace
trim() {
    local line=$1

    ltrim "$(rtrim "${line}")"
}

# Returns ${line} stripping the trailing comment
strip_comment() {
    local line=$1

    rtrim "${line%%//*}"
}

# Returns the comment in ${line} if one exists or an empty string if not
get_comment() {
    local line=$1

    comment=${line##*//}
    if [ "${comment}" != "${line}" ]; then
        trim "${comment}"
    else
        echo ""
    fi
}

# Replaces comment(s) at the end of the ReplaceDirective for $modulepath with $newcomment
update_comment() {
    local modulepath=$1
    local newcomment=$2

    oldline=$(get_replace_directive "$(pwd)/go.mod" "${modulepath}")
    newline="$(strip_comment "${oldline}") // ${newcomment}"
    sed -i "s|${oldline}|${newline}|" "$(pwd)/go.mod"
}

# Validate that ${component} is in the allowed list for the lookup, else exit
valid_component_or_exit() {
    local component=$1

    if [[ ! " ${EMBEDDED_COMPONENTS/hyperkube/kubernetes} " =~ " ${component} " ]]; then
        echo "error: component must be one of [${EMBEDDED_COMPONENTS/hyperkube/kubernetes}], have ${component}"
        exit 1
    fi
}

# Return all o/k staging repos (borrowed from k/k's hack/lib/util.sh)
list_staging_repos() {
  (
    cd "${STAGING_DIR}/kubernetes/staging/src/k8s.io" && \
    find . -mindepth 1 -maxdepth 1 -type d | cut -c 3- | sort
  )
}

# Updates current dir's go.mod file by updating each ReplaceDirective's
# new modulepath-version with that of one of the embedded components.
# The go.mod file needs to specify which component to take this data from
# and this is driven from keywords added as comments after each line of
# ReplaceDirectives:
#   // from ${component}     selects the replacement from the go.mod of ${component}
#   // staging kubernetes    selects the replacement from the staging dir of openshift/kubernetes
#   // release ${component}  uses the commit of ${component} as specified in the release image
#   // override [${reason}]  keep existing replacement
# Note directives without keyword comment are skipped with a warning.
update_go_mod() {
    title "# Updating $(basename "$(pwd)")/go.mod"

    # Require updated versions of RCM and CPC
    require_using_component_commit github.com/openshift/cluster-policy-controller cluster-policy-controller
    require_using_component_commit github.com/openshift/route-controller-manager route-controller-manager

    # For all repos in o/k staging, ensure a RequireDirective of v0.0.0
    # and a ReplaceDirective to an absolute modulepath to o/k staging.
    for repo in $(list_staging_repos); do
        go mod edit -require "k8s.io/${repo}@v0.0.0"
        update_modulepath_to_kubernetes_staging "k8s.io/${repo}"
        if [ -z "$(get_comment "$(get_replace_directive "$(pwd)/go.mod" "k8s.io/${repo}")")" ]; then
            update_comment "k8s.io/${repo}" "staging kubernetes"
        fi
    done

    # Update existing replace directives
    replaced_modulepaths=$(go mod edit -json | jq -r '.Replace // []' | jq -r '.[].Old.Path' | xargs)
    for modulepath in ${replaced_modulepaths}; do
        current_replace_directive=$(get_replace_directive "$(pwd)/go.mod" "${modulepath}")
        comment=$(get_comment "${current_replace_directive}")
        command=${comment%% *}
        arguments=${comment#${command} }
        case "${command}" in
        from)
            component=${arguments%% *}
            valid_component_or_exit "${component}"
            update_modulepath_version_from_component "${modulepath}" "${component}"
            ;;
        staging)
            update_modulepath_to_kubernetes_staging "${modulepath}"
            ;;
        release)
            component=${arguments%% *}
            reponame="${component}"
            if [[ "${arguments}" =~ " via " ]]; then
                argarray=( ${arguments} )
                reponame="${argarray[2]}"
            fi
            valid_component_or_exit "${component}"
            update_modulepath_version_from_release "${modulepath}" "${component}" "${reponame}"
            ;;
        override)
            echo "skipping modulepath ${modulepath}: override [${arguments}]"
            ;;
        *)
            echo "skipping modulepath ${modulepath}: no or unknown command [${comment}]"
            ;;
        esac
    done

    go mod tidy
}

# Updates go.mod file in dirs defined in GO_MOD_DIRS
update_go_mods() {
    # Update Go versions in the go.mod based on values in kubernetes' and etcd's go.mod
    kubernetes_go_version=$(go mod edit -json "${STAGING_DIR}/kubernetes/go.mod" | jq -r '.Go')
    go mod edit -go="${kubernetes_go_version}" "${REPOROOT}/go.mod"

    etcd_go_version=$(go mod edit -json "${STAGING_DIR}/etcd/go.mod" | jq -r '.Go')
    go mod edit -go="${etcd_go_version}" "${REPOROOT}/etcd/go.mod"

    for d in "${GO_MOD_DIRS[@]}"; do
        pushd "${d}" > /dev/null
        update_go_mod
        popd > /dev/null
    done
 }

# Regenerates OpenAPIs after patching the vendor directory
regenerate_openapi() {
    pushd "${STAGING_DIR}/kubernetes" >/dev/null

    title "Regenerating kube OpenAPI"
    make gen_openapi
    cp ./pkg/generated/openapi/zz_generated.openapi.go "${REPOROOT}/vendor/k8s.io/kubernetes/pkg/generated/openapi"

    popd >/dev/null
}


# Returns the list of release image names from a release_${arch}.go file
get_release_images() {
    file=$1

    awk "BEGIN {output=0} /^}/ {output=0} {if (output == 1) print substr(\$1, 2, length(\$1)-3)} /^var Image/ {output=1}" "${file}"
}

# Updates the image digests in pkg/release/release*.go
update_images() {
    if [ ! -f "${STAGING_DIR}/release_amd64.json" ] || [ ! -f "${STAGING_DIR}/release_arm64.json" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Rebasing release_*.json"
    for goarch in amd64 arm64; do
        arch=${GOARCH_TO_UNAME_MAP["${goarch}"]:-noarch}

        # Update the base release
        base_release=$(jq -r ".metadata.version" "${STAGING_DIR}/release_${goarch}.json")
        jq --arg base "${base_release}" '
            .release.base = $base
            ' "${REPOROOT}/assets/release/release-${arch}.json" > "${REPOROOT}/assets/release/release-${arch}.json.tmp"
        mv "${REPOROOT}/assets/release/release-${arch}.json.tmp" "${REPOROOT}/assets/release/release-${arch}.json"

        # Get list of MicroShift's container images
        images=$(jq -r '.images | keys[]' "${REPOROOT}/assets/release/release-${arch}.json" | xargs)

        # Extract the pullspecs for these images from OCP's release info
        jq --arg images "$images" '
            reduce .references.spec.tags[] as $img ({}; . + {($img.name): $img.from.name})
            | with_entries(select(.key == ($images | split(" ")[])))
            ' "release_${goarch}.json" > "update_${goarch}.json"

        # Update MicroShift's release info with these pullspecs
        jq --slurpfile updates "update_${goarch}.json" '
            .images += $updates[0]
            ' "${REPOROOT}/assets/release/release-${arch}.json" > "${REPOROOT}/assets/release/release-${arch}.json.tmp"
        mv "${REPOROOT}/assets/release/release-${arch}.json.tmp" "${REPOROOT}/assets/release/release-${arch}.json"

        # Update crio's pause image
        pause_image_digest=$(jq -r '
            .references.spec.tags[] | select(.name == "pod") | .from.name
            ' "release_${goarch}.json")
        sed -i "s|pause_image =.*|pause_image = \"${pause_image_digest}\"|g" \
            "${REPOROOT}/packaging/crio.conf.d/10-microshift_${goarch}.conf"
    done

    popd >/dev/null

    go fmt "${REPOROOT}"/pkg/release
}


copy_manifests() {
    if [ ! -f "${STAGING_DIR}/release_amd64.json" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    title "Copying manifests"
    "$REPOROOT/scripts/auto-rebase/handle_assets.py" "./scripts/auto-rebase/assets.yaml"
}


# Updates embedded component manifests by gathering these from various places
# in the staged repos and copying them into the asset directory.
update_openshift_manifests() {
    pushd "${STAGING_DIR}" >/dev/null

    title "Modifying OpenShift manifests"

    #-- Kubelet -------------------------------------------
    # Drop MCO's boilerplate and keep KubeletConfiguration only
    yq -i '.contents.inline' "${REPOROOT}/assets/core/kubelet.yaml"

    yq -i '.authentication.x509.clientCAFile = "{{ .clientCAFile }}" | .authentication.x509.clientCAFile style="double"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.clusterDNS = [ "{{ .clusterDNSIP }}" ] | .clusterDNS[] style="double"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.tlsCertFile = "{{ .tlsCertFile }}" | .tlsCertFile style="double"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.tlsPrivateKeyFile = "{{ .tlsPrivateKeyFile }}" | .tlsPrivateKeyFile style="double"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.volumePluginDir = "{{ .volumePluginDir }}" | .volumePluginDir style="double"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.failSwapOn = false' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.enforceNodeAllocatable = []' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.containerRuntimeEndpoint = "unix:///var/run/crio/crio.sock"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.cgroupsPerQOS = true' "${REPOROOT}/assets/core/kubelet.yaml"

    yq -i 'del(.podPidsLimit)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.protectKernelDefaults)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.staticPodPath)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.cgroupRoot)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.enableSystemLogQuery)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.podPidsLimit)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.protectKernelDefaults)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.systemCgroups)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.nodeStatusUpdateFrequency)' "${REPOROOT}/assets/core/kubelet.yaml"

    yq -i '.rotateCertificates = false | .rotateCertificates line_comment="TODO"' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.serverTLSBootstrap = false | .serverTLSBootstrap line_comment="TODO"' "${REPOROOT}/assets/core/kubelet.yaml"

    yq -i 'del(.tlsMinVersion)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.tlsCipherSuites)' "${REPOROOT}/assets/core/kubelet.yaml"

    # Clear and re-create featureGates
    yq -i 'del(.featureGates.AlibabaPlatform)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.AzureWorkloadIdentity)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.BuildCSIVolumes)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.CloudDualStackNodeIPs)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.ExternalCloudProvider)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.ExternalCloudProviderAzure)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.ExternalCloudProviderGCP)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.ExternalCloudProviderExternal)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.OpenShiftPodSecurityAdmission)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.PrivateHostedZoneAWS)' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i 'del(.featureGates.RetroactiveDefaultStorageClass)' "${REPOROOT}/assets/core/kubelet.yaml"

    yq -i '.featureGates.APIPriorityAndFairness = true' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.featureGates.PodSecurity = true' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.featureGates.DownwardAPIHugePages = true' "${REPOROOT}/assets/core/kubelet.yaml"
    yq -i '.featureGates.RotateKubeletServerCertificate = false | .featureGates.RotateKubeletServerCertificate line_comment="TODO"' "${REPOROOT}/assets/core/kubelet.yaml"

    # Sort the document, except for kind and apiVersion
    yq -i 'sort_keys(..) | pick((["kind","apiVersion"] + keys) | unique)' "${REPOROOT}/assets/core/kubelet.yaml"

    # Add optional resolvConf
    cat << 'EOF' >> "${REPOROOT}/assets/core/kubelet.yaml"
{{- if .resolvConf }}
resolvConf: "{{ .resolvConf }}"
{{- end }}
EOF

    #-- OpenShift control plane ---------------------------
    yq -i 'with(.admission.pluginConfig.PodSecurity.configuration.defaults;
        .enforce = "restricted" | .audit = "restricted" | .warn = "restricted" |
        .enforce-version = "latest" | .audit-version = "latest" | .warn-version = "latest")' "${REPOROOT}"/assets/controllers/kube-apiserver/defaultconfig.yaml
    yq -i 'del(.extendedArguments.pv-recycler-pod-template-filepath-hostpath)' "${REPOROOT}"/assets/controllers/kube-controller-manager/defaultconfig.yaml
    yq -i 'del(.extendedArguments.pv-recycler-pod-template-filepath-nfs)' "${REPOROOT}"/assets/controllers/kube-controller-manager/defaultconfig.yaml
    yq -i 'del(.extendedArguments.flex-volume-plugin-dir)' "${REPOROOT}"/assets/controllers/kube-controller-manager/defaultconfig.yaml
    yq -i '.spec.names.shortNames = ["scc"]' "${REPOROOT}"/assets/crd/0000_03_security-openshift_01_scc.crd.yaml

    #-- openshift-dns -------------------------------------
    # Render operand manifest templates like the operator would
    #    Render the DNS DaemonSet
    yq -i '.metadata += {"name": "dns-default", "namespace": "openshift-dns"}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.metadata += {"labels": {"dns.operator.openshift.io/owning-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.selector = {"matchLabels": {"dns.operator.openshift.io/daemonset-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.metadata += {"labels": {"dns.operator.openshift.io/daemonset-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.containers[0].image = "{{ .ReleaseImage.coredns }}"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.containers[1].image = "{{ .ReleaseImage.kube_rbac_proxy }}"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.nodeSelector = {"kubernetes.io/os": "linux"}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.volumes[0].configMap.name = "dns-default"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.volumes[1] += {"secret": {"defaultMode": 420, "secretName": "dns-default-metrics-tls"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.tolerations = [{"key": "node-role.kubernetes.io/master", "operator": "Exists"}]' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml

    #    Render the node-resolver script into the DaemonSet template
    export NODE_RESOLVER_SCRIPT="$(sed 's|^.|          &|' "${REPOROOT}"/assets/components/openshift-dns/node-resolver/update-node-resolver.sh)"
    envsubst < "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml.tmpl > "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml

    #    Render the DNS service
    yq -i '.metadata += {"annotations": {"service.beta.openshift.io/serving-cert-secret-name": "dns-default-metrics-tls"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    yq -i '.metadata += {"name": "dns-default", "namespace": "openshift-dns"}' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    yq -i '.spec.clusterIP = "{{.ClusterIP}}"' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    yq -i '.spec.selector = {"dns.operator.openshift.io/daemonset-dns": "default"}' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    sed -i '/#.*automatically managed/d' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml

    # Fix missing imagePullPolicy
    yq -i '.spec.template.spec.containers[1].imagePullPolicy = "IfNotPresent"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    # Temporary workaround for MicroShift's missing config parameter when rendering this DaemonSet
    sed -i 's|OPENSHIFT_MARKER=|NAMESERVER=${DNS_DEFAULT_SERVICE_HOST}\n          OPENSHIFT_MARKER=|' "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml


    #-- openshift-router ----------------------------------
    # Render operand manifest templates like the operator would
    yq -i '.metadata += {"name": "router-default", "namespace": "openshift-ingress"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.metadata += {"labels": {"ingresscontroller.operator.openshift.io/owning-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.minReadySeconds = 30' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.selector = {"matchLabels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.metadata += {"labels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].image = "{{ .ReleaseImage.haproxy_router }}"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "STATS_PORT", "value": "1936"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "RELOAD_INTERVAL", "value": "5s"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_ALLOW_WILDCARD_ROUTES", "value": "false"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_CANONICAL_HOSTNAME", "value": "router-default.apps.{{ .BaseDomain }}"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_CIPHERS", "value": "ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_CIPHERSUITES", "value": "TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384:TLS_CHACHA20_POLY1305_SHA256"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_DISABLE_HTTP2", "value": "true"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_DISABLE_NAMESPACE_OWNERSHIP_CHECK", "value": "false"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_LOAD_BALANCE_ALGORITHM", "value": "random"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # TODO: Generate and volume mount the metrics-certs secret
    # yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_METRICS_TLS_CERT_FILE", "value": "/etc/pki/tls/metrics-certs/tls.crt"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_METRICS_TLS_KEY_FILE", "value": "/etc/pki/tls/metrics-certs/tls.key"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_METRICS_TYPE", "value": "haproxy"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_SERVICE_NAME", "value": "default"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_SET_FORWARDED_HEADERS", "value": "append"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_TCP_BALANCE_SCHEME", "value": "source"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_THREADS", "value": "4"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "SSL_MIN_VERSION", "value": "TLSv1.2"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # TODO: Generate and volume mount the router-stats-default secret
    # yq -i '.spec.template.spec.containers[0].env += {"name": "STATS_PASSWORD_FILE", "value": "/var/lib/haproxy/conf/metrics-auth/statsPassword"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # yq -i '.spec.template.spec.containers[0].env += {"name": "STATS_USERNAME_FILE", "value": "/var/lib/haproxy/conf/metrics-auth/statsUsername"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports = []' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports += {"name": "http", "containerPort": 80, "protocol": "TCP"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports += {"name": "https", "containerPort": 443, "protocol": "TCP"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports += {"name": "metrics", "containerPort": 1936, "protocol": "TCP"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].args = ["-v=4"]' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.restartPolicy = "Always"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.terminationGracePeriodSeconds = 3600' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.dnsPolicy = "ClusterFirst"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.nodeSelector = {"kubernetes.io/os": "linux", "node-role.kubernetes.io/worker": ""}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.serviceAccount = "router"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.securityContext = {}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.schedulerName = "default-scheduler"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.volumes[0].secret.secretName = "router-certs-default"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    sed -i '/#.*at runtime/d' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    sed -i '/#.*at run-time/d' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.metadata.labels += {"ingresscontroller.operator.openshift.io/owning-ingresscontroller": "default"}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    yq -i '.metadata += {"name": "router-internal-default", "namespace": "openshift-ingress"}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    yq -i '.spec.selector = {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml

    # MicroShift-specific changes
    #-- ingress ----------------------------------------
    yq -i 'del(.metadata.annotations)' "${REPOROOT}"/assets/components/openshift-router/configmap.yaml
    #    Set replica count to 1, as we're single-node.
    yq -i '.spec.replicas = 1' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    #    Set deployment strategy type to Recreate.
    yq -i '.spec.strategy.type = "Recreate"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    #    Add hostPorts for routes and metrics (needed as long as there is no load balancer)
    yq -i '.spec.template.spec.containers[0].ports[0].hostPort = 80' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports[1].hostPort = 443' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    #    Not use proxy protocol due to lack of load balancer support
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_USE_PROXY_PROTOCOL", "value": "false"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "GRACEFUL_SHUTDOWN_DELAY", "value": "1s"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_DOMAIN", "value": "apps.{{ .BaseDomain }}"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml

    #-- service-ca ----------------------------------------
    # Render operand manifest templates like the operator would
    # TODO: Remove the following annotations once CPC correctly creates them automatically
    yq -i '.spec.template.spec.containers[0].args = ["-v=2"]' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i '.spec.template.spec.volumes[0].secret.secretName = "{{.TLSSecret}}"' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i '.spec.template.spec.volumes[1].configMap.name = "{{.CAConfigMap}}"' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i '.spec.template.spec.containers[0].image = "{{ .ReleaseImage.service_ca_operator }}"' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i 'del(.metadata.labels)' "${REPOROOT}"/assets/components/service-ca/ns.yaml

    # Make MicroShift-specific changes
    yq -i '.spec.replicas = 1' "${REPOROOT}"/assets/components/service-ca/deployment.yaml

    #-- ovn-kubernetes -----------------------------------
    # NOTE: As long as MicroShift is still based on OpenShift releases that do not yet contain the MicroShift-specific
    #       manifests we're manually updating them as needed for now.
    # TODO: Enable in assets.yaml and handle modifications

    #-- csi-snapshot-controller ---------------------------
    local target="${REPOROOT}/assets/components/csi-snapshot-controller/csi_controller_deployment.yaml"
    yq -i '.metadata.namespace = "kube-system"' $target
    yq -i '.spec.template.spec.containers[0].image = "{{ .ReleaseImage.csi_snapshot_controller }}"' $target
    yq -i '.spec.template.spec.containers[0].args = [ "--v=2", "--leader-election=false"]' $target
    yq -i 'del(.spec.template.spec.priorityClassName) | del(.spec.template.spec.containers[0].securityContext.seccompProfile)' $target
    yq -i 'with(.spec.template.spec.containers[0].securityContext; .runAsUser = 65534)' $target

    local target="${REPOROOT}/assets/components/csi-snapshot-controller/webhook_deployment.yaml"
    yq -i '.metadata.namespace = "kube-system"' $target
    yq -i '.spec.template.spec.containers[0].image = "{{ .ReleaseImage.csi_snapshot_validation_webhook }}"' $target
    yq -i 'with(.spec.template.spec.containers[0].args;  .[] |= sub("\${LOG_LEVEL}", "2") )' $target
    yq -i 'del(.spec.template.spec.priorityClassName)' $target
    yq -i 'with(.spec.template.spec.containers[0].securityContext; .runAsUser = 65534)' $target

    yq -i '.metadata.namespace = "kube-system"' "${REPOROOT}/assets/components/csi-snapshot-controller/webhook_service.yaml"

    yq -i '.webhooks[0].clientConfig.service.namespace="kube-system"' "${REPOROOT}/assets/components/csi-snapshot-controller/webhook_config.yaml"

    yq -i '.metadata.namespace = "kube-system"' "${REPOROOT}/assets/components/csi-snapshot-controller/serviceaccount.yaml"

    local target="${REPOROOT}/assets/components/csi-snapshot-controller/05_operand_rbac.yaml"
    yq -i '(.. | select(has("namespace")).namespace) = "kube-system"' $target
    # snapshotter's rbac is defined as a multidoc, which MicroShift is too picky to work with. Split into separate files
    yq 'select(.kind == "ClusterRole")' $target > "$(dirname $target)/clusterrole.yaml"
    yq 'select(.kind == "ClusterRoleBinding")' $target > "$(dirname $target)/clusterrolebinding.yaml"

    update_olm_images

    popd >/dev/null
}

update_version_makefile() {
    local arch="$1"
    local uname_i="$2"

    local release_file
    case "$arch" in
        amd64|x86_64)   release_file="${REPOROOT}/assets/release/release-x86_64.json"   ;;
        arm64|aarch64)  release_file="${REPOROOT}/assets/release/release-aarch64.json"  ;;
    esac

    local -r version_makefile="${REPOROOT}/Makefile.version.${uname_i}.var"
    local -r ocp_version=$(jq -r '.release.base' "$release_file")

    cat <<EOF > "$version_makefile"
OCP_VERSION := ${ocp_version}
EOF
}

# Updates buildfiles like the Makefile
update_buildfiles() {
    KUBE_ROOT="${STAGING_DIR}/kubernetes"
    if [ ! -d "${KUBE_ROOT}" ]; then
        >&2 echo "No kubernetes repo found at ${KUBE_ROOT}, you need to download a release first."
        exit 1
    fi

    pushd "${KUBE_ROOT}" >/dev/null

    title "Rebasing Makefile"
    source hack/lib/version.sh
    kube::version::get_version_vars

    local -r kube_version=$(jq -j \
        '.references.spec.tags[] | select(.name == "hyperkube") | .annotations["io.openshift.build.versions"] | split("=") | .[1]' \
        "${STAGING_DIR}/release_amd64.json")
    local -r kube_major=$(echo "${kube_version}" | awk -F'.' '{print $1}')
    local -r kube_minor=$(echo "${kube_version}" | awk -F'.' '{print $2}')

    cat <<EOF > "${REPOROOT}/Makefile.kube_git.var"
KUBE_GIT_MAJOR=${kube_major}
KUBE_GIT_MINOR=${kube_minor}
KUBE_GIT_VERSION=v${kube_version}
KUBE_GIT_COMMIT=${KUBE_GIT_COMMIT-}
KUBE_GIT_TREE_STATE=${KUBE_GIT_TREE_STATE-}
EOF

    popd >/dev/null

    update_version_makefile amd64 x86_64
    update_version_makefile arm64 aarch64
}


# Builds a list of the changes for each repository touched in this rebase
update_changelog() {
    local new_commits_file="${STAGING_DIR}/new-commits.txt"
    local old_commits_file="${REPOROOT}/scripts/auto-rebase/commits.txt"
    local changelog="${REPOROOT}/scripts/auto-rebase/changelog.txt"

    local repo # the URL to the repository
    local new_commit # the SHA of the commit to which we're updating
    local purpose # the purpose of the repo

    rm -f "$changelog"
    touch "$changelog"

    while read repo purpose new_commit
    do
        # Look for repo URL anchored at start of the line with a space
        # after it because some repos may have names that are
        # substrings of other repos.
        local old_commit=$(grep "^${repo} ${purpose} " "${old_commits_file}" | cut -f3 -d' ' | head -n 1)

        if [[ -z "${old_commit}" ]]
        then
            echo "- ${repo##*/} is a new ${purpose} dependency" >> "${changelog}"
            echo >> "${changelog}"
            continue
        fi

        if [[ "${old_commit}" == "${new_commit}" ]]; then
            # emit a message, but not to the changelog, to avoid cluttering it
            echo "${repo##*/} ${purpose} no change"
            echo
            continue
        fi

        local repodir
        case "${purpose}" in
            embedded-component)
                repodir="${STAGING_DIR}/${repo##*/}"
                ;;
            image-*)
                local image_arch=$(echo $purpose | cut -f2 -d-)
                repodir="${STAGING_DIR}/${image_arch}/${repo##*/}"
                ;;
            *)
                echo "- Unknown commit purpose \"${purpose}\" for ${repo}" >> "${changelog}"
                continue
                ;;
        esac
        pushd "${repodir}" >/dev/null
        echo "- ${repo##*/} ${purpose} ${old_commit} to ${new_commit}" >> "${changelog}"
        (git log \
             --no-merges \
             --pretty="format:  - %h %cI %s" \
             --no-decorate \
             "${old_commit}..${new_commit}" \
             || echo "There was an error determining the changes") >> "${changelog}"
        echo >> "${changelog}"
        echo >> "${changelog}"
        popd >/dev/null
    done < "${new_commits_file}"

    cp "${new_commits_file}" "${old_commits_file}"

    (cd "${REPOROOT}" && \
         if test -n "$(git status -s scripts/auto-rebase/changelog.txt scripts/auto-rebase/commits.txt)"; then \
             title "## Committing changes to changelog" && \
             git add scripts/auto-rebase/commits.txt scripts/auto-rebase/changelog.txt && \
             git commit -m "update changelog"; \
         fi)

}

checkout_rebase_branch() {
    ver_stream="$(cat ${STAGING_DIR}/release_amd64.json | jq -r '.config.config.Labels["io.openshift.release"]')"
    amd64_date="$(cat ${STAGING_DIR}/release_amd64.json | jq -r .config.created | cut -f1 -dT)"
    arm64_date="$(cat ${STAGING_DIR}/release_arm64.json | jq -r .config.created | cut -f1 -dT)"

    local onto_branch=""
    # PULL_BASE_REF is expected in CI environments and should identify the active branch.
    if [[ "${PULL_BASE_REF:-}" == "" ]]; then
        >&2 echo 'Warning: env var PULL_BASE_REF not found or empty, falling back to local active branch.'
        onto_branch=$(git rev-parse --abbrev-ref HEAD)
        if [[ "${onto_branch}" == 'HEAD' ]]; then
            >&2 echo 'Warning: detected HEAD in detached state, falling back to current abbreviated commit'
            onto_branch="$(git rev-parse --short HEAD)"
        fi
    else
        onto_branch="${PULL_BASE_REF}"
    fi
    # If onto_branch is not null, pad with "-", else onto_branch is empty so do not pad
    rebase_branch="rebase${onto_branch:+"-$onto_branch"}-${ver_stream}_amd64-${amd64_date}_arm64-${arm64_date}"
    git branch -D "${rebase_branch}" || true
    git checkout -b "${rebase_branch}"
}

update_olm_images() {
    title "Rebasing operator-lifecycle-manager manifests"

    # Replace hardcoded image refs with variables. Variables will be provided via kustomize's patches (added to `env`).
    # Expr with --util-image finds line with --util-image and edits the line after.
    sed -i \
        -e 's,--configmapServerImage=.*$,--configmapServerImage=$(OPERATOR_REGISTRY_IMAGE),g' \
        -e 's,--opmImage=.*$,--opmImage=$(OPERATOR_REGISTRY_IMAGE),g' \
        -e '/--util-image/{n;s,-.*,- \$\(OLM_IMAGE\),;}' \
        "${REPOROOT}/assets/optional/operator-lifecycle-manager/0000_50_olm_08-catalog-operator.deployment.yaml"

    # Remove tag/digest from images so they don't change on rebase unless the image name itself changes.
    # This way, check_for_manifests_changes() will be able to focus on limited list of files.
    sed -i -r 's,(image: [a-z./-]*)[@:].*,\1,' "${REPOROOT}/assets/optional/operator-lifecycle-manager/0000_50_olm_07-olm-operator.deployment.yaml"
    sed -i -r 's,(image: [a-z./-]*)[@:].*,\1,' "${REPOROOT}/assets/optional/operator-lifecycle-manager/0000_50_olm_08-catalog-operator.deployment.yaml"

    # If namespaces sourced from openshift/operator-lifecycle-manager do not contain openshift-marketplace - add it.
    # This conditional will make the transition easier when the namespace is added in upstream manifests.
    if ! grep -qw "openshift-marketplace" "${REPOROOT}/assets/optional/operator-lifecycle-manager/0000_50_olm_00-namespace.yaml"; then
        cat <<EOF >> "${REPOROOT}/assets/optional/operator-lifecycle-manager/0000_50_olm_00-namespace.yaml"
---
apiVersion: v1
kind: Namespace
metadata:
  name: openshift-marketplace
  labels:
    pod-security.kubernetes.io/enforce: restricted
    pod-security.kubernetes.io/enforce-version: "v1.24"
    openshift.io/scc: ""
  annotations:
    openshift.io/node-selector: ""
    workload.openshift.io/allowed: "management"
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    capability.openshift.io/name: "OperatorLifecycleManager"
EOF
    fi

    for goarch in amd64 arm64; do
        arch=${GOARCH_TO_UNAME_MAP["${goarch}"]:-noarch}

        # Get the images for the OLM operator from ${STAGING_DIR}/release_${arch}.json:
        # - operator-lifecycle-manager
        # - operator-registry
        # - kube-rbac-proxy
        local release_file="${STAGING_DIR}/release_${goarch}.json"
        local olm_image_refs_file="${REPOROOT}/assets/optional/operator-lifecycle-manager/image-references"
        local kustomization_arch_file="${REPOROOT}/assets/optional/operator-lifecycle-manager/kustomization.${arch}.yaml"
        local olm_release_json="${REPOROOT}/assets/optional/operator-lifecycle-manager/release-olm-${arch}.json"

        # Create the OLM release-${arch}.json file, this is the file included on the RPM.
        local base_release
        base_release=$(jq -r ".metadata.version" "${release_file}")
        jq -n "{\"release\": {\"base\": \"$base_release\"}, \"images\": {}}" > "${olm_release_json}"

        # Create extra kustomization for each arch in separate file.
        # Right file (depending on arch) should be appended during rpmbuild to kustomization.yaml.
        cat <<EOF > "${kustomization_arch_file}"

images:
EOF

        # Read from the global release file, to find the images we need to use.
        local containers=$(yq -r '.spec.tags[].name' "${olm_image_refs_file}")
        for container in ${containers[@]}; do
            # Get image (registry.com/image) without the tag or digest that is present in OLM's yamls and image-references.
            local orig_image_name
            orig_image_name=$(yq -r ".spec.tags[] | select(.name == \"${container}\") | .from.name" "${olm_image_refs_file}" | awk -F '[@:]' '{ print $1; }')

            # Get name and digest to replace OLM's image
            local new_image
            new_image=$(jq -r ".references.spec.tags[] | select(.name == \"${container}\") | .from.name" "${release_file}")
            local new_image_name="${new_image%@*}"
            local new_image_digest="${new_image#*@}"

            cat <<EOF >> "${kustomization_arch_file}"
  - name: ${orig_image_name}
    newName: ${new_image_name}
    digest: ${new_image_digest}
EOF

            yq -i -o json ".images += {\"${container}\": \"${new_image}\"}" "${olm_release_json}"
        done  # for container

        # kustomize's `images` transformer does not replace image refs outside `image: ` fields.
        # Following builds a patch to add required images as envs to be used with sed expressions executed above.
        local olm_image
        olm_image=$(jq -r '.references.spec.tags[] | select(.name == "operator-lifecycle-manager") | .from.name' "${release_file}")
        local registry_image
        registry_image=$(jq -r '.references.spec.tags[] | select(.name == "operator-registry") | .from.name' "${release_file}")

        cat << EOF >> "${kustomization_arch_file}"

patches:
  - patch: |-
     - op: add
       path: /spec/template/spec/containers/0/env/-
       value:
         name: OPERATOR_REGISTRY_IMAGE 
         value: ${registry_image}
     - op: add
       path: /spec/template/spec/containers/0/env/-
       value:
         name: OLM_IMAGE 
         value: ${olm_image}
    target:
      kind: Deployment
      labelSelector: app=catalog-operator
EOF
    done  # for goarch
}

check_for_manifests_changes() {
    # Changes to ignore:
    # - `release-$ARCH.json` files
    # - OLM's image-references and kustomization.$ARCH.yaml files
    local ignores="(release-(aarch64|x86_64).json|image-references|kustomization.(aarch64|x86_64).yaml)"

    title "Checking assets for unexpected changes"
    git status -s assets

    if [[ "$(git status -s assets | grep -vE "${ignores}" | wc -l)" -eq 0 ]]; then
        return 0
    fi

    title "\nERROR: Unexpected changes in assets"
    git --no-pager diff --unified=0 $(git status -s assets | grep -vE "${ignores}" | cut -d' ' -f3)
    return 1
}

# Runs each OCP rebase step in sequence, commiting the step's output to git
rebase_to() {
    local release_image_amd64=$1
    local release_image_arm64=$2

    title "# Rebasing to ${release_image_amd64} and ${release_image_arm64}"
    download_release "${release_image_amd64}" "${release_image_arm64}"

    checkout_rebase_branch

    update_last_rebase "${release_image_amd64}" "${release_image_arm64}"

    update_changelog
    update_go_mods
    for dirpath in "${GO_MOD_DIRS[@]}"; do
        dirname=$(basename "${dirpath}")
        if [[ -n "$(git status -s "${dirpath}/go.mod" "${dirpath}/go.sum")" ]]; then
            title "## Committing changes to ${dirname}/go.mod"
            git add "${dirpath}/go.mod" "${dirpath}/go.sum"
            git commit -m "update ${dirname}/go.mod"

            title "## Updating ${dirname}/vendor directory"
            (cd "${dirpath}" && make vendor)
            if [[ -n "$(git status -s "${dirpath}/vendor")" ]]; then
                title "## Commiting changes to ${dirname}/vendor directory"
                git add "${dirpath}/vendor"
                git commit -m "update ${dirname}/vendor"
            fi
        else
            echo "No changes in ${dirname}/go.mod."
        fi
    done

    update_images
    if [[ -n "$(git status -s pkg/release packaging/crio.conf.d)" ]]; then
        title "## Committing changes to pkg/release"
        git add pkg/release
        git add packaging/crio.conf.d
        git commit -m "update component images"
    else
        echo "No changes in component images."
    fi

    copy_manifests
    update_openshift_manifests
    if [[ -n "$(git status -s assets)" ]]; then
        if [[ -n "${FAIL_ON_MANIFEST_CHANGE+x}" ]] && [[ "${FAIL_ON_MANIFEST_CHANGE}" == "1" ]]; then
            check_for_manifests_changes
        fi
        title "## Committing changes to assets and pkg/assets"
        git add assets pkg/assets
        git commit -m "update manifests"
    else
        echo "No changes to assets."
    fi

    update_buildfiles
    if [[ -n "$(git status -s Makefile* packaging/rpm/microshift.spec)" ]]; then
        title "## Committing changes to buildfiles"
        git add Makefile* packaging/rpm/microshift.spec
        git commit -m "update buildfiles"
    else
        echo "No changes to buildfiles."
    fi

    title "# Removing staging directory"
    rm -rf "${STAGING_DIR}"
}

to_just_images() {
    local release_image_amd64=$1
    local release_image_arm64=$2

    download_release "${release_image_amd64}" "${release_image_arm64}"
    checkout_rebase_branch
    update_images

    if [[ -n "$(git status -s assets/release packaging/crio.conf.d)" ]]; then
        local last_rebase_script="${REPOROOT}/scripts/auto-rebase/last_rebase.sh"
        rm -f "${last_rebase_script}"
        cat - >"${last_rebase_script}" <<EOF
    #!/bin/bash -x
    ./scripts/auto-rebase/rebase.sh images-to "${release_image_amd64}" "${release_image_arm64}"
EOF
        chmod +x "${last_rebase_script}"

        title "## Committing new image references"
        git add assets/release
        git add packaging/crio.conf.d
        git add scripts/auto-rebase/last_rebase.sh
        git commit -m "update component images"
    else
        echo "No changes in component images."
    fi
}

usage() {
    echo "Usage:"
    echo "$(basename "$0") to RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM         Performs all the steps to rebase to a release image. Specify both amd64 and arm64 OCP releases."
    echo "$(basename "$0") download RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM   Downloads the content of a release image to disk in preparation for rebasing. Specify both amd64 and arm64 OCP releases."
    echo "$(basename "$0") images-to RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM  Downloads release image and only updates image references. Specify both amd64 and arm64 OCP releases."
    echo "$(basename "$0") buildfiles                                       Updates build files (Makefile, Dockerfile, .spec)"
    echo "$(basename "$0") go.mod                                           Updates the go.mod file to the downloaded release"
    echo "$(basename "$0") generated-apis                                   Regenerates OpenAPIs"
    echo "$(basename "$0") images                                           Rebases the component images to the downloaded release"
    echo "$(basename "$0") manifests                                        Rebases the component manifests to the downloaded release"
    exit 1
}

check_preconditions

command=${1:-help}
case "$command" in
    to)
        # FIXME: Allow more than 3 arguments until pipelines stop passing LVMS value.
        [[ $# -lt 3 ]] && usage
        rebase_to "$2" "$3"
        ;;
    download)
        # FIXME: Allow more than 3 arguments until pipelines stop passing LVMS value.
        [[ $# -lt 3 ]] && usage
        download_release "$2" "$3"
        ;;
    changelog) update_changelog;;
    buildfiles) update_buildfiles;;
    go.mod) update_go_mods;;
    generated-apis) regenerate_openapi;;
    images)
        update_images
        ;;
    images-to)
        to_just_images "$2" "$3"
        ;;
    manifests)
        copy_manifests
        update_openshift_manifests
        ;;
    *) usage;;
esac
