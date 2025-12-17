#!/bin/bash

# shellcheck disable=all
set -o nounset
set -o errexit
set -o pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR=$(realpath "${SCRIPT_DIR}/../..")
DEFAULT_DEST_DIR="${ROOT_DIR}/_output/releasenotes"
DEST_DIR="${DEST_DIR:-${DEFAULT_DEST_DIR}}"

function usage() {
    echo ""
    echo "Usage: $(basename "$0") {mirror, rhocp} [script_opts...]"
    echo "   mirror         Execute gen_gh_releases_from_mirror.py: Generate GH releases of pre-release (ECs, RCs) MicroShift RPMs"
    echo "   rhocp          Execute gen_gh_releases_from_rhocp.py: Generate GH release of released (on RHOCP repositories) MicroShift RPMs"
    echo "   script_opts    Options passed through to executed python script"
    echo ""
    exit 1
}

if [ ! -d "${DEST_DIR}" ]; then
    echo "Setting up virtualenv in ${DEST_DIR}"
    python3 -m venv --system-site-packages "${DEST_DIR}"
    "${DEST_DIR}/bin/python3" -m pip install --upgrade pip
    "${DEST_DIR}/bin/python3" -m pip install -r "${SCRIPT_DIR}/requirements.txt"
fi
source "${DEST_DIR}/bin/activate"

if [ "$#" -eq 0 ]; then
    echo "Script requires at least 1 argument."
    usage
fi

cmd=$1
shift

if [[ "${cmd}" == "mirror" ]]; then
    python3 "${SCRIPT_DIR}/gen_gh_releases_from_mirror.py" "$@"
elif [[ "${cmd}" == "rhocp" ]]; then
    python3 "${SCRIPT_DIR}/gen_gh_releases_from_rhocp.py" "$@"
else
    echo "Unrecognized argument: ${cmd}"
    usage
fi
