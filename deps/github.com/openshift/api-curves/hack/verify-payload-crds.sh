#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"
source "$(dirname "${BASH_SOURCE}")/update-payload-crds.sh"

files=""

# Check there's no diff between the files in their canonical location
# and the payload-manifests location.
for f in ${crd_globs}; do
    basename=$(basename "${f}")
    files+=${basename},
    echo "Verifying diff on ${basename}"
    diff "$f" "${SCRIPT_ROOT}/payload-manifests/crds/${basename}"
done
	
files=$(echo "${files}" | tr "," "\n")

# Check that we haven't accidentally added any files that aren't tracked
# by the crd_globs into the payload CRDs folder.
for f in "${SCRIPT_ROOT}/payload-manifests/crds/"*; do
    basename=$(basename "${f}")
    if ! echo "${files}" | grep -F -q -x "${basename}"; then
        echo "Found untracked file ${basename} in payload CRD manifests.  Please add the file to crd_globs in hack/update-payload-crds.sh."
        exit 1
    fi
done
