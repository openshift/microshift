#!/bin/bash
#
# This script should be run on the image build server (usually the
# same as the hypervisor).

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

mkdir -p "${IMAGEDIR}"

if [ $# -ne 0 ]; then
    TEMPLATES="$*"
    BUILD_INSTALLER=false
else
    TEMPLATES="${TESTDIR}/image-blueprints/*.toml"
    BUILD_INSTALLER=true
fi

# Determine the version of the RPM in the local repo so we can use it
# in the blueprint templates.
if [ ! -d "${LOCAL_REPO}" ]; then
    error "Run ${SCRIPTDIR}/create_local_repo.sh before building images."
    exit 1
fi
release_info_rpm=$(find "${LOCAL_REPO}" -name 'microshift-release-info-*.rpm')
if [ -z "${release_info_rpm}" ]; then
    error "Failed to find microshift-release-info RPM in ${LOCAL_REPO}"
    exit 1
fi
SOURCE_VERSION=$(rpm -q --queryformat '%{version}' "${release_info_rpm}")

## TEMPLATE VARIABLES
#
# Machine platform type ("x86_64")
UNAME_M=$(uname -m)
export UNAME_M
export LOCAL_REPO  # defined in common.sh
export SOURCE_VERSION  # defined earlier

# Add our sources. It is OK to run these steps repeatedly, if the
# details change they are updated in the service.
mkdir -p "${IMAGEDIR}/package-sources"
# shellcheck disable=SC2231  # allow glob expansion without quotes in for loop
for template in ${TESTDIR}/package-sources/*.toml; do
    name=$(basename "${template}" .toml)
    outfile="${IMAGEDIR}/package-sources/${name}.toml"
    echo "Rendering ${template} to ${outfile}"
    envsubst <"${template}" >"${outfile}"
    echo "Adding package source from ${outfile}"
    sudo composer-cli sources add "${outfile}"
done

# Show details about the available sources to make debugging easier.
for name in $(sudo composer-cli sources list); do
    echo
    echo "Package source: ${name}"
    sudo composer-cli sources info "${name}" | sed -e 's/gpgkeys.*/gpgkeys = .../g'
done

# Given a blueprint filename, extract the name value. It does not have
# to match the filename, but some commands take the file and others
# take the name, so we need to be able to have both.
get_blueprint_name() {
    local filename="${1}"
    tomcli-get "${filename}" name
}

# Track some dynamically created values
BUILDIDS=""

# Upload the blueprint definitions
mkdir -p "${IMAGEDIR}/blueprints"
mkdir -p "${IMAGEDIR}/builds"
# shellcheck disable=SC2231  # allow glob expansion without quotes in for loop
for template in ${TEMPLATES}; do
    echo
    echo "Blueprint ${template}"

    blueprint_file="${IMAGEDIR}/blueprints/$(basename "${template}")"
    echo "Rendering ${template} to ${blueprint_file}"
    envsubst <"${template}" >"${blueprint_file}"

    blueprint=$(get_blueprint_name "${blueprint_file}")

    if sudo composer-cli blueprints list | grep -q "^${blueprint}$"; then
        echo "Removing existing definition of ${blueprint}"
        sudo composer-cli blueprints delete "${blueprint}"
    fi

    echo "Loading new definition of ${blueprint}"
    sudo composer-cli blueprints push "${blueprint_file}"

    echo "Resolving dependencies for ${blueprint}"
    sudo composer-cli blueprints depsolve "${blueprint}"

    echo "Building edge-commit from ${blueprint}"
    buildid=$(sudo composer-cli compose start-ostree \
                   --ref "${blueprint}" \
                   "${blueprint}" \
                   edge-commit \
                  | awk '{print $2}')
    echo "Build ID ${buildid}"
    echo "${buildid}" > "${IMAGEDIR}/builds/${blueprint}.edge-commit"
    BUILDIDS="${BUILDIDS} ${buildid}"
done

if ${BUILD_INSTALLER}; then
    # In the future we may need to build multiple images with different
    # formats but for now we just have one special case to build an
    # installer image in a different format.
    echo "Building image-installer from ${INSTALLER_IMAGE_BLUEPRINT}"
    buildid=$(sudo composer-cli compose start \
                   "${INSTALLER_IMAGE_BLUEPRINT}" \
                   image-installer \
                  | awk '{print $2}')
    echo "Build ID ${buildid}"
    echo "${buildid}" > "${IMAGEDIR}/builds/${blueprint}.image-installer"
    BUILDIDS="${BUILDIDS} ${buildid}"
fi

echo "Waiting for builds to complete..."
# shellcheck disable=SC2086  # pass command arguments quotes to allow word splitting
time "${SCRIPTDIR}/wait_images.py" ${BUILDIDS}
