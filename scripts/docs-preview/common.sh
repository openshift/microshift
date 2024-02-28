#!/usr/bin/env bash

# shellcheck disable=SC2034  # used elsewhere
ROOTDIR="$(cd "${SCRIPTDIR}/../.." && pwd)"

# shellcheck disable=SC2034  # used elsewhere
OUTPUTDIR="${ROOTDIR}/_output"

# shellcheck disable=SC2034  # used elsewhere
PREVIEWDIR="${OUTPUTDIR}/openshift-docs-preview"

# shellcheck disable=SC2034  # used elsewhere
PODMAN_IMAGE_TAG="ushift-docs-asciibinder"

# shellcheck disable=SC2034  # used elsewhere
PODMAN_CONTAINER_NAME="ushift-docs-httpd"

# shellcheck disable=SC2034  # used elsewhere
DOCS_BRANCH=${DOCS_BRANCH:-enterprise-4.16}
