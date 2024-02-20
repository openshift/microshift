#!/bin/bash

set -e

SCRIPTDIR="$(dirname "${BASH_SOURCE[0]}")"
REPOROOT="$(git rev-parse --show-toplevel)"
OUTPUT_DIR="${REPOROOT}/_output"
ENVDIR="${OUTPUT_DIR}/jira"

if [ ! -d "${ENVDIR}" ]; then
    echo "Setting up required tools..."
    mkdir -p "${OUTPUT_DIR}"
    python3 -m venv "${ENVDIR}"
    "${ENVDIR}/bin/pip3" install -r "${SCRIPTDIR}/requirements.txt"
fi

version_file="${REPOROOT}/Makefile.version.$(uname -m).var"
version=$(cut -f2 -d= "${version_file}" | cut -f1-2 -d. | sed -e 's/ //g')
export DEFAULT_TARGET_VERSION="openshift-${version}"

"${ENVDIR}/bin/python3" "${SCRIPTDIR}/manage_ticket.py" "$@"
