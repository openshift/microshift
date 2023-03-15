#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail
set -x

echo "Environment:"
printenv

if [[ "$JOB_NAME" == rehearse* ]]; then
    echo "INFO: \$JOB_NAME starts with rehearse - running in DRY RUN mode"
    export DRY_RUN=y
fi

cp /secrets/ci-pull-secret/.dockercfg "$HOME/.pull-secret.json" || {
    echo "WARN: Could not copy registry secret file"
}

release_amd64="$(oc get configmap/release-release-images-nightly-amd64 -o yaml |
    yq '.data."release-images-nightly-amd64.yaml"' |
    jq -r '.metadata.name')"
release_arm64="$(oc get configmap/release-release-images-nightly-arm64 -o yaml |
    yq '.data."release-images-nightly-arm64.yaml"' |
    jq -r '.metadata.name')"

pullspec_release_amd64="registry.ci.openshift.org/ocp/release:${release_amd64}"
pullspec_release_arm64="registry.ci.openshift.org/ocp-arm64/release-arm64:${release_arm64}"

./scripts/auto-rebase/rebase.sh to "${pullspec_release_amd64}" "${pullspec_release_arm64}"

APP_ID=$(cat /secrets/pr-creds/app_id) \
KEY=/secrets/pr-creds/key.pem \
ORG=${ORG:-openshift} \
REPO=${REPO:-microshift} \
AMD64_RELEASE=${pullspec_release_amd64} \
ARM64_RELEASE=${pullspec_release_arm64} \
    ./scripts/auto-rebase/rebase.py
