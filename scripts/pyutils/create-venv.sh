#!/usr/bin/env bash

# shellcheck disable=all
set -o nounset
set -o errexit
set -o pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR=$(realpath "${SCRIPT_DIR}/../..")
DEFAULT_DEST_DIR="${ROOT_DIR}/_output/pyutils"
DEST_DIR="${DEST_DIR:-${DEFAULT_DEST_DIR}}"

if [ ! -d "${DEST_DIR}" ]; then
    echo "Setting up virtualenv in ${DEST_DIR}"
    python3 -m venv "${DEST_DIR}"
    "${DEST_DIR}/bin/python3" -m pip install --upgrade pip
    "${DEST_DIR}/bin/python3" -m pip install -r "${SCRIPT_DIR}/requirements.txt"
fi
