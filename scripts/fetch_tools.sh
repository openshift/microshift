#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

ARCH="$(uname -p)"

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
DEST_DIR="${DEST_DIR:-${SCRIPT_DIR}/../_output/bin}"
[ -d "${DEST_DIR}" ] || mkdir -p "${DEST_DIR}"
DEST_DIR="$(realpath "${DEST_DIR}")"

_install() {
    local url="$1"
    local checksum="$2"
    local filename="$3"
    local initial_filename="$4"
    local dest="${DEST_DIR}/${filename}"

    [[ -e "${dest}" ]] && return 0
    echo "Installing ${filename} to ${DEST_DIR}"

    tmp=$(mktemp -d)
    trap 'rm -rfv ${tmp} &>/dev/null' EXIT

    filename="$(basename "${url}")"
    echo -n "${checksum} -" >"${tmp}/checksum.txt"

    curl -sSfL --retry 5 --retry-delay 3 -o "${tmp}/${filename}" "${url}"

    if ! sha256sum -c "${tmp}/checksum.txt" < "${tmp}/${filename}" &>/dev/null; then
        echo "  Checksum for ${filename} doesn't match"
        echo "    Expected: ${checksum}"
        echo "         Got: $(sha256sum < "${tmp}/${filename}" | cut -d' ' -f1)"
        return 1
    fi

    # Check type of downloaded file - if it's not executable, then assume it is an archive and needs extracting
    if [[ "$(file --brief --mime-type "${tmp}/${filename}")" != "application/x-executable" ]]; then
        # Extract binary from the archive. 
        # --transform removes any leading dirs leaving just filenames so binary is extracted directly into ${tmp}
        # --wildcards match binary's name so only that file is extracted
        (cd "${tmp}" && tar xvf "${filename}" --transform 's,.*\/,,g' --wildcards "*/${initial_filename}" >/dev/null)
    fi

    chmod +x "${tmp}/${initial_filename}"
    mkdir -p "$(dirname "${dest}")"
    mv "${tmp}/${initial_filename}" "${dest}"
}

get_golangci-lint() {
    local ver="1.52.1"
    declare -A checksums=(
        ["x86_64"]="f31a6dc278aff92843acdc2671f17c753c6e2cb374d573c336479e92daed161f" 
        ["aarch64"]="30dbea4ddde140010981b491740b4dd9ba973ce53a1a2f447a5d57053efe51cf")

    declare -A arch_map=(
        ["x86_64"]="amd64" 
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="golangci-lint"

    local url="https://github.com/golangci/golangci-lint/releases/download/v${ver}/golangci-lint-${ver}-linux-${arch}.tar.gz"

    _install "${url}" "${checksum}" "${filename}" "${filename}"
}

get_shellcheck() {
    local ver="v0.9.0"
    declare -A checksums=(
        ["x86_64"]="700324c6dd0ebea0117591c6cc9d7350d9c7c5c287acbad7630fa17b1d4d9e2f" 
        ["aarch64"]="179c579ef3481317d130adebede74a34dbbc2df961a70916dd4039ebf0735fae")

    declare -A arch_map=(
        ["x86_64"]="x86_64"
        ["aarch64"]="aarch64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="shellcheck"
    local url="https://github.com/koalaman/shellcheck/releases/download/${ver}/shellcheck-${ver}.linux.${arch}.tar.xz"

    _install "${url}" "${checksum}" "${filename}" "${filename}"
}

get_kuttl() {
    local ver="0.15.0"
    declare -A checksums=(
        ["x86_64"]="f6edcf22e238fc71b5aa389ade37a9efce596017c90f6994141c45215ba0f862" 
        ["aarch64"]="a3393f2824e632a9aa0f17fdd5c763f9b633f7a7d3f58696e94885c6b3b8af96")

    declare -A arch_map=(
        ["x86_64"]="x86_64" 
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="kuttl"
    local url="https://github.com/kudobuilder/kuttl/releases/download/v${ver}/kubectl-kuttl_${ver}_linux_${arch}"

    _install "${url}" "${checksum}" "${filename}" "kubectl-kuttl_${ver}_linux_${arch}"
}

get_yq() {
    local ver="4.26.1"
    declare -A checksums=(
        ["x86_64"]="4d3afe5ddf170ac7e70f4c23eea2969eca357947b56d5d96b8516bdf9ce56577" 
        ["aarch64"]="837a659c5a04599f3ee7300b85bf6ccabdfd7ce39f5222de27281e0ea5bcc477")

    declare -A arch_map=(
        ["x86_64"]="amd64" 
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="yq"
    local url="https://github.com/mikefarah/yq/releases/download/v${ver}/yq_linux_${arch}.tar.gz"

    _install "${url}" "${checksum}" "${filename}" "yq_linux_${arch}"
}

tool_getters=$(declare -F |  cut -d' ' -f3 | grep "get_" | sed 's/get_//g')

usage() {
    local msg="${1:-}"
    echo "Script for downloading various tools"
    echo ""
    echo "Usage: $(basename "$0") <all | specific-tool...>"
    echo "Destination can be changed using DEST_DIR environmental variable. Default: '${DEST_DIR}'"
    echo ""
    echo "Tools: "
    # shellcheck disable=SC2001
    echo "${tool_getters}" | sed 's/^/ - /g'

    [ -n "${msg}" ] && echo -e "\nERROR: ${msg}"
    exit 1
}

[[ "$(uname -o)" == "GNU/Linux" ]] || { echo "Script only runs on Linux"; exit 1; }
[[ "${ARCH}" =~ x86_64|aarch64 ]] || { echo "Only x86_64 and aarch64 architectures are supported"; exit 1; }

[ $# -eq 0 ] && usage "Expected at least one argument"

tools_to_install=()
if echo "$@" | grep -q all; then
    readarray -t tools_to_install <<<"${tool_getters}"
else
    for arg in "$@"; do 
        if ! echo "${tool_getters}" | grep -q "${arg}" || [ "$(echo "${tool_getters}" | grep "${arg}")" != "${arg}" ]; then
            usage "Unknown tool: ${arg}"
        fi
        tools_to_install+=("${arg}")
    done
fi

for f in "${tools_to_install[@]}"; do
    "get_${f}"
done
