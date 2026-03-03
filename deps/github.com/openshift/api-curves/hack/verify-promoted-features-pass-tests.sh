#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

TMP_ROOT="${SCRIPT_ROOT}/_tmp"

cleanup() {
  rm -rf "${TMP_ROOT}"
}
trap "cleanup" EXIT SIGINT

cleanup

mkdir -p ${TMP_ROOT}/previous-openshift-api/payload-manifests/featuregates

# Use PULL_BASE_REF for CI, otherwise use master unless overridden.
COMPARISON_BASE=${COMPARISON_BASE:-${PULL_BASE_SHA:-"master"}}
echo "comparing against ${COMPARISON_BASE}"


featureGateFiles=$(git show ${COMPARISON_BASE}:payload-manifests/featuregates  | tail -n +3)

while IFS= read -r featureGateFile; do
  echo "writing ${COMPARISON_BASE}:${featureGateFile} to temp"
  git show ${COMPARISON_BASE}:payload-manifests/featuregates/${featureGateFile} > ${TMP_ROOT}/previous-openshift-api/payload-manifests/featuregates/${featureGateFile}
done <<< "${featureGateFiles}"

# this appears to prevent the cleanup from doing anything
# Use a trap so this gets printed at the end of the log even when we exit early due to an error/failure
#trap 'echo This verifier checks changed feature gates have tests.' EXIT

# Build codegen-crds when it's not present and not overridden for a specific file.
if [ -z "${CODEGEN:-}" ];then
  ${TOOLS_MAKE} codegen
  CODEGEN="${TOOLS_OUTPUT}/codegen"
fi

"${CODEGEN}" featuregate-test-analyzer