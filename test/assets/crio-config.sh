#!/bin/bash

set -euo pipefail

# TEMPORARY compatibility for COS 10. Remove this helper when OCPBUGS-129494 is
# resolved and the image exposes its vendor signature configuration where CRI-O
# expects it:
# https://redhat.atlassian.net/browse/OCPBUGS-129494
readonly vendor_dir=/usr/share/containers
readonly local_dir=/etc/containers

if (( $# != 0 )); then
    echo "error: crio-config.sh does not accept arguments" >&2
    exit 1
fi

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

validate_source_file() {
    local -r source="$1"

    if [[ ! -f "${source}" || -L "${source}" || ! -r "${source}" ]]; then
        echo "error: expected a readable regular vendor file: ${source}" >&2
        return 1
    fi
}

validate_destination_file() {
    local -r destination="$1"

    if [[ -e "${destination}" || -L "${destination}" ]] &&
        [[ ! -f "${destination}" || -L "${destination}" ]]; then
        echo "error: expected a regular local file: ${destination}" >&2
        return 1
    fi
}

validate_vendor_directory() {
    local -r directory="$1"

    if [[ ! -d "${directory}" || -L "${directory}" ||
        ! -r "${directory}" || ! -x "${directory}" ]]; then
        echo "error: expected a readable, searchable, non-symlink vendor directory: ${directory}" >&2
        return 1
    fi
}

validate_local_directory() {
    local -r directory="$1"

    if [[ ! -d "${directory}" || -L "${directory}" ||
        ! -w "${directory}" || ! -x "${directory}" ]]; then
        echo "error: expected a writable, searchable, non-symlink local directory: ${directory}" >&2
        return 1
    fi
}

copy_if_missing() {
    local -r source="$1"
    local -r destination="$2"

    validate_destination_file "${destination}"
    if [[ -e "${destination}" || -L "${destination}" ]]; then
        echo "Preserving local container configuration: ${destination}"
        return
    fi
    validate_source_file "${source}"

    install -m "$(stat --format='%a' "${source}")" "${source}" "${destination}"
    echo "Copied missing vendor container configuration: ${destination}"
}

readonly vendor_registries_dir="${vendor_dir}/registries.d"
readonly local_registries_dir="${local_dir}/registries.d"
readonly policy_source="${vendor_dir}/policy.json"
readonly policy_destination="${local_dir}/policy.json"
readonly local_parent_dir="${local_dir%/*}"

# Preflight every source and destination before creating or installing anything.
validate_vendor_directory "${vendor_dir}"
validate_vendor_directory "${vendor_registries_dir}"
validate_source_file "${policy_source}"
validate_destination_file "${policy_destination}"

validate_local_directory "${local_parent_dir}"
if [[ -e "${local_dir}" || -L "${local_dir}" ]]; then
    validate_local_directory "${local_dir}"
fi
if [[ -e "${local_registries_dir}" || -L "${local_registries_dir}" ]]; then
    validate_local_directory "${local_registries_dir}"
fi

registry_configs=()
shopt -s lastpipe
if ! find "${vendor_registries_dir}" -mindepth 1 -maxdepth 1 -type f -print0 |
    mapfile -d '' -t registry_configs; then
    echo "error: failed to enumerate vendor registry configuration: ${vendor_registries_dir}" >&2
    exit 1
fi

for source in "${registry_configs[@]}"; do
    validate_source_file "${source}"
    validate_destination_file "${local_registries_dir}/${source##*/}"
done

ensure_directory "${local_dir}"
ensure_directory "${local_registries_dir}"

copy_if_missing "${policy_source}" "${policy_destination}"
for source in "${registry_configs[@]}"; do
    copy_if_missing "${source}" "${local_registries_dir}/${source##*/}"
done
