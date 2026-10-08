#!/bin/bash

set -euo pipefail

# TEMPORARY until OCPBUGS-129494 is resolved. Remove this helper after
# https://redhat.atlassian.net/browse/OCPBUGS-129494 is fixed.
#
# The optional root argument is used by non-privileged fixture tests. The
# helper runs only on a disposable COS 10 test VM in normal use.
root="${1:-/}"

if [[ "${root}" != /* || ! -d "${root}" || -L "${root}" ]]; then
    echo "error: root must be an absolute, non-symlink directory: ${root}" >&2
    exit 1
fi

root="${root%/}"
vendor_dir="${root}/usr/share/containers"
local_dir="${root}/etc/containers"

reject_symlink_components() {
    local -r path="$1"
    local remaining="${path:${#root}}"
    local current="${root}"
    local component

    while [[ -n "${remaining}" ]]; do
        remaining="${remaining#/}"
        [[ -n "${remaining}" ]] || break
        component="${remaining%%/*}"
        current="${current}/${component}"
        if [[ -L "${current}" ]]; then
            echo "error: symlink path component is not allowed: ${current}" >&2
            return 1
        fi
        if [[ "${remaining}" == */* ]]; then
            remaining="${remaining#*/}"
        else
            break
        fi
    done
}

ensure_directory() {
    local -r directory="$1"

    if [[ -e "${directory}" || -L "${directory}" ]]; then
        if [[ ! -d "${directory}" || -L "${directory}" ]]; then
            echo "error: expected a non-symlink directory: ${directory}" >&2
            return 1
        fi
        return
    fi
    mkdir -p "${directory}"
}

copy_if_missing() {
    local -r source="$1"
    local -r destination="$2"

    if [[ -e "${destination}" || -L "${destination}" ]]; then
        if [[ ! -f "${destination}" || -L "${destination}" ]]; then
            echo "error: expected a regular local file: ${destination}" >&2
            return 1
        fi
        echo "Preserving local container configuration: ${destination}"
        return
    fi
    if [[ ! -f "${source}" || -L "${source}" ]]; then
        echo "error: expected a regular vendor file: ${source}" >&2
        return 1
    fi

    ensure_directory "$(dirname "${destination}")"
    install -m "$(stat --format='%a' "${source}")" "${source}" "${destination}"
    echo "Copied missing vendor container configuration: ${destination}"
}

vendor_registries_dir="${vendor_dir}/registries.d"
policy_source="${vendor_dir}/policy.json"
policy_destination="${local_dir}/policy.json"

reject_symlink_components "${policy_source}"
reject_symlink_components "${policy_destination}"
reject_symlink_components "${vendor_registries_dir}"
reject_symlink_components "${local_dir}/registries.d"

if [[ ! -d "${vendor_registries_dir}" || -L "${vendor_registries_dir}" ]]; then
    echo "error: expected a non-symlink vendor directory: ${vendor_registries_dir}" >&2
    exit 1
fi

registry_config_list="$(mktemp)"
trap 'rm -f "${registry_config_list}"' EXIT
if ! find "${vendor_registries_dir}" -maxdepth 1 -type f -print0 | sort -z > "${registry_config_list}"; then
    echo "error: failed to enumerate vendor registry configuration: ${vendor_registries_dir}" >&2
    exit 1
fi

while IFS= read -r -d '' source; do
    reject_symlink_components "${source}"
    reject_symlink_components "${local_dir}/registries.d/$(basename "${source}")"
done < "${registry_config_list}"

copy_if_missing "${policy_source}" "${policy_destination}"

while IFS= read -r -d '' source; do
    copy_if_missing \
        "${source}" \
        "${local_dir}/registries.d/$(basename "${source}")"
done < "${registry_config_list}"
