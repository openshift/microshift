#!/bin/bash
set -euo pipefail

BUILD_ARCH=$(uname -m)
TEMP_PATH=$(mktemp -d)
trap 'rm -rf ${TEMP_PATH}' 'EXIT'

if [ $# -ne 2 ] ; then
   echo "Usage: $(basename "$0") <microshift_version> <download_path>"
   echo ""
   echo "Download MicroShift RPMs for the ${BUILD_ARCH} architecture and the specified version"
   exit 1
fi

VERSION=$1
DOWNLOAD_PATH=$2
mkdir -p "${DOWNLOAD_PATH}"
OSVERSION=9
[ "${VERSION}" = "4.12" ] && OSVERSION=8

OCP_REPO_NAME="rhocp-${VERSION}-for-rhel-${OSVERSION}-${BUILD_ARCH}-rpms"

# Download all RPMs from the repository to a temporary directory
# (reposync cannot filter by name)
echo "Downloading MicroShift RPMs from ${OCP_REPO_NAME}..."
reposync -n -a "${BUILD_ARCH}" -a noarch --download-path "${TEMP_PATH}" --repo="${OCP_REPO_NAME}" >/dev/null

# Copy MicroShift RPMs to the download path
find "${TEMP_PATH}" -type f -name microshift\* -exec cp -f {} "${DOWNLOAD_PATH}" \;
echo "MicroShift RPMs downloaded to ${DOWNLOAD_PATH}"
