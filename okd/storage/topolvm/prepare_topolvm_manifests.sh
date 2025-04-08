#!/bin/bash

set -euo pipefail

# prepares manifests for topolvm upsteam running on single node microshift.

CERT_MANAGER_VERSION=v1.16.1
TOPO_LVM_VERSION=v15.5.2

cat - <<EOF > manifests/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: topolvm-system
  labels:
    openshift.io/run-level: "0"
    pod-security.kubernetes.io/enforce: privileged
    pod-security.kubernetes.io/audit: privileged
    pod-security.kubernetes.io/warn: privileged  
EOF


wget https://github.com/cert-manager/cert-manager/releases/download/${CERT_MANAGER_VERSION}/cert-manager.yaml -O manifests/cert-manager.yaml

# install helm
curl -fsSL -o /tmp/get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3
chmod 700 /tmp/get_helm.sh
/tmp/get_helm.sh

# generate manifests using helm
# NOTE: this will produce multi-arch manifest, support both amd64 and arm64 
helm template --include-crds --namespace=topolvm-system --version=${TOPO_LVM_VERSION} topolvm topolvm/topolvm > manifests/topolvm.yaml

# patch replicas to 1
# shellcheck disable=SC2016
yq 'select(.kind == "Deployment").spec.replicas = 1' -i manifests/topolvm.yaml

# generate kustomize
cat - <<EOF >manifests/kustomization.yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - namespace.yaml
  - cert-manager.yaml
  - topolvm.yaml
EOF
