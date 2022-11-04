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
set -o errtrace
set -o nounset
set -o pipefail

shopt -s expand_aliases
shopt -s extglob

trap 'echo "Script exited with error."' ERR

# debugging options
#trap 'echo "# $BASH_COMMAND"' DEBUG
#set -x

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/../..")"
STAGING_DIR="$REPOROOT/_output/staging"
PULL_SECRET_FILE="${HOME}/.pull-secret.json"

EMBEDDED_COMPONENTS="route-controller-manager cluster-policy-controller hyperkube etcd"
EMBEDDED_COMPONENT_OPERATORS="cluster-kube-apiserver-operator cluster-kube-controller-manager-operator cluster-openshift-controller-manager-operator cluster-kube-scheduler-operator machine-config-operator"
LOADED_COMPONENTS="cluster-dns-operator cluster-ingress-operator service-ca-operator cluster-network-operator"


title() {
    echo -e "\E[34m$1\E[00m";
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
    commits=$(oc adm release info ${authentication} --commits -o json "${release_image_amd64}")
    echo "${commits}" | jq -r '.references.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location") \(.annotations."io.openshift.build.commit.id")"' > source-commits

    git config --global advice.detachedHead false
    git config --global init.defaultBranch main
    while IFS="" read -r line || [ -n "$line" ]
    do
        component=$(echo "${line}" | cut -d ' ' -f 1)
        repo=$(echo "${line}" | cut -d ' ' -f 2)
        commit=$(echo "${line}" | cut -d ' ' -f 3)
        if [[ "${EMBEDDED_COMPONENTS}" == *"${component}"* ]] || [[ "${LOADED_COMPONENTS}" == *"${component}"* ]] || [[ "${EMBEDDED_COMPONENT_OPERATORS}" == *"${component}"* ]]; then
            title "## Cloning ${repo} at commit ${commit}..."
            git init "${repo##*/}"
            pushd "${repo##*/}" >/dev/null
            git remote add origin "${repo}"
            git fetch origin --depth=1 "${commit}"
            git checkout "${commit}"
            popd >/dev/null
            echo
        fi
    done < source-commits

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

    if [[ ${pseudoversions[${component}]+foo} ]]; then
        echo "go mod edit -replace ${modulepath}=${new_modulepath}@${pseudoversions[${component}]}"
        go mod edit -replace "${modulepath}=${new_modulepath}@${pseudoversions[${component}]}"
    else
        commit=$( cd "${STAGING_DIR}/${component}" && git rev-parse HEAD )
        echo "go mod edit -replace ${modulepath}=${new_modulepath}@${commit}"
        go mod edit -replace "${modulepath}=${new_modulepath}@${commit}"
        go mod tidy # needed to replace commit with pseudoversion before next invocation of go mod edit
        pseudoversion=$(grep_pseudoversion "$(get_replace_directive "${REPOROOT}/go.mod" "${modulepath}")")
        pseudoversions["${component}"]="${pseudoversion}"
    fi
}

# Updates the ReplaceDirective for an old ${modulepath} with the new modulepath
# and version as per the staged checkout of ${component}.
update_modulepath_version_from_release() {
    local modulepath=$1
    local component=$2

    path=""
    if [ "${component}" = "etcd" ]; then
        path="${modulepath#go.etcd.io/etcd}"
    fi
    repo=$( cd "${STAGING_DIR}/${component}" && git config --get remote.origin.url )
    new_modulepath="${repo#https://}${path}"
    replace_using_component_commit "${modulepath}" "${new_modulepath}" "${component}"
}

# Updates the ReplaceDirective for an old ${modulepath} with the new modulepath
# in the staging directory of openshift/kubernetes at the released version.
update_modulepath_to_kubernetes_staging() {
    local modulepath=$1

    new_modulepath="github.com/openshift/kubernetes/staging/src/${modulepath}"
    replace_using_component_commit "${modulepath}" "${new_modulepath}" "kubernetes"
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
        update_modulepath_version_from_release "${modulepath}" "${component}"
        return
    fi

    replace_directive=$(get_replace_directive "${STAGING_DIR}/${component}/go.mod" "${modulepath}")
    replace_directive=$(strip_comment "${replace_directive}")
    replacement=$(echo "${replace_directive}" | sed -E "s|.*=>[[:space:]]*(.*)[[:space:]]*|\1|")
    if [[ "${replacement}" =~ ^./staging ]]; then
        new_modulepath=$(echo "${replacement}" | sed 's|^./staging/|github.com/openshift/kubernetes/staging/|')
        replace_using_component_commit "${modulepath}" "${new_modulepath}" "${component}"
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

    oldline=$(get_replace_directive "${REPOROOT}/go.mod" "${modulepath}")
    newline="$(strip_comment "${oldline}") // ${newcomment}"
    sed -i "s|${oldline}|${newline}|" "${REPOROOT}/go.mod"
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

# Updates MicroShift's go.mod file by updating each ReplaceDirective's
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
    pushd "${STAGING_DIR}" >/dev/null

    title "# Updating go.mod"

    # Require updated versions of RCM and CPC
    require_using_component_commit github.com/openshift/cluster-policy-controller cluster-policy-controller
    require_using_component_commit github.com/openshift/route-controller-manager route-controller-manager

    # For all repos in o/k staging, ensure a RequireDirective of v0.0.0
    # and a ReplaceDirective to an absolute modulepath to o/k staging.
    for repo in $(list_staging_repos); do
        go mod edit -require "k8s.io/${repo}@v0.0.0"
        update_modulepath_to_kubernetes_staging "k8s.io/${repo}"
        if [ -z "$(get_comment "$(get_replace_directive "${REPOROOT}/go.mod" "k8s.io/${repo}")")" ]; then
            update_comment "k8s.io/${repo}" "staging kubernetes"
        fi
    done

    # Update existing replace directives
    replaced_modulepaths=$(go mod edit -json | jq -r '.Replace // []' | jq -r '.[].Old.Path' | xargs)
    for modulepath in ${replaced_modulepaths}; do
        current_replace_directive=$(get_replace_directive "${REPOROOT}/go.mod" "${modulepath}")
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
            valid_component_or_exit "${component}"
            update_modulepath_version_from_release "${modulepath}" "${component}"
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

    popd >/dev/null
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

    title "Rebasing release_*.go"

    # Update the base release
    base_release=$(jq -r ".metadata.version" "${STAGING_DIR}/release_amd64.json")
    sed -i "/^var Base/c\var Base = \"${base_release}\"" "${REPOROOT}/pkg/release/release.go"

    # Update the image digests for all architectures
    images="$(get_release_images "${REPOROOT}/pkg/release/release.go" | xargs)"
    for arch in amd64 arm64; do
        # Compute the max length of image names incl. enclosing quotes
        w=$(awk "BEGIN {n=split(\"${images}\", images, \" \"); max=0; for (i=1;i<=n;i++) {if (length(images[i]) > max) {max=length(images[i])}}; print max+2; exit}")
        for i in ${images}; do
            digest=$(jq -r ".references.spec.tags[] | select(.name == \"${i//_/-}\") | .from.name" release_${arch}.json)
            if [[ -n "${digest}" ]]; then
                echo "Updating image ${i//_/-} (${arch}) to ${digest}."
                awk "!/\"${i}\"/ {print \$0} /\"${i}\"/ {printf(\"\\t\\t%-${w}s  %s\n\", \"\\\"${i}\\\":\", \"\\\"${digest}\\\",\")}" \
                    "${REPOROOT}/pkg/release/release_${arch}.go" > t
                mv t "${REPOROOT}/pkg/release/release_${arch}.go"
                if [[ "$i" == "pod" ]]; then
                    echo "Updating image pod (${arch}) in packaging/crio.conf.d/microshift_${arch}.conf"
                    sed -i "s|pause_image =.*|pause_image = \"${digest}\"|g" \
                        "${REPOROOT}/packaging/crio.conf.d/microshift_${arch}.conf"
                fi
            else
                echo "Skipping ${i//_/-} (${arch}): Not part of release image."
            fi
        done
    done

    popd >/dev/null

    go fmt "${REPOROOT}"/pkg/release
}


# Updates embedded component manifests by gathering these from various places
# in the staged repos and copying them into the asset directory.
update_manifests() {
    if [ ! -f "${STAGING_DIR}/release_amd64.json" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Rebasing manifests"

    #-- OpenShift control plane ---------------------------
    # 1) Adopt resource manifests
    #    Selectively copy in only those core manifests that MicroShift is already using
    cp "${STAGING_DIR}/cluster-openshift-controller-manager-operator/bindata/v3.11.0/openshift-controller-manager/ns.yaml" "${REPOROOT}"/assets/core/0000_50_cluster-openshift-controller-manager_00_namespace.yaml
    cp "${STAGING_DIR}/cluster-kube-apiserver-operator/bindata/assets/config/config-overrides.yaml" "${REPOROOT}"/assets/components/kube-apiserver
    cp "${STAGING_DIR}/cluster-kube-apiserver-operator/bindata/assets/config/defaultconfig.yaml" "${REPOROOT}"/assets/components/kube-apiserver
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/csr_approver_clusterrole.yaml" "${REPOROOT}"/assets/core
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/csr_approver_clusterrolebinding.yaml" "${REPOROOT}"/assets/core
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/namespace-openshift-infra.yaml" "${REPOROOT}"/assets/core
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/ns.yaml" "${REPOROOT}"/assets/core/namespace-openshift-kube-controller-manager.yaml
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/namespace-security-allocation-controller-clusterrole.yaml" "${REPOROOT}"/assets/core
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/namespace-security-allocation-controller-clusterrolebinding.yaml" "${REPOROOT}"/assets/core
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/podsecurity-admission-label-syncer-controller-clusterrole.yaml" "${REPOROOT}"/assets/core
    cp "${STAGING_DIR}/cluster-kube-controller-manager-operator/bindata/assets/kube-controller-manager/podsecurity-admission-label-syncer-controller-clusterrolebinding.yaml" "${REPOROOT}"/assets/core
    #    Selectively copy in only those CRD manifests that MicroShift is already using
    cp "${STAGING_DIR}/cluster-ingress-operator/vendor/github.com/openshift/api/route/v1/route.crd.yaml" "${REPOROOT}"/assets/crd
    cp "${STAGING_DIR}/release-manifests/0000_03_security-openshift_01_scc.crd.yaml" "${REPOROOT}"/assets/crd
    cp "${STAGING_DIR}/release-manifests/0000_03_securityinternal-openshift_02_rangeallocation.crd.yaml" "${REPOROOT}"/assets/crd
    # The following manifests are just MicroShift specific and are not present in any other OpenShift repo.
    # - assets/crd/authorizationv1-local-apiservice.yaml (local API service for authorization API group, needed if OpenShift API server is not present)
    # - assets/crd/securityv1-local-apiservice.yaml (local API service for security API group, needed if OpenShift API server is not present)

    #    Replace all SCC manifests and their CRs/CRBs
    rm -f "${REPOROOT}"/assets/scc/*.yaml
    cp "${STAGING_DIR}"/release-manifests/0000_20_kube-apiserver-operator_00_cr-*.yaml "${REPOROOT}"/assets/scc || true
    cp "${STAGING_DIR}"/release-manifests/0000_20_kube-apiserver-operator_00_crb-*.yaml "${REPOROOT}"/assets/scc || true
    cp "${STAGING_DIR}"/release-manifests/0000_20_kube-apiserver-operator_00_scc-*.yaml "${REPOROOT}"/assets/scc || true
    # 2) Render operand manifest templates like the operator would
    #    n/a
    # 3) Make MicroShift-specific changes
    #    Add the missing scc shortName
    yq -i '.spec.names.shortNames = ["scc"]' "${REPOROOT}"/assets/crd/0000_03_security-openshift_01_scc.crd.yaml
    # 4) Replace MicroShift templating vars (do this last, as yq trips over Go templates)
    #    n/a

    #-- openshift-dns -------------------------------------
    # 1) Adopt resource manifests
    #    Replace all openshift-dns operand manifests
    rm -f "${REPOROOT}"/assets/components/openshift-dns/dns/*
    cp "${STAGING_DIR}"/cluster-dns-operator/assets/dns/* "${REPOROOT}"/assets/components/openshift-dns/dns || true 
    rm -f "${REPOROOT}"/assets/components/openshift-dns/node-resolver/*
    cp "${STAGING_DIR}/"cluster-dns-operator/assets/node-resolver/* "${REPOROOT}"/assets/components/openshift-dns/node-resolver || true
    #    Restore the openshift-dns ConfigMap. It's content is the Corefile that the operator generates
    #    in https://github.com/openshift/cluster-dns-operator/blob/master/pkg/operator/controller/controller_dns_configmap.go
    git restore "${REPOROOT}"/assets/components/openshift-dns/dns/configmap.yaml
    #    Restore the template for the node-resolver DaemonSet. It matches what's programmatically created by the operator
    #    in https://github.com/openshift/cluster-dns-operator/blob/master/pkg/operator/controller/controller_dns_node_resolver_daemonset.go
    git restore "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml.tmpl
    # 2) Render operand manifest templates like the operator would
    #    Render the DNS DaemonSet
    yq -i '.metadata += {"name": "dns-default", "namespace": "openshift-dns"}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.metadata += {"labels": {"dns.operator.openshift.io/owning-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.selector = {"matchLabels": {"dns.operator.openshift.io/daemonset-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.metadata += {"labels": {"dns.operator.openshift.io/daemonset-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.containers[0].image = "REPLACE_COREDNS_IMAGE"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.containers[1].image = "REPLACE_RBAC_PROXY_IMAGE"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
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
    yq -i '.spec.clusterIP = "REPLACE_CLUSTER_IP"' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    yq -i '.spec.selector = {"dns.operator.openshift.io/daemonset-dns": "default"}' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    sed -i '/#.*automatically managed/d' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml
    # 3) Make MicroShift-specific changes
    #    Fix missing imagePullPolicy
    yq -i '.spec.template.spec.containers[1].imagePullPolicy = "IfNotPresent"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    #    Temporary workaround for MicroShift's missing config parameter when rendering this DaemonSet
    sed -i 's|OPENSHIFT_MARKER=|NAMESERVER=${DNS_DEFAULT_SERVICE_HOST}\n          OPENSHIFT_MARKER=|' "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml
    # 4) Replace MicroShift templating vars (do this last, as yq trips over Go templates)
    sed -i 's|REPLACE_COREDNS_IMAGE|{{ .ReleaseImage.coredns }}|' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    sed -i 's|REPLACE_RBAC_PROXY_IMAGE|{{ .ReleaseImage.kube_rbac_proxy }}|' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    sed -i 's|REPLACE_CLUSTER_IP|{{.ClusterIP}}|' "${REPOROOT}"/assets/components/openshift-dns/dns/service.yaml


    #-- openshift-router ----------------------------------
    # 1) Adopt resource manifests
    #    Replace all openshift-router operand manifests
    rm -f "${REPOROOT}"/assets/components/openshift-router/*
    git restore "${REPOROOT}"/assets/components/openshift-router/serving-certificate.yaml
    cp "${STAGING_DIR}"/cluster-ingress-operator/assets/router/* "${REPOROOT}"/assets/components/openshift-router 2>/dev/null || true
    # Copy embedded manifests applied by the COCMO rather than CVO
    cp "${STAGING_DIR}"/cluster-openshift-controller-manager-operator/bindata/v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrole.yaml "${REPOROOT}"/assets/components/openshift-router || true
    cp "${STAGING_DIR}"/cluster-openshift-controller-manager-operator/bindata/v3.11.0/openshift-controller-manager/ingress-to-route-controller-clusterrolebinding.yaml "${REPOROOT}"/assets/components/openshift-router || true
    #    Restore the openshift-router's service-ca ConfigMap
    git restore "${REPOROOT}"/assets/components/openshift-router/configmap.yaml
    rm -r "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    # 2) Render operand manifest templates like the operator would
    yq -i '.metadata += {"name": "router-default", "namespace": "openshift-ingress"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.metadata += {"labels": {"ingresscontroller.operator.openshift.io/owning-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.minReadySeconds = 30' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.selector = {"matchLabels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.metadata += {"labels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].image = "REPLACE_ROUTER_IMAGE"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "STATS_PORT", "value": "1936"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "RELOAD_INTERVAL", "value": "5s"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_ALLOW_WILDCARD_ROUTES", "value": "false"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_CANONICAL_HOSTNAME", "value": "router-default.apps.REPLACE_CLUSTER_DOMAIN"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
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
    yq -i '.spec.template.spec.containers[0].env += {"name": "ROUTER_USE_PROXY_PROTOCOL", "value": "true"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].env += {"name": "SSL_MIN_VERSION", "value": "TLSv1.2"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # TODO: Generate and volume mount the router-stats-default secret
    # yq -i '.spec.template.spec.containers[0].env += {"name": "STATS_PASSWORD_FILE", "value": "/var/lib/haproxy/conf/metrics-auth/statsPassword"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # yq -i '.spec.template.spec.containers[0].env += {"name": "STATS_USERNAME_FILE", "value": "/var/lib/haproxy/conf/metrics-auth/statsUsername"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports = []' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports += {"name": "http", "containerPort": 80, "protocol": "TCP"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports += {"name": "https", "containerPort": 443, "protocol": "TCP"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports += {"name": "metrics", "containerPort": 1936, "protocol": "TCP"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
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
    # 3) Make MicroShift-specific changes
    #    Set replica count to 1, as we're single-node.
    yq -i '.spec.replicas = 1' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    #    Add hostPorts for routes and metrics (needed as long as there is no load balancer)
    yq -i '.spec.template.spec.containers[0].ports[0].hostPort = 80' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports[1].hostPort = 443' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    # 4) Replace MicroShift templating vars (do this last, as yq trips over Go templates)
    sed -i 's|REPLACE_CLUSTER_DOMAIN|{{ .ClusterDomain }}|' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    sed -i 's|REPLACE_ROUTER_IMAGE|{{ .ReleaseImage.haproxy_router }}|' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml


    #-- service-ca ----------------------------------------
    # 1) Adopt resource manifests
    #    Replace all service-ca operand manifests
    rm -f "${REPOROOT}"/assets/components/service-ca/*
    cp "${STAGING_DIR}"/service-ca-operator/bindata/v4.0.0/controller/* "${REPOROOT}"/assets/components/service-ca || true
    # 2) Render operand manifest templates like the operator would
    # TODO: Remov the following annotations once CPC correctly creates them automatically
    yq -i '.spec.template.spec.containers[0].args = ["-v=2"]' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i '.spec.template.spec.volumes[0].secret.secretName = "REPLACE_TLS_SECRET"' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i '.spec.template.spec.volumes[1].configMap.name = "REPLACE_CA_CONFIG_MAP"' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    yq -i 'del(.metadata.labels)' "${REPOROOT}"/assets/components/service-ca/ns.yaml
    # 3) Make MicroShift-specific changes
    #    Set replica count to 1, as we're single-node.
    yq -i '.spec.replicas = 1' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    # 4) Replace MicroShift templating vars (do this last, as yq trips over Go templates)
    sed -i 's|\${IMAGE}|{{ .ReleaseImage.service_ca_operator }}|' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    sed -i 's|REPLACE_TLS_SECRET|{{.TLSSecret}}|' "${REPOROOT}"/assets/components/service-ca/deployment.yaml
    sed -i 's|REPLACE_CA_CONFIG_MAP|{{.CAConfigMap}}|' "${REPOROOT}"/assets/components/service-ca/deployment.yaml

    #-- ovn-kubernetes -----------------------------------
    # 1) Adopt resource manifests
    #    Replace all ovn-kubernetes manifests
    #
    # NOTE: As long as MicroShift is still based on OpenShift releases that do not yet contain the MicroShift-specific
    #       manifests we're manually updating them as needed for now.
    # TODO: Remove this comments and uncomment the lines below once we've rebased to 4.12.
    #rm -rf "${REPOROOT}"/assets/components/ovn/*
    #cp -r "${STAGING_DIR}"/cluster-network-operator/bindata/network/ovn-kubernetes/microshift/* "${REPOROOT}"/assets/components/ovn || true

    popd >/dev/null
}


# Runs each rebase step in sequence, commiting the step's output to git
rebase_to() {
    local release_image_amd64=$1
    local release_image_arm64=$2

    title "# Rebasing to ${release_image_amd64} and ${release_image_arm64}"
    download_release "${release_image_amd64}" "${release_image_arm64}"

    rebase_branch=rebase-${release_image_amd64#*:}
    rebase_branch=${rebase_branch%-x86_64}
    git branch -D "${rebase_branch}" || true
    git checkout -b "${rebase_branch}"

    update_go_mod
    go mod tidy
    if [[ -n "$(git status -s go.mod go.sum)" ]]; then
        title "## Committing changes to go.mod"
        git add go.mod go.sum
        git commit -m "update go.mod"

        title "## Updating vendor directory"
        make vendor
        if [[ -n "$(git status -s vendor)" ]]; then
            title "## Commiting changes to vendor directory"
            git add vendor
            git commit -m "update vendoring"
        fi
    else
        echo "No changes in go.mod."
    fi

    update_images
    if [[ -n "$(git status -s pkg/release)" ]]; then
        title "## Committing changes to pkg/release"
        git add pkg/release
        git add packaging/crio.conf.d
        git commit -m "update component images"
    else
        echo "No changes in component images."
    fi

    update_manifests
    if [[ -n "$(git status -s assets)" ]]; then
        title "## Committing changes to assets and pkg/assets"
        git add assets pkg/assets
        git commit -m "update manifests"
    else
        echo "No changes to assets."
    fi
}


usage() {
    echo "Usage:"
    echo "$(basename "$0") to RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM         Performs all the steps to rebase to a release image. Specify both amd64 and arm64."
    echo "$(basename "$0") download RELEASE_IMAGE_INTEL RELEASE_IMAGE_ARM   Downloads the content of a release image to disk in preparation for rebasing. Specify both amd64 and arm64."
    echo "$(basename "$0") go.mod                                           Updates the go.mod file to the downloaded release"
    echo "$(basename "$0") generated-apis                                   Regenerates OpenAPIs"
    echo "$(basename "$0") images                                           Rebases the component images to the downloaded release"
    echo "$(basename "$0") manifests                                        Rebases the component manifests to the downloaded release"
    exit 1
}

command=${1:-help}
case "$command" in
    to)
        [[ $# -ne 3 ]] && usage
        rebase_to "$2" "$3"
        ;;
    download)
        [[ $# -ne 3 ]] && usage
        download_release "$2" "$3"
        ;;
    go.mod) update_go_mod;;
    generated-apis) regenerate_openapi;;
    images) update_images;;
    manifests) update_manifests;;
    *) usage;;
esac
