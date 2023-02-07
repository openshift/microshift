#!/bin/bash

set -o nounset
set -o errexit
set -o pipefail
set -x

echo "Environment:"
printenv

cp /secrets/ci-pull-secret/.dockercfg "$HOME/.pull-secret.json" || {
    echo "WARN: Could not copy registry secret file"
}

release_amd64="$(oc get configmap/release-release-images-latest -o yaml \
    | yq '.data."release-images-latest.yaml"' \
    | jq -r '.metadata.name')"
release_arm64="$(oc get configmap/release-release-images-arm64-latest -o yaml \
    | yq '.data."release-images-arm64-latest.yaml"' \
    | jq -r '.metadata.name')"
# LVMS is not tracked in the OCP release image.  Instead, rely on the latest X.Y stream as the release image.
#  LVMS also does not cut nightly releases where ocp-release does.  This means that latest ocp-releases' y-stream
#  can increment independently from LVMS, and will usually be 1 y-stream ahead of LVMS in-between OCP releases.
#  For example, ocp-release at 4.13 will more often than not correspond to 4.12 LVMS, until the official 4.13 release when
#  both components will be 4.13.
release_lvms="v4.12"

pullspec_release_amd64="registry.ci.openshift.org/ocp/release:${release_amd64}"
pullspec_release_arm64="registry.ci.openshift.org/ocp-arm64/release-arm64:${release_arm64}"
# Since LVMS is not part of the release payload, it is not kept in CI. Use the latest z-stream that coincides with the release payload's X.Y version
pullspec_release_lvms="registry.access.redhat.com/lvms4/lvms-operator-bundle:${release_lvms}"

APP_ID=$(cat /secrets/pr-creds/app_id) \
KEY=/secrets/pr-creds/key.pem \
ORG=${ORG:-openshift} \
REPO=${REPO:-microshift} \
AMD64_RELEASE=${pullspec_release_amd64} \
ARM64_RELEASE=${pullspec_release_arm64} \
LVMS_RELEASE=${pullspec_release_lvms} \
./scripts/auto-rebase/rebase.py
