#!/bin/bash
# shellcheck disable=all
set -o nounset
set -o errexit
set -o pipefail
set -x

check_semver_suffix() {
    local version=$1

    # Check if the version is not empty
    if [[ -z "$version" ]]; then
        echo "false"
        return 1
    fi

    # Check if the version string contains a numeric suffix of the form -xx
    if [[ $version =~ -[0-9]+$ ]]; then
        echo "true"
        return 0
    else
        echo "false"
        return 1
    fi
}


echo "Environment:"
printenv

if [[ "$JOB_NAME" == rehearse* ]]; then
    echo "INFO: \$JOB_NAME starts with rehearse - running in DRY RUN mode"
    export DRY_RUN=y
fi

# This file is executed by the rebase jobs.
# Rebase jobs run in CI as a containers and they use 'root' images defined per branch such as:
# - rhel-9-release-golang-1.20-openshift-4.15
# - rhel-9-release-golang-1.21-openshift-4.16
# - rhel-9-release-golang-1.22-openshift-4.17

# Following code updates Go version in configure-vm.sh and microshift.spec
# based on the version of Go inside the rebase job's container.
#
# It's not part of rebase.sh because we don't want this during manual rebases.
# It's before rebase.py because we want this to be part of the rebase PR.
# Go version in go.mods are updated in rebase.sh.
go_version=$(go version 2>/dev/null | awk '{print $3}' | tr -d '[a-z]')
sed -i "s/^GO_VER=.*/GO_VER=${go_version}/" ./scripts/devenv-builder/configure-vm.sh
go_version_xy="$(echo "${go_version}" | cut -f1-2 -d.)"
sed -i "s/^%global golang_version .*/%global golang_version ${go_version_xy}/" ./packaging/rpm/microshift.spec

if [[ -n "$(git status -s ./scripts/devenv-builder/configure-vm.sh ./packaging/rpm/microshift.spec)" ]]; then
    echo "Updating Go versions in microshift.spec and configure-vm.sh"
    git add ./scripts/devenv-builder/configure-vm.sh ./packaging/rpm/microshift.spec
    git commit -m "Update Go version"
fi

cp /secrets/import-secret/.dockercfg "$HOME/.pull-secret.json" || {
    echo "WARN: Could not copy registry secret file"
}

# log in into cluster's registry
oc registry login --to=/tmp/registry.json
release_amd64="$(oc image info --registry-config=/tmp/registry.json $OPENSHIFT_RELEASE_IMAGE -o json | jq -r '.config.config.Labels."io.openshift.release"')"
release_arm64="$(oc image info --registry-config=/tmp/registry.json $OPENSHIFT_RELEASE_IMAGE_ARM -o json | jq -r '.config.config.Labels."io.openshift.release"')"

pullspec_release_amd64="registry.ci.openshift.org/ocp/release:${release_amd64}"
pullspec_release_arm64="registry.ci.openshift.org/ocp-arm64/release-arm64:${release_arm64}"

# RHOAI is not part of the OpenShift release image and it has
# different cadence and versioning scheme.
# Following variable with RHOAI Operator Bundle reference needs to be updated manually,
# preferably including only X.Y and skipping Z, so the rebase automatically picks up new Z-stream releases.
#
# New references can be obtained from:
# https://catalog.redhat.com/software/containers/rhoai/odh-operator-bundle/659803ca929f3c931af06f28
rhoai_release="registry.redhat.io/rhoai/odh-operator-bundle:v2.18"

APP_ID=$(cat /secrets/pr-creds/app_id) \
KEY=/secrets/pr-creds/key.pem \
ORG=${ORG:-openshift} \
REPO=${REPO:-microshift} \
AMD64_RELEASE=${pullspec_release_amd64} \
ARM64_RELEASE=${pullspec_release_arm64} \
RHOAI_RELEASE=${rhoai_release} \
./scripts/auto-rebase/rebase.py

# LVMS is not tracked in the OCP release image.  Instead, rely on the
#  latest X.Y stream as the release image.  LVMS also does not cut
#  nightly releases where ocp-release does.  This means that latest
#  ocp-releases' y-stream can increment independently from LVMS, and
#  will usually be 1 y-stream ahead of LVMS in-between OCP releases.
#  For example, ocp-release at 4.13 will more often than not
#  correspond to 4.12 LVMS, until the official 4.13 release when both
#  components will be 4.13.
release_lvms="v4.17.0-43"

# Since LVMS is not part of the release payload, it is not kept in
# CI. Use the latest z-stream that coincides with the release
# payload's X.Y version
pullspec_release_lvms="registry.redhat.io/lvms4/lvms-operator-bundle:${release_lvms}"
# A unreleased candidate doesnt exist in the official registry, so fallback to the lvms_dev namespace, which contains
# the latest lvms release candidate replicated from CPaaS into quay
pullspec_release_lvms_fallback="quay.io/lvms_dev/lvms4-lvms-operator-bundle:${release_lvms}"

lvms_is_candidate_build=$(check_semver_suffix "${release_lvms}")
if [ "$lvms_is_candidate_build" == "false" ]; then
    ./scripts/auto-rebase/rebase-lvms.sh to "${pullspec_release_lvms}"
else
    ./scripts/auto-rebase/rebase-lvms.sh to "${pullspec_release_lvms_fallback}"
fi


if [[ "${JOB_TYPE}" == "presubmit" ]]; then
    # Verify the assets after the rebase to make sure
    # nightly job will not fail on assets verification.
    make verify-assets
fi
