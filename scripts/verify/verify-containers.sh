#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)

"${ROOTDIR}/scripts/fetch_tools.sh" gomplate
export GOMPLATE=${ROOTDIR}/_output/bin/gomplate

cd "${ROOTDIR}"
FILES=$(find . -iname '*containerfile*' -o -iname '*dockerfile*' | grep -v "vendor\|_output\|origin\|.git")
# When run inside a container, the file contents are redirected via stdin and
# the output of errors does not contain the file path. Work around this issue
# by replacing the '^-:' token in the output by the actual file name.
tmpdir=$(mktemp -d)
for f in ${FILES} ; do
    echo "${f}"
    temp_file="${tmpdir}/$(basename "${f}").temp"
    "${GOMPLATE}" --file "${f}" >"${temp_file}"
    podman run --rm -i \
        -v "${ROOTDIR}/.hadolint.yaml:/.hadolint.yaml:ro" \
        ghcr.io/hadolint/hadolint:2.12.0 < "${temp_file}" | sed "s|^-:|${f}:|"
done
trap 'rm -rf "${tmpdir}"' EXIT
