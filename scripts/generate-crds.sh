#!/bin/bash

set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
CONTROLLER_BIN="${ROOTDIR}/_output/bin/controller-gen"
CODEGEN_DIR="${ROOTDIR}/deps/github.com/openshift/kubernetes/staging/src/k8s.io/code-generator"

OUTPUT_PKG="github.com/openshift/microshift/pkg/generated"

pushd "${ROOTDIR}" &>/dev/null

echo "Generating deepcopy methods"
${CONTROLLER_BIN} object paths=./pkg/apis/microshift/v1alpha1/

echo "Generating CRD YAML"
${CONTROLLER_BIN} crd paths=./pkg/apis/microshift/v1alpha1/ output:crd:artifacts:config=assets/crd/

echo "Generating typed clientset, listers, informers"
# shellcheck source=/dev/null
source "${CODEGEN_DIR}/kube_codegen.sh"

kube::codegen::gen_client \
    --output-dir "${ROOTDIR}/pkg/generated" \
    --output-pkg "${OUTPUT_PKG}" \
    --boilerplate "${ROOTDIR}/scripts/boilerplate.go.txt" \
    --with-watch \
    "${ROOTDIR}/pkg/apis"

popd &>/dev/null

echo "Done"
