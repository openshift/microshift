#!/bin/bash

set -euo pipefail
set -x

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/docs-preview/common.sh
source "${SCRIPTDIR}/common.sh"

podman run --rm \
       --detach \
       --name openshift-docs-preview-httpd \
       -p 8081:80 \
       -v "${PREVIEWDIR}/_preview/microshift/${BRANCH}:/usr/local/apache2/htdocs:Z" \
       docker.io/library/httpd
