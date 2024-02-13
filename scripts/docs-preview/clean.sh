#!/bin/bash

set -euo pipefail
set -x

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/docs-preview/common.sh
source "${SCRIPTDIR}/common.sh"

# Delete sequence:
# - Running docs container
# - Build image
# - All unused images, i.e. Fedora, httpd, etc.
# - Documentation sources
podman rm  -f "${PODMAN_CONTAINER_NAME}"
podman rmi -f "${PODMAN_IMAGE_TAG}" 
podman rmi -a || true
rm -rf "${PREVIEWDIR}"
