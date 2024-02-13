#!/usr/bin/env bash

# shellcheck disable=SC2034  # used elsewhere
ROOTDIR="$(cd "${SCRIPTDIR}/../.." && pwd)"

# shellcheck disable=SC2034  # used elsewhere
OUTPUTDIR="${ROOTDIR}/_output"

# shellcheck disable=SC2034  # used elsewhere
PREVIEWDIR="${OUTPUTDIR}/openshift-docs-preview"

# shellcheck disable=SC2034  # used elsewhere
BRANCH=enterprise-4.15
