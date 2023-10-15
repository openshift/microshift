#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR=$(realpath "${SCRIPT_DIR}/..")
DEFAULT_DEST_DIR="${ROOT_DIR}/_output/conformance-tests"
DEST_DIR="${DEST_DIR:-${DEFAULT_DEST_DIR}}"
[ -d "${DEST_DIR}" ] || mkdir -p "${DEST_DIR}"
DEST_DIR="$(realpath "${DEST_DIR}")"

cd "${SCRIPT_DIR}"

echo "Building openshift-tests"
make build

if ! which oc &>/dev/null; then
    echo "oc binary not found"
    exit 1
fi

if ! which kubectl &>/dev/null; then
    echo "kubectl binary not found"
    exit 1
fi

echo "Executing conformance tests. Using Kubeconfig ${KUBECONFIG}. Writing output to ${DEST_DIR}"

"${SCRIPT_DIR}"/openshift-tests run openshift/conformance -v 2 --provider=none -f "${SCRIPT_DIR}/suite.txt" -o "${DEST_DIR}/e2e.log" --junit-dir "${DEST_DIR}/junit"
