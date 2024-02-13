#!/bin/bash

set -euo pipefail
set -x

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/docs-preview/common.sh
source "${SCRIPTDIR}/common.sh"

podman build -f "${SCRIPTDIR}/Containerfile" --tag asciibinder

if [ ! -d "${PREVIEWDIR}" ]; then
    cd "${OUTPUTDIR}"
    git clone https://github.com/openshift/openshift-docs.git \
        --branch "${BRANCH}" \
        openshift-docs-preview
fi
cd "${PREVIEWDIR}"

git pull

podman run -it --rm -v "$(pwd):/docs:Z" asciibinder
