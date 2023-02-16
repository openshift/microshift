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

release_amd64="$(oc get configmap/release-release-images-latest -o yaml \
    | yq '.data."release-images-latest.yaml"' \
    | jq -r '.metadata.name')"
release_arm64="$(oc get configmap/release-release-images-arm64-latest -o yaml \
    | yq '.data."release-images-arm64-latest.yaml"' \
    | jq -r '.metadata.name')"

pullspec_release_amd64="registry.ci.openshift.org/ocp/release:${release_amd64}"
pullspec_release_arm64="registry.ci.openshift.org/ocp-arm64/release-arm64:${release_arm64}"

branch_name="$(git branch --show-current)"
if [[ "${branch_name}" == release-4* ]]; then
    pullspec_release_lvms="registry.access.redhat.com/lvms4/lvms-operator-bundle:v${branch_name#*-}"
else
    quay_tags=$(curl https://quay.io/api/v1/repository/rhceph-dev/lvms4-lvms-operator-bundle/tag/)
    latest_digest=$(echo "${quay_tags}" | jq -r '.tags | sort_by(.start_ts) | reverse | .[0].manifest_digest')
    pullspec_release_lvms="quay.io/rhceph-dev/lvms4-lvms-operator-bundle@${latest_digest}"
fi

APP_ID=$(cat /secrets/pr-creds/app_id) \
KEY=/secrets/pr-creds/key.pem \
ORG=${ORG:-openshift} \
REPO=${REPO:-microshift} \
AMD64_RELEASE=${pullspec_release_amd64} \
ARM64_RELEASE=${pullspec_release_arm64} \
LVMS_RELEASE=${pullspec_release_lvms} \
./scripts/auto-rebase/rebase.py
