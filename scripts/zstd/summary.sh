#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_SCRIPT="${SCRIPT_DIR}/images.py"

# Default values
SOURCE_JSON=""
TARGET_JSON=""
REPO=""
OUTPUT_FILE="/tmp/output.txt"
VERBOSE=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

usage() {
    cat << EOF
Usage: $0 [OPTIONS] <source_json> <target_json> <repository>

zstd:chunked image pull simulation tool

This tool simulates container image upgrades using zstd:chunked compression
to demonstrate network bandwidth savings and storage efficiency.

ARGUMENTS:
    source_json     JSON file or URL containing source images
    target_json     JSON file or URL containing target images
    repository      Base repository for chunked images (e.g., quay.io/username)

OPTIONS:
    -o, --output FILE       Save detailed output to file
    -v, --verbose          Enable verbose output
    -h, --help             Show this help message

EXAMPLES:
    # Basic simulation
    $0 source.json target.json quay.io/myrepo

    # With output file
    $0 -o simulation.log source.json target.json quay.io/myrepo

    # Using URLs
    $0 https://example.com/source.json target.json quay.io/myrepo

REQUIREMENTS:
    - Podman installed and configured
    - Access to push to the specified repository
    - Storage configuration for zstd:chunked support:
      * enable_partial_images=true
      * use_hard_links=true

The simulation will:
1. Pull source images and push them with zstd:chunked compression
2. Pull target images and push them with zstd:chunked compression
3. Demonstrate reduced download sizes when pulling chunked images
4. Show storage size changes throughout the process

EOF
}

log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&2
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" >&2
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

check_requirements() {
    log "Checking requirements..."

    # Check if podman is installed
    if ! command -v podman &> /dev/null; then
        error "Podman is not installed or not in PATH"
        exit 1
    fi

    # Check if Python script exists
    if [[ ! -f "${PYTHON_SCRIPT}" ]]; then
        error "Python script not found at: ${PYTHON_SCRIPT}"
        exit 1
    fi

    # Check if Python 3 is available
    if ! command -v python3 &> /dev/null; then
        error "Python 3 is not installed or not in PATH"
        exit 1
    fi

    success "All requirements satisfied"
}

check_podman_config() {
    log "Checking Podman storage configuration..."

    local -r storage_conf="${HOME}/.config/containers/storage.conf"
    local issues=()

    if [[ -f "${storage_conf}" ]]; then
        if ! grep -q "enable_partial_images.*=.*true" "${storage_conf}" 2>/dev/null; then
            issues+=("enable_partial_images=true not found")
        fi
        if ! grep -q "use_hard_links.*=.*true" "${storage_conf}" 2>/dev/null; then
            issues+=("use_hard_links=true not found")
        fi
    else
        issues+=("Local storage configuration file not found, using default")
    fi

    if [[ ${#issues[@]} -gt 0 ]]; then
        warn "Podman storage configuration issues detected:"
        for issue in "${issues[@]}"; do
            warn "  - ${issue}"
        done
        warn "For optimal zstd:chunked performance, ensure your storage.conf includes:"
        warn "  enable_partial_images=true"
        warn "  use_hard_links=true"
        echo
    else
        success "Podman storage configuration looks good"
    fi
}

run_images() {
    log "Starting zstd:chunked compression simulation..."
    echo

    local cmd=(script -c "python3 ${PYTHON_SCRIPT} ${SOURCE_JSON} ${TARGET_JSON} ${REPO}" "${OUTPUT_FILE}")

    log "Output will be saved to: ${OUTPUT_FILE}"
    "${cmd[@]}"

    echo

    if [[ -n "${OUTPUT_FILE}" ]]; then
        log "Detailed output saved to: ${OUTPUT_FILE}"
    fi
}

run_summary() {
    log "Analyzing output..."
    echo

    if [[ ! -f "${OUTPUT_FILE}" ]]; then
        warn "Output file not found: ${OUTPUT_FILE}"
        return 1
    fi

    convert_to_bytes() {
        local size="$1"
        local unit="$2"

        case "${unit}" in
            "B")
                echo "${size}"
                ;;
            "KiB")
                echo "$((size * 1024))"
                ;;
            "MiB")
                echo "$((size * 1024 * 1024))"
                ;;
            "GiB")
                echo "$((size * 1024 * 1024 * 1024))"
                ;;
            *)
                echo "0"
                ;;
        esac
    }

    bytes_to_human() {
        local bytes="$1"

        if ((bytes >= 1024*1024*1024)); then
            echo "$((bytes / (1024*1024*1024))) GiB"
        elif ((bytes >= 1024*1024)); then
            echo "$((bytes / (1024*1024))) MiB"
        elif ((bytes >= 1024)); then
            echo "$((bytes / 1024)) KiB"
        else
            echo "${bytes} B"
        fi
    }

    local total_skipped_bytes=0
    local total_original_bytes=0

    log "Parsing zstd:chunked compression results..."

    # shellcheck disable=SC2002
    while IFS= read -r line; do
        if [[ "${line}" =~ Copying\ blob\ [a-f0-9]+\ done[\ \t]+([0-9]+(\.[0-9]+)?)(MiB|KiB|GiB|B)\ \/\ ([0-9]+(\.[0-9]+)?)(MiB|KiB|GiB|B)\ \(skipped:\ ([0-9]+(\.[0-9]+)?)(MiB|KiB|GiB|B)\ =.*\) ]]; then
            local total_size="${BASH_REMATCH[4]}"
            local total_unit="${BASH_REMATCH[6]}"
            local skipped_size="${BASH_REMATCH[7]}"
            local skipped_unit="${BASH_REMATCH[9]}"
            local total_bytes
            total_bytes=$(convert_to_bytes "${total_size%.*}" "${total_unit}")
            local skipped_bytes
            skipped_bytes=$(convert_to_bytes "${skipped_size%.*}" "${skipped_unit}")
            total_skipped_bytes=$((total_skipped_bytes + skipped_bytes))
            total_original_bytes=$((total_original_bytes + total_bytes))
        fi
    done < <(cat /tmp/output.txt | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' | grep '(skipped:' | sort | uniq)

    local savings_percentage=$(((total_skipped_bytes * 100) / total_original_bytes))

    echo
    success "zstd:chunked Compression Summary:"
    echo "  📥 Total original size: $(bytes_to_human ${total_original_bytes})"
    echo "  📦 Total downloaded: $(bytes_to_human $((total_original_bytes - total_skipped_bytes)))"
    echo "  ⏭️  Total skipped: $(bytes_to_human ${total_skipped_bytes}) (${savings_percentage}%)"
    echo
}

parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -o|--output)
                OUTPUT_FILE="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            -*)
                error "Unknown option: $1"
                usage
                exit 1
                ;;
            *)
                if [[ -z "${SOURCE_JSON}" ]]; then
                    SOURCE_JSON="$1"
                elif [[ -z "${TARGET_JSON}" ]]; then
                    TARGET_JSON="$1"
                elif [[ -z "${REPO}" ]]; then
                    REPO="$1"
                else
                    error "Too many arguments"
                    usage
                    exit 1
                fi
                shift
                ;;
        esac
    done

    if [[ -z "${SOURCE_JSON}" || -z "${TARGET_JSON}" || -z "${REPO}" ]]; then
        error "Missing required arguments"
        usage
        exit 1
    fi
}

main() {
    echo -e "${BLUE}========================================${NC}"
    echo -e "${BLUE} zstd:chunked Compression Simulation${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo

    parse_args "$@"

    if [[ "${VERBOSE}" == "true" ]]; then
        set -x
    fi

    check_requirements
    check_podman_config

    echo
    log "Configuration:"
    log "  Source: ${SOURCE_JSON}"
    log "  Target: ${TARGET_JSON}"
    log "  Repository: ${REPO}"
    log "  Output File: ${OUTPUT_FILE}"
    echo

    run_images
    run_summary
}

main "$@"
