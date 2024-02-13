#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR=$(realpath "${SCRIPT_DIR}/..")
DEFAULT_DEST_DIR="${ROOT_DIR}/_output/conformance-tests"
DEST_DIR="${DEST_DIR:-${DEFAULT_DEST_DIR}}"
[ -d "${DEST_DIR}" ] || mkdir -p "${DEST_DIR}"
DEST_DIR="$(realpath "${DEST_DIR}")"
RESTRICTED="${RESTRICTED:-true}"

cd "${SCRIPT_DIR}"

echo "Preparing submodule"
git submodule update --init
pushd origin
git reset --hard HEAD
popd

echo "Applying MicroShift patches"
pushd origin
for patch_file in "${SCRIPT_DIR}"/patches/*.patch; do
    echo "Checking patch ${patch_file}"
    if git apply --check "${patch_file}" 2> /dev/null; then
        git apply "${patch_file}"
        echo "Patch applied"
    else
        echo "Patch was already applied"
    fi
done
popd

echo "Building openshift-tests"
pushd origin
make openshift-tests
mv openshift-tests "${DEST_DIR}/"
popd

if ! which oc &>/dev/null; then
    echo "oc binary not found"
    exit 1
fi

if ! which kubectl &>/dev/null; then
    echo "kubectl binary not found"
    exit 1
fi

echo "Executing conformance tests. Using Kubeconfig ${KUBECONFIG}. Writing output to ${DEST_DIR}"

SUITE_FILE=""
if [ "${RESTRICTED}" = "true" ]; then
    SUITE_FILE="-f ${SCRIPT_DIR}/suite.txt"
fi

"${DEST_DIR}"/openshift-tests run openshift/conformance -v 2 --provider=none ${SUITE_FILE} -o "${DEST_DIR}/e2e.log" --junit-dir "${DEST_DIR}/junit"
