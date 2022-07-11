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
set -o nounset
set -o pipefail

shopt -s expand_aliases
shopt -s extglob

# debugging options
#trap 'echo "# $BASH_COMMAND"' DEBUG
#set -x

REPOROOT="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")/..")"
STAGING_DIR="$REPOROOT/_output/staging"
PULL_SECRET_FILE="${HOME}/.pull-secret.json"

EMBEDDED_COMPONENTS="openshift-apiserver openshift-controller-manager oauth-apiserver hyperkube etcd"
EMBEDDED_COMPONENT_OPERATORS="cluster-kube-apiserver-operator cluster-openshift-apiserver-operator cluster-kube-controller-manager-operator cluster-openshift-controller-manager-operator cluster-kube-scheduler-operator machine-config-operator"
LOADED_COMPONENTS="cluster-dns-operator cluster-ingress-operator service-ca-operator"


title() {
    echo -e "\E[34m$1\E[00m";
}


# Downloads a release's tools and manifest content into a staging directory,
# then checks out the required components for the rebase at the release's commit.
download_release() {
    local release_image=$1

    rm -rf "${STAGING_DIR}"
    mkdir -p "${STAGING_DIR}"
    pushd "${STAGING_DIR}" >/dev/null

    authentication=""
    if [ -f "${PULL_SECRET_FILE}" ]; then
        authentication="-a ${PULL_SECRET_FILE}"
    else
        >&2 echo "Warning: no pull secret found at ${PULL_SECRET_FILE}"
    fi

    title "# Downloading and extracting ${release_image} tools"
    oc adm release extract ${authentication} --tools "${release_image}"
    echo "Content of release.txt:"
    cat release.txt

    title "# Extracing ${release_image} manifest content"
    mkdir -p release-manifests
    pushd release-manifests >/dev/null
    content=$(oc adm release info ${authentication} --contents "${release_image}")
    echo "${content}" | awk '{ if ($0 ~ /^# [A-Za-z0-9._-]+.yaml$/ || $0 ~ /^# image-references$/ || $0 ~ /^# release-metadata$/) filename = $2; else print >filename;}'
    popd >/dev/null

    title "# Cloning ${release_image} component repos"
    commits=$(oc adm release info ${authentication} --commits -o json "${release_image}")
    echo "${commits}" | jq -r '.references.spec.tags[] | "\(.name) \(.annotations."io.openshift.build.source-location") \(.annotations."io.openshift.build.commit.id")"' > source-commits

    git config --global advice.detachedHead false
    while IFS="" read -r line || [ -n "$line" ]
    do
        component=$(echo "${line}" | cut -d ' ' -f 1)
        repo=$(echo "${line}" | cut -d ' ' -f 2)
        commit=$(echo "${line}" | cut -d ' ' -f 3)
        if [[ "${EMBEDDED_COMPONENTS}" == *"${component}"* ]] || [[ "${LOADED_COMPONENTS}" == *"${component}"* ]] || [[ "${EMBEDDED_COMPONENT_OPERATORS}" == *"${component}"* ]]; then
            title "## Cloning ${repo} at commit ${commit}..."
            git clone "${repo}"
            pushd "${repo##*/}" >/dev/null
            git checkout "${commit}"
            popd >/dev/null
            echo
        fi
    done < source-commits

    popd >/dev/null
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
    commit=$( cd "${STAGING_DIR}/${component}" && git rev-parse HEAD )
    new_modulepath="${repo#https://}${path}"

    echo "go mod edit -replace ${modulepath}=${new_modulepath}@${commit}"
    go mod edit -replace "${modulepath}=${new_modulepath}@${commit}"
    go mod tidy
}

# Returns the line (including trailing comment) in the #{gomod_file} containing the ReplaceDirective for ${module_path}
get_replace_directive() {
    local gomod_file=$1
    local module_path=$2

    # TODO: Handle special case of keyword "replace" being included in the line
    go mod edit -print "${gomod_file}" | grep "^[[:space:]]${module_path}[[:space:]][[:alnum:][:space:].-]*=>"
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
        commit=$( cd "${STAGING_DIR}/kubernetes" && git rev-parse HEAD )
        echo "go mod edit -replace ${modulepath}=${new_modulepath}@${commit}"
        go mod edit -replace "${modulepath}=${new_modulepath}@${commit}"
        go mod tidy
    else
        echo "go mod edit -replace ${modulepath}=${replacement/ /@}"
        go mod edit -replace "${modulepath}=${replacement/ /@}"
    fi
}

# Returns ${line} stripping the trailing comment
strip_comment() {
    local line=$1

    echo "${line%%//*}"
}

# Returns the comment in ${line} if one exists or an empty string if not
get_comment() {
    local line=$1

    comment=${line##*//}
    if [ "${comment}" != "${line}" ]; then
        echo ${comment}
    else
        echo ""
    fi
}

# Validate that ${component} is in the allowed list for the lookup, else exit
valid_component_or_exit() {
    local component=$1
    if [[ ! " etcd kubernetes openshift-apiserver openshift-controller-manager " =~ " ${component} " ]]; then
        echo "error: release reference must be one of [etcd kubernetes openshift-apiserver openshift-controller-manager], have ${component}"
        exit 1
    fi
}

# Updates MicroShift's go.mod file by updating each ReplaceDirective's
# new modulepath-version with that of one of the embedded components.
# The go.mod file needs to specify which component to take this data from
# and this is driven from keywords added as comments after each line of
# ReplaceDirectives:
#   // from ${component}     selects the replacement from the go.mod of ${component}
#   // release ${component}  uses the commit of ${component} as specified in the release image
#   // override [${reason}]  keep existing replacement
# Note directives without keyword comment are skipped with a warning.
update_go_mod() {
    pushd "${STAGING_DIR}" >/dev/null

    title "# Updating go.mod"

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
    if [ ! -f "${STAGING_DIR}/release.txt" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Rebasing release_*.go"
    base_release=$(grep -o -P "^[[:space:]]+Version:[[:space:]]+\K([[:alnum:].-]+)" "${STAGING_DIR}"/release.txt)

    images="$(get_release_images "${REPOROOT}/pkg/release/release.go" | xargs)"

    for arch in amd64; do
        w=$(awk "BEGIN {n=split(\"${images}\", images, \" \"); max=0; for (i=1;i<=n;i++) {if (length(images[i]) > max) {max=length(images[i])}}; print max+2; exit}")
        for i in ${images}; do
            digest=$(awk "/ ${i//_/-} / {print \$2}" release.txt)
            if [[ -n "${digest}" ]]; then
                awk "!/\"${i}\"/ {print \$0} /\"${i}\"/ {printf(\"\\t\\t%-${w}s  %s\n\", \"\\\"${i}\\\":\", \"\\\"${digest}\\\",\")}" \
                    "${REPOROOT}/pkg/release/release_${arch}.go" > t
                mv t "${REPOROOT}/pkg/release/release_${arch}.go"
            fi
        done
    done

    sed -i "/^var Base/c\var Base = \"${base_release}\"" "${REPOROOT}/pkg/release/release.go"

    popd >/dev/null
}


# Updates embedded component manifests by gathering these from various places
# in the staged repos and copying them into the asset directory.
update_manifests() {
    if [ ! -f "${STAGING_DIR}/release.txt" ]; then
        >&2 echo "No release found in ${STAGING_DIR}, you need to download one first."
        exit 1
    fi
    pushd "${STAGING_DIR}" >/dev/null

    title "Extracting timestamp"
    local bindata_timestamp
    bindata_timestamp=$(grep "^Created:" release.txt | cut -f2- -d:)
    date --date="${bindata_timestamp}" '+%s' > "${REPOROOT}/assets/bindata_timestamp.txt"

    title "Rebasing manifests"

    #-- OpenShift control plane ---------------------------
    # 1) Adopt resource manifests
    #    Selectively copy in only those CRD manifests that MicroShift is already using
    for crd in "${REPOROOT}"/assets/crd/*.yaml; do
        cp "${STAGING_DIR}/release-manifests/$(basename "${crd}")" "${REPOROOT}"/assets/crd || true
    done
    #    Replace all SCC manifests.
    rm -f "${REPOROOT}"/assets/scc/*.yaml
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
    cp "${STAGING_DIR}"/cluster-dns-operator/assets/dns/* "${REPOROOT}"/assets/components/openshift-dns/dns 2>/dev/null || true 
    rm -f "${REPOROOT}"/assets/components/openshift-dns/node-resolver/*
    cp "${STAGING_DIR}/"cluster-dns-operator/assets/node-resolver/* "${REPOROOT}"/assets/components/openshift-dns/node-resolver 2>/dev/null || true
    #    Restore the openshift-dns ConfigMap. It's content is the Corefile that the operator generates
    #    in https://github.com/openshift/cluster-dns-operator/blob/master/pkg/operator/controller/controller_dns_configmap.go
    git restore "${REPOROOT}"/assets/components/openshift-dns/dns/configmap.yaml
    #    Restore the template for the node-resolver DaemonSet. It matches what's programmatically created by the operator
    #    in https://github.com/openshift/cluster-dns-operator/blob/master/pkg/operator/controller/controller_dns_node_resolver_daemonset.go
    git restore "${REPOROOT}"/assets/components/openshift-dns/node-resolver/daemonset.yaml.tmpl
    # 2) Render operand manifest templates like the operator would
    #    Render the DNS DaemonSet
    yq -i '.metadata += {"name": "dns-default", "namespace": "openshift-dns"}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.selector = {"matchLabels": {"dns.operator.openshift.io/daemonset-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.metadata += {"labels": {"dns.operator.openshift.io/daemonset-dns": "default"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.containers[0].image = "REPLACE_COREDNS_IMAGE"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.containers[1].image = "REPLACE_RBAC_PROXY_IMAGE"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.nodeSelector = {"kubernetes.io/os": "linux"}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.volumes[0].configMap.name = "dns-default"' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.volumes[1] += {"secret": {"defaultMode": 420, "secretName": "dns-default-metrics-tls"}}' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    yq -i '.spec.template.spec.tolerations = [{"operator": "Exists"}]' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-dns/dns/daemonset.yaml
    #    Render the node-resolver script into the DaemonSet template
    export NODE_RESOLVER_SCRIPT="$(sed 's|^|          |' "${REPOROOT}"/assets/components/openshift-dns/node-resolver/update-node-resolver.sh)"
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
    cp "${STAGING_DIR}"/cluster-ingress-operator/assets/router/* "${REPOROOT}"/assets/components/openshift-router 2>/dev/null || true
    #    Restore the openshift-router's service-ca ConfigMap
    git restore "${REPOROOT}"/assets/components/openshift-router/configmap.yaml
    # 2) Render operand manifest templates like the operator would
    yq -i '.metadata += {"name": "router-default", "namespace": "openshift-ingress"}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.metadata += {"labels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.selector = {"matchLabels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.metadata += {"labels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.volumes[0].secret.secretName = "router-certs-default"' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.metadata += {"annotations": {"service.alpha.openshift.io/serving-cert-secret-name": "router-certs-default"}}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    yq -i '.metadata += {"labels": {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    yq -i '.metadata += {"name": "router-internal-default", "namespace": "openshift-ingress"}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    yq -i '.spec.selector = {"ingresscontroller.operator.openshift.io/deployment-ingresscontroller": "default"}' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-router/service-internal.yaml
    yq -i '.metadata += {"annotations": {"service.alpha.openshift.io/serving-cert-secret-name": "router-certs-default"}}' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    yq -i '.metadata += {"name": "router-external-default"}' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    sed -i '/#.*set at runtime/d' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    # sed -i 's|# Name is set at runtime.|name: router-external-default|' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    #    Set replica count to 1, as we're single-node.
    yq -i '.spec.replicas = 1' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    #    Add hostPorts for routes and metrics
    yq -i '.spec.template.spec.containers[0].ports[0].hostPort = 80' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports[1].hostPort = 443' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    yq -i '.spec.template.spec.containers[0].ports[2].hostPort = 1936' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml
    #    Change LoadBalancer to NodePort as long as we do not add a default LB. Add the necessary nodePorts
    yq -i '.spec.type = "NodePort"' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    yq -i '.spec.ports[0].nodePort = 30001' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    yq -i '.spec.ports[1].nodePort = 30002' "${REPOROOT}"/assets/components/openshift-router/service-cloud.yaml
    # 4) Replace MicroShift templating vars (do this last, as yq trips over Go templates)
    sed -i 's|REPLACE_ROUTER_IMAGE|{{ .ReleaseImage.haproxy_router }}|' "${REPOROOT}"/assets/components/openshift-router/deployment.yaml


    #-- service-ca ----------------------------------------
    # 1) Adopt resource manifests
    #    Replace all service-ca operand manifests
    rm -f "${REPOROOT}"/assets/components/service-ca/*
    cp "${STAGING_DIR}"/service-ca-operator/bindata/v4.0.0/controller/* "${REPOROOT}"/assets/components/service-ca 2>/dev/null || true
    # 2) Render operand manifest templates like the operator would
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

    popd >/dev/null
}


# Runs each rebase step in sequence, commiting the step's output to git
rebase_to() {
    local release_image=$1

    title "# Rebasing to ${release_image}"
    download_release "${release_image}"

    rebase_branch=rebase-${release_image#*:}
    git branch -D "${rebase_branch}" >/dev/null || true
    git checkout -b "${rebase_branch}"

    update_go_mod
    go mod tidy
    if [[ -n "$(git status -s go.mod go.sum)" ]]; then
        title "## Committing changes to go.mod"
        git add go.mod go.sum
        git commit -m "update go.mod"

        title "## Updating vendor directory"
        go mod vendor
        title "## Patching vendor directory"
        git apply scripts/rebase_patches/*
        regenerate_openapi
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
        git commit -m "update component images"
    else
        echo "No changes in component images."
    fi

    update_manifests
    if [[ -n "$(git status -s assets)" ]]; then
        title "## Updating bindata"
        "${REPOROOT}"/scripts/bindata.sh
        title "## Committing changes to assets and pkg/assets"
        git add assets pkg/assets
        git commit -m "update manifests"
    else
        echo "No changes to assets."
    fi
}


usage() {
    echo "Usage:"
    echo "$(basename "$0") to RELEASE_IMAGE          Performs all the steps to rebase to RELEASE_IMAGE"
    echo "$(basename "$0") download RELEASE_IMAGE    Downloads the content of RELEASE_IMAGE to disk in preparation for rebasing"
    echo "$(basename "$0") go.mod                    Updates the go.mod file to the downloaded release"
    echo "$(basename "$0") generated-apis            Regenerates OpenAPIs"
    echo "$(basename "$0") images                    Rebases the component images to the downloaded release"
    echo "$(basename "$0") manifests                 Rebases the component manifests to the downloaded release"
    exit 1
}

command=${1:-help}
case "$command" in
    to)
        [[ $# -ne 2 ]] && usage
        rebase_to "$2"
        ;;
    download)
        [[ $# -ne 2 ]] && usage
        download_release "$2"
        ;;
    go.mod) update_go_mod;;
    generated-apis) regenerate_openapi;;
    images) update_images;;
    manifests) update_manifests;;
    *) usage;;
esac
