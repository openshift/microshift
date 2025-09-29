#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

ARCH="$(uname -m)"

# Set GitHub token for private repository access
GITHUB_TOKEN="$(cat /tmp/token-git 2>/dev/null || echo '')"
export GITHUB_TOKEN

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR=$(realpath "${SCRIPT_DIR}/..")
DEFAULT_DEST_DIR="${ROOT_DIR}/_output/bin"
DEST_DIR="${DEST_DIR:-${DEFAULT_DEST_DIR}}"
[ -d "${DEST_DIR}" ] || mkdir -p "${DEST_DIR}"
DEST_DIR="$(realpath "${DEST_DIR}")"
WORK_DIR=$(mktemp -d)
trap 'rm -rfv ${WORK_DIR} &>/dev/null' EXIT

_install() {
    local url="$1"
    local checksum="$2"
    local filename="$3"
    local initial_filename="$4"
    local dest="${DEST_DIR}/${filename}"

    [[ -e "${dest}" ]] && return 0
    echo "Installing ${filename} to ${DEST_DIR}"

    filename="$(basename "${url}")"
    echo -n "${checksum} -" >"${WORK_DIR}/checksum.txt"

    curl -sSfL --retry 5 --retry-delay 3 -o "${WORK_DIR}/${filename}" "${url}"

    if ! sha256sum -c "${WORK_DIR}/checksum.txt" < "${WORK_DIR}/${filename}" &>/dev/null; then
        echo "  Checksum for ${filename} doesn't match"
        echo "    Expected: ${checksum}"
        echo "         Got: $(sha256sum < "${WORK_DIR}/${filename}" | cut -d' ' -f1)"
        return 1
    fi

    # Check type of downloaded file - if it's not executable, then assume it is an archive and needs extracting
    if [[ "$(file --brief --mime-type "${WORK_DIR}/${filename}")" != "application/x-executable" ]]; then
        # Extract binary from the archive.
        # --transform removes any leading dirs leaving just filenames so binary is extracted directly into ${WORK_DIR}
        # --wildcards match binary's name so only that file is extracted
        (cd "${WORK_DIR}" && tar xvf "${filename}" --transform 's,.*\/,,g' --wildcards "*/${initial_filename}" >/dev/null)
    fi

    chmod +x "${WORK_DIR}/${initial_filename}"
    mkdir -p "$(dirname "${dest}")"
    mv "${WORK_DIR}/${initial_filename}" "${dest}"
}

gettool_golangci-lint() {
    local ver="2.1.6"
    declare -A checksums=(
        ["x86_64"]="e55e0eb515936c0fbd178bce504798a9bd2f0b191e5e357768b18fd5415ee541"
        ["aarch64"]="582eb73880f4408d7fb89f12b502d577bd7b0b63d8c681da92bb6b9d934d4363")

    declare -A arch_map=(
        ["x86_64"]="amd64"
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="golangci-lint"

    local url="https://github.com/golangci/golangci-lint/releases/download/v${ver}/golangci-lint-${ver}-linux-${arch}.tar.gz"

    _install "${url}" "${checksum}" "${filename}" "${filename}"
}

gettool_shellcheck() {
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

gettool_kuttl() {
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

gettool_yq() {
    local ver="4.44.2"
    declare -A checksums=(
        ["x86_64"]="e4c2570249e3993e33ffa44e592b5eee8545bd807bfbeb596c2986d86cb6c85c"
        ["aarch64"]="79c22d98b2ff517cb8b1c20499350cbc1e8c753483c8f72a37a299e6e9872a98")

    declare -A arch_map=(
        ["x86_64"]="amd64"
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="yq"
    local url="https://github.com/mikefarah/yq/releases/download/v${ver}/yq_linux_${arch}.tar.gz"

    _install "${url}" "${checksum}" "${filename}" "yq_linux_${arch}"
}

gettool_lichen() {
    local ver="v0.1.7"
    GOBIN=${DEST_DIR} GOFLAGS="" go install github.com/uw-labs/lichen@${ver}
}

gettool_govulncheck() {
    # Must use latest to get up-to-date vulnerability checks
    local ver="latest"
    GOBIN=${DEST_DIR} GOFLAGS="" go install -mod=mod golang.org/x/vuln/cmd/govulncheck@${ver}
}

gettool_controller-gen() {
    local ver="v0.18.0"
    GOBIN=${DEST_DIR} GOFLAGS="" go install sigs.k8s.io/controller-tools/cmd/controller-gen@${ver}
}

gettool_gomplate() {
    local ver="v3.11.5"
    declare -A checksums=(
        ["x86_64"]="16f6a01a0ff22cae1302980c42ce4f98ca20f8c55443ce5a8e62e37fc23487b3"
        ["aarch64"]="fd980f9d233902e50f3f03f10ea65f36a2705385358a87aa18b19fb7cdf54c1d")

    declare -A arch_map=(
        ["x86_64"]="amd64"
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"
    local checksum="${checksums[${ARCH}]}"
    local filename="gomplate"
    local url="https://github.com/hairyhenderson/gomplate/releases/download/${ver}/gomplate_linux-${arch}"

    _install "${url}" "${checksum}" "${filename}" "gomplate_linux-${arch}"
}

gettool_robotframework() {
    local venv

    if [ "${DEST_DIR}" = "${DEFAULT_DEST_DIR}" ]; then
        # Probably running as the user, not in CI.
        venv="${ROOT_DIR}/_output/robotenv"
    else
        # Probably running in automation environment where the output
        # location has been changed.
        venv="${DEST_DIR}"
    fi

    if [ ! -f "${venv}/bin/robot" ]; then
        python3 -m venv "${venv}"
        "${venv}/bin/python3" -m pip install --upgrade pip
        "${venv}/bin/python3" -m pip install -r "${ROOT_DIR}/test/requirements.txt"
    fi
}

gettool_awscli() {
    # Download AWS CLI
    pushd "${WORK_DIR}" &>/dev/null

    curl -s "https://awscli.amazonaws.com/awscli-exe-linux-$(uname -m).zip" -o "awscliv2.zip" && \
        unzip -q awscliv2.zip
    ./aws/install --update --install-dir "$(realpath "${DEST_DIR}/../awscli")" --bin-dir "${DEST_DIR}"

    popd &>/dev/null
}

gettool_oc() {
    declare -A arch_map=(
        ["x86_64"]="x86_64"
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"

    pushd "${WORK_DIR}" &>/dev/null

    curl -s -f "https://mirror.openshift.com/pub/openshift-v4/${arch}/clients/ocp/latest/openshift-client-linux.tar.gz" -L -o "openshift-client-linux.tar.gz"
    tar xvzf openshift-client-linux.tar.gz
    sudo cp oc /usr/bin/oc
    sudo cp kubectl /usr/bin/kubectl

    popd &>/dev/null
}

gettool_opm() {
    declare -A arch_map=(
        ["x86_64"]="x86_64"
        ["aarch64"]="arm64")

    local arch="${arch_map[${ARCH}]}"

    pushd "${WORK_DIR}" &>/dev/null

    curl -s -f "https://mirror.openshift.com/pub/openshift-v4/${arch}/clients/ocp/latest/opm-linux-rhel9.tar.gz" -L -o "opm-linux-rhel9.tar.gz"
    tar xvzf opm-linux-rhel9.tar.gz 
    mv opm-rhel9 "${DEST_DIR}"/opm

    popd &>/dev/null
}

gettool_brew() {
    # See https://spaces.redhat.com/display/Brew/Using+the+Brew+Prod+environment#UsingtheBrewProdenvironment-Fedora
    if ! command -v koji &>/dev/null ; then
        sudo dnf install -y koji
    fi

    if ! command -v brew &>/dev/null ; then
        sudo ln -sv /usr/bin/koji /usr/bin/brew
    fi

    if [ ! -f ~/.koji/config ] ; then
        mkdir -p ~/.koji
        cat > ~/.koji/config <<EOF
[brew]
server = https://brewhub.engineering.redhat.com/brewhub
topdir = /mnt/redhat/brewroot
weburl = https://brewweb.engineering.redhat.com
topurl = http://download.devel.redhat.com/brewroot
EOF
    fi
}

gettool_tar-diff() {
    # See https://github.com/containers/tar-diff
    local ver="v0.1.2"
    for pkg in tar-diff tar-patch ; do
        GOBIN=${DEST_DIR} GOFLAGS="" go install -mod=mod github.com/containers/tar-diff/cmd/${pkg}@${ver}
    done
}

gettool_ginkgo() {
    # shellcheck source=test/bin/common.sh
    source "${SCRIPT_DIR}/../test/bin/common.sh"
    # shellcheck source=test/bin/common_versions.sh
    source "${SCRIPT_DIR}/../test/bin/common_versions.sh"

    local -r repo_url="https://${GITHUB_TOKEN}@github.com/openshift/openshift-tests-private.git"
    local -r binary_path="${GINKGO_TEST_BINARY}"
    local -r release_branch="release-4.${MINOR_VERSION}"

    # Skip if binary already exists
    [[ -f "${binary_path}" ]] && return 0

    # Check if Go is available
    if ! command -v go &> /dev/null; then
        echo "Error: Go is required to build openshift-tests-private binary"
        return 1
    fi

    # Ensure binary directory exists
    mkdir -p "$(dirname "${binary_path}")"

    # Use global WORK_DIR for cloning
    local clone_dir="${WORK_DIR}/openshift-tests-private"

    # Clone repository with release branch preference
    if ! git clone --depth 1 --branch "${release_branch}" "${repo_url}" "${clone_dir}" 2>/dev/null; then
        echo "Branch ${release_branch} not found, cloning default branch..."
        if ! git clone --depth 1 "${repo_url}" "${clone_dir}"; then
            echo "Error: Failed to clone repository"
            return 1
        fi
    fi

    # Build the binary
    pushd "${clone_dir}" &>/dev/null

    local test_binary="./bin/extended-platform-tests"
    make build

    # Copy binary to centralized location
    if [[ -f "${test_binary}" ]]; then
        cp "${test_binary}" "${binary_path}"
        chmod +x "${binary_path}"
        echo "Binary installed to ${binary_path}"

	# Copy handleresult.py to the tools directory
        if [[ -f "${clone_dir}/pipeline/handleresult.py" ]] && [[ -f "${HANDLERESULT_SCRIPT}" ]]; then
            cp "${clone_dir}/pipeline/handleresult.py" "${HANDLERESULT_SCRIPT}"
            chmod +x "${HANDLERESULT_SCRIPT}"
            echo "handleresult.py installed to ${HANDLERESULT_SCRIPT}"
        else
            echo "Warning: pipeline/handleresult.py not found in repository"
        fi

        popd &>/dev/null
    else
        echo "Error: Test binary not found after build"
        popd &>/dev/null
        return 1
    fi
}

tool_getters=$(declare -F | awk '$3 ~ /^gettool_/ {print $3}' | sed 's/^gettool_//g')

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
if grep -qw all <<<"$@"; then
    readarray -t tools_to_install <<<"${tool_getters}"
else
    for arg in "$@"; do
        if ! grep -wq "${arg}" <<<"${tool_getters}" ; then
            usage "Unknown tool: \"${arg}\""
        fi
        tools_to_install+=("${arg}")
    done
fi

for f in "${tools_to_install[@]}"; do
    "gettool_${f}"
done
