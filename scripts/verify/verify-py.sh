#!/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
VENV_DIR="${ROOTDIR}/_output"
VENV="${VENV_DIR}/venv"
REQ_FILE=${ROOTDIR}/scripts/requirements.txt

create_venv() {
    local vpython="${VENV}/bin/python3"

    [ -f "${REQ_FILE}" ] || { echo "${REQ_FILE} is not present"; exit 1; }
    [ -d "${VENV_DIR}" ] || mkdir -p "${VENV_DIR}"

    echo "Creating venv in '${VENV}' and installing packages..."
    python3 -m venv "${VENV}"
    ${vpython} -m pip install --upgrade pip
    ${vpython} -m pip install -r "${REQ_FILE}"
    echo "Done!"
}

run_pylint() {
    local pylint="${VENV}/bin/pylint"

    if ! command -v "${pylint}" &>/dev/null; then
        echo "Installing pylint..."
        create_venv
    fi

    local pyfiles
    pyfiles=$(find . -type d \( -path ./_output -o -path ./vendor -o -path ./assets -o -path ./etcd/vendor \) -prune -o -name '*.py' -print)
    echo "Checking: ${pyfiles}"

    # shellcheck disable=SC2086
    ${pylint} --rcfile="${ROOTDIR}/.pylintrc" ${pyfiles}

}

run_pylint
