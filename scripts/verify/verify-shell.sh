#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
pushd "${ROOTDIR}" &> /dev/null

# Install the tool
./scripts/fetch_tools.sh shellcheck
SHELL_CHECK=./_output/bin/shellcheck

# Ignore paths containing external sources
IGNORE_PATHS="-path ./_output -o -path ./vendor -o -path ./assets -o -path ./etcd/vendor -o -path ./hack -o -path ./docs"
# Ignore files managed upstream
IGNORE_FILES="configure-ovs.sh"

# Find the files to be checked
# shellcheck disable=SC2086
# The path list must not be quoted to allow multiple arguments
CHECK_FILE_LIST=$(find . \( -type d \( ${IGNORE_PATHS} \) -o -name "${IGNORE_FILES}" \) -prune -o -name '*.sh' -print)

for f in ${CHECK_FILE_LIST} ; do
    echo "shellcheck: ${f}"
    "${SHELL_CHECK}" -x "${f}"
done
