#!/bin/bash

set -euo pipefail
set -x

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/docs-preview/common.sh
source "${SCRIPTDIR}/common.sh"

podman rm  -f "${PODMAN_CONTAINER_NAME}"
podman run --rm \
       --detach \
       --name "${PODMAN_CONTAINER_NAME}" \
       -p 8081:80 \
       -v "${PREVIEWDIR}/_preview/microshift/${DOCS_BRANCH}:/usr/local/apache2/htdocs:Z" \
       docker.io/library/httpd

echo "Open the http://0.0.0.0:8081/microshift_welcome/ URL"
