#!/bin/bash

set -euo pipefail
set -x

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/docs-preview/common.sh
source "${SCRIPTDIR}/common.sh"

# Build the documentation builder image
podman build -f "${SCRIPTDIR}/Containerfile" \
    --tag "${PODMAN_IMAGE_TAG}"

if [ ! -d "${PREVIEWDIR}" ]; then
    cd "${OUTPUTDIR}"
    git clone https://github.com/openshift/openshift-docs.git \
        --branch "${DOCS_BRANCH}" \
        openshift-docs-preview
fi
cd "${PREVIEWDIR}"

git checkout "${DOCS_BRANCH}"
git pull

# Run the documentation build
podman run -it --rm -v "$(pwd):/docs:Z" "${PODMAN_IMAGE_TAG}"
