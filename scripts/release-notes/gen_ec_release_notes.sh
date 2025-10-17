#!/bin/bash

# shellcheck disable=all
set -o nounset
set -o errexit
set -o pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR=$(realpath "${SCRIPT_DIR}/../..")
DEFAULT_DEST_DIR="${ROOT_DIR}/_output/releasenotes"
DEST_DIR="${DEST_DIR:-${DEFAULT_DEST_DIR}}"

if [ ! -d "${DEST_DIR}" ]; then
    echo "Setting up virtualenv in ${DEST_DIR}"
    python3 -m venv --system-site-packages "${DEST_DIR}"
    "${DEST_DIR}/bin/python3" -m pip install --upgrade pip
    "${DEST_DIR}/bin/python3" -m pip install -r "${SCRIPT_DIR}/requirements.txt"
fi
source "${DEST_DIR}/bin/activate"

if [ "$#" -ne 0 ] && [ "$1" == "--rhocp" ]; then
    shift
    python3 "${SCRIPT_DIR}/gen_release_notes.py" "$@"
else
    python3 "${SCRIPT_DIR}/gen_ec_release_notes.py" "$@"
fi
