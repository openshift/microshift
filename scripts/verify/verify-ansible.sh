#!/usr/bin/bash
set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
UV_VERSION="0.12.3"
UV="${ROOTDIR}/_output/bin/uv-${UV_VERSION}"
VENV="${ROOTDIR}/_output/ansibleenv"
PYTHON_VERSION="3.12.13"
PYTHON_STAMP="${VENV}/.python-version"
REQ_FILE="${ROOTDIR}/scripts/requirements-ansible.txt"
REQ_STAMP="${VENV}/.requirements.txt"
REQ_LOCK_FILE="${ROOTDIR}/scripts/requirements-ansible.lock"
REQ_LOCK_STAMP="${VENV}/.requirements.lock"
COLLECTION_REQ_FILE="${ROOTDIR}/scripts/requirements-ansible.yml"
COLLECTION_REQ_STAMP="${VENV}/.requirements.yml"
COLLECTIONS_PATH="${VENV}/collections"

export UV_CACHE_DIR="${ROOTDIR}/_output/uv-cache"
export UV_MANAGED_PYTHON=1
export UV_PYTHON_INSTALL_DIR="${ROOTDIR}/_output/uv-python"
export ANSIBLE_COLLECTIONS_PATH="${COLLECTIONS_PATH}"
export ANSIBLE_HOME="${ROOTDIR}/_output/ansible-home"
export ANSIBLE_INVENTORY="${ROOTDIR}/ansible/inventory/inventory"
export ANSIBLE_LOCAL_TEMP="${ROOTDIR}/_output/ansible-tmp/local"
export ANSIBLE_REMOTE_TEMP=/tmp
export XDG_CACHE_HOME="${ROOTDIR}/_output/ansible-cache"

mkdir -p "${ANSIBLE_HOME}" "${ANSIBLE_LOCAL_TEMP}" "${XDG_CACHE_HOME}"

create_venv() {
    echo "Creating Ansible lint environment in '${VENV}'..."
    mkdir -p "${ROOTDIR}/_output"
    "${UV}" python install --no-bin "${PYTHON_VERSION}"
    "${UV}" venv --clear --python "${PYTHON_VERSION}" "${VENV}"
    "${UV}" pip install --python "${VENV}/bin/python3" \
        --require-hashes \
        --requirements "${REQ_LOCK_FILE}"
    "${VENV}/bin/ansible-galaxy" collection install \
        --collections-path "${COLLECTIONS_PATH}" \
        --requirements-file "${COLLECTION_REQ_FILE}"
    printf '%s\n' "${PYTHON_VERSION}" >"${PYTHON_STAMP}"
    install -m 0644 "${REQ_FILE}" "${REQ_STAMP}"
    install -m 0644 "${REQ_LOCK_FILE}" "${REQ_LOCK_STAMP}"
    install -m 0644 "${COLLECTION_REQ_FILE}" "${COLLECTION_REQ_STAMP}"
}

"${ROOTDIR}/scripts/fetch_tools.sh" uv

if [ ! -x "${VENV}/bin/ansible-lint" ] || \
    ! [ -f "${PYTHON_STAMP}" ] || \
    ! grep -Fqx "${PYTHON_VERSION}" "${PYTHON_STAMP}" || \
    ! cmp -s "${REQ_FILE}" "${REQ_STAMP}" || \
    ! cmp -s "${REQ_LOCK_FILE}" "${REQ_LOCK_STAMP}" || \
    ! cmp -s "${COLLECTION_REQ_FILE}" "${COLLECTION_REQ_STAMP}"; then
    create_venv
fi

cd "${ROOTDIR}"

exec "${VENV}/bin/ansible-lint" ansible/
