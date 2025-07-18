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

run_summary() {
    log "Analyzing output..."
    echo

    if [[ ! -f "${OUTPUT_FILE}" ]]; then
        warn "Output file not found: ${OUTPUT_FILE}"
        return 1
    fi

    log "Parsing zstd:chunked compression results..."
    echo
    success "zstd:chunked Compression Summary:"
    run_summary_skipped
    run_summary_time
    run_summary_memory
    run_summary_cpu
    run_summary_storage
}

run_summary_skipped() {
    local total_skipped_bytes=0
    local total_original_bytes=0
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
    done < <(cat "${OUTPUT_FILE}" | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' | grep '(skipped:' | sort | uniq)

    local savings_percentage=$(((total_skipped_bytes * 100) / total_original_bytes))

    echo "  📥 Image network usage:"
    echo "     • Non-chunked: $(bytes_to_human ${total_original_bytes})"
    echo "     • Chunked: $(bytes_to_human $((total_original_bytes - total_skipped_bytes)))"
    echo "     • Difference: $(bytes_to_human ${total_skipped_bytes}) (${savings_percentage}% improvement)"
}

run_summary_time() {
    local total_target_time=0
    local total_chunked_target_time=0

    # shellcheck disable=SC2002
    while IFS= read -r line; do
        if [[ "${line}" =~ Successfully\ pulled\ target\ image\.\ ([0-9]+(\.[0-9]+)?)s\ \| ]]; then
            local duration="${BASH_REMATCH[1]}"
            total_target_time=$(echo "${total_target_time} + ${duration}" | bc -l)
        elif [[ "${line}" =~ Successfully\ pulled\ chunked\ target\ image\.\ ([0-9]+(\.[0-9]+)?)s\ \| ]]; then
            local duration="${BASH_REMATCH[1]}"
            total_chunked_target_time=$(echo "${total_chunked_target_time} + ${duration}" | bc -l)
        fi
    done < <(cat "${OUTPUT_FILE}" | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' | grep -E "(Successfully pulled target image|Successfully pulled chunked target image)")

    local -r time_difference=$(echo "scale=2; ${total_target_time} - ${total_chunked_target_time}" | bc -l)
    local -r percentage_improvement=$(echo "scale=2; (${time_difference} * 100) / ${total_target_time}" | bc -l)

    echo "  ⏱️  Image Pull Times:"
    echo "    • Non-chunked: ${total_target_time}s"
    echo "    • Chunked:     ${total_chunked_target_time}s"
    echo "    • Difference:  ${time_difference}s (${percentage_improvement}% improvement)"
}

run_summary_memory() {
    local total_target_memory_avg=0
    local total_target_memory_peak=0
    local total_chunked_target_memory_avg=0
    local total_chunked_target_memory_peak=0
    local target_count=0
    local chunked_target_count=0

    # shellcheck disable=SC2002
    while IFS= read -r line; do
        if [[ "${line}" =~ Successfully\ pulled\ target\ image\.\ [0-9]+(\.[0-9]+)?s\ \|\ ([0-9]+(\.[0-9]+)?)MB\ \|\ ([0-9]+(\.[0-9]+)?)MB\ \| ]]; then
            local memory_avg="${BASH_REMATCH[2]}"
            local memory_peak="${BASH_REMATCH[4]}"
            total_target_memory_avg=$(echo "${total_target_memory_avg} + ${memory_avg}" | bc -l)
            total_target_memory_peak=$(echo "${total_target_memory_peak} + ${memory_peak}" | bc -l)
            target_count=$((target_count + 1))
        elif [[ "${line}" =~ Successfully\ pulled\ chunked\ target\ image\.\ [0-9]+(\.[0-9]+)?s\ \|\ ([0-9]+(\.[0-9]+)?)MB\ \|\ ([0-9]+(\.[0-9]+)?)MB\ \| ]]; then
            local memory_avg="${BASH_REMATCH[2]}"
            local memory_peak="${BASH_REMATCH[4]}"
            total_chunked_target_memory_avg=$(echo "${total_chunked_target_memory_avg} + ${memory_avg}" | bc -l)
            total_chunked_target_memory_peak=$(echo "${total_chunked_target_memory_peak} + ${memory_peak}" | bc -l)
            chunked_target_count=$((chunked_target_count + 1))
        fi
    done < <(cat "${OUTPUT_FILE}" | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' | grep -E "(Successfully pulled target image|Successfully pulled chunked target image)")

    if [[ ${target_count} -gt 0 && ${chunked_target_count} -gt 0 ]]; then
        local -r avg_target_memory_avg=$(echo "scale=1; ${total_target_memory_avg} / ${target_count}" | bc -l)
        local -r avg_target_memory_peak=$(echo "scale=1; ${total_target_memory_peak} / ${target_count}" | bc -l)
        local -r avg_chunked_target_memory_avg=$(echo "scale=1; ${total_chunked_target_memory_avg} / ${chunked_target_count}" | bc -l)
        local -r avg_chunked_target_memory_peak=$(echo "scale=1; ${total_chunked_target_memory_peak} / ${chunked_target_count}" | bc -l)

        local -r memory_avg_difference=$(echo "scale=1; ${avg_target_memory_avg} - ${avg_chunked_target_memory_avg}" | bc -l)
        local -r memory_peak_difference=$(echo "scale=1; ${avg_target_memory_peak} - ${avg_chunked_target_memory_peak}" | bc -l)
        local -r avg_percentage_improvement=$(echo "scale=2; (${memory_avg_difference} * 100) / ${avg_target_memory_avg}" | bc -l)
        local -r peak_percentage_improvement=$(echo "scale=2; (${memory_peak_difference} * 100) / ${avg_target_memory_peak}" | bc -l)

        echo "  🧠 Image Memory Usage:"
        echo "    • Non-chunked avg: ${avg_target_memory_avg}MB | peak: ${avg_target_memory_peak}MB"
        echo "    • Chunked avg:     ${avg_chunked_target_memory_avg}MB | peak: ${avg_chunked_target_memory_peak}MB"
        echo "    • Difference avg:  ${memory_avg_difference}MB (${avg_percentage_improvement}% improvement)"
        echo "    • Difference peak: ${memory_peak_difference}MB (${peak_percentage_improvement}% improvement)"
    fi
}

run_summary_cpu() {
    local total_target_cpu_avg=0
    local total_target_cpu_peak=0
    local total_chunked_target_cpu_avg=0
    local total_chunked_target_cpu_peak=0
    local target_count=0
    local chunked_target_count=0

    # shellcheck disable=SC2002
    while IFS= read -r line; do
        if [[ "${line}" =~ Successfully\ pulled\ target\ image\.\ [0-9]+(\.[0-9]+)?s\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ ([0-9]+(\.[0-9]+)?)%\ \|\ ([0-9]+(\.[0-9]+)?)%\ \| ]]; then
            local cpu_avg="${BASH_REMATCH[4]}"
            local cpu_peak="${BASH_REMATCH[6]}"
            total_target_cpu_avg=$(echo "${total_target_cpu_avg} + ${cpu_avg}" | bc -l)
            total_target_cpu_peak=$(echo "${total_target_cpu_peak} + ${cpu_peak}" | bc -l)
            target_count=$((target_count + 1))
        elif [[ "${line}" =~ Successfully\ pulled\ chunked\ target\ image\.\ [0-9]+(\.[0-9]+)?s\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ ([0-9]+(\.[0-9]+)?)%\ \|\ ([0-9]+(\.[0-9]+)?)%\ \| ]]; then
            local cpu_avg="${BASH_REMATCH[4]}"
            local cpu_peak="${BASH_REMATCH[6]}"
            total_chunked_target_cpu_avg=$(echo "${total_chunked_target_cpu_avg} + ${cpu_avg}" | bc -l)
            total_chunked_target_cpu_peak=$(echo "${total_chunked_target_cpu_peak} + ${cpu_peak}" | bc -l)
            chunked_target_count=$((chunked_target_count + 1))
        fi
    done < <(cat "${OUTPUT_FILE}" | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' | grep -E "(Successfully pulled target image|Successfully pulled chunked target image)")

    if [[ ${target_count} -gt 0 && ${chunked_target_count} -gt 0 ]]; then
        local -r avg_target_cpu_avg=$(echo "scale=1; ${total_target_cpu_avg} / ${target_count}" | bc -l)
        local -r avg_target_cpu_peak=$(echo "scale=1; ${total_target_cpu_peak} / ${target_count}" | bc -l)
        local -r avg_chunked_target_cpu_avg=$(echo "scale=1; ${total_chunked_target_cpu_avg} / ${chunked_target_count}" | bc -l)
        local -r avg_chunked_target_cpu_peak=$(echo "scale=1; ${total_chunked_target_cpu_peak} / ${chunked_target_count}" | bc -l)

        local -r cpu_avg_difference=$(echo "scale=1; ${avg_target_cpu_avg} - ${avg_chunked_target_cpu_avg}" | bc -l)
        local -r cpu_peak_difference=$(echo "scale=1; ${avg_target_cpu_peak} - ${avg_chunked_target_cpu_peak}" | bc -l)
        local -r avg_percentage_improvement=$(echo "scale=2; (${cpu_avg_difference} * 100) / ${avg_target_cpu_avg}" | bc -l)
        local -r peak_percentage_improvement=$(echo "scale=2; (${cpu_peak_difference} * 100) / ${avg_target_cpu_peak}" | bc -l)

        echo "  🔥 Image CPU Usage:"
        echo "    • Non-chunked avg: ${avg_target_cpu_avg}% | peak: ${avg_target_cpu_peak}%"
        echo "    • Chunked avg:     ${avg_chunked_target_cpu_avg}% | peak: ${avg_chunked_target_cpu_peak}%"
        echo "    • Difference avg:  ${cpu_avg_difference}% (${avg_percentage_improvement}% improvement)"
        echo "    • Difference peak: ${cpu_peak_difference}% (${peak_percentage_improvement}% improvement)"
    fi
}

run_summary_storage() {
    local total_target_storage=0
    local total_chunked_target_storage=0
    local total_target_read_iops=0
    local total_target_write_iops=0
    local total_target_peak_iops=0
    local total_chunked_target_read_iops=0
    local total_chunked_target_write_iops=0
    local total_chunked_target_peak_iops=0
    local target_count=0
    local chunked_target_count=0

    # shellcheck disable=SC2002
    while IFS= read -r line; do
        if [[ "${line}" =~ Successfully\ pulled\ target\ image\.\ [0-9]+(\.[0-9]+)?s\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ [0-9]+(\.[0-9]+)?%\ \|\ [0-9]+(\.[0-9]+)?%\ \|\ ([0-9]+(\.[0-9]+)?)\ MB\ \|\ R:([0-9]+(\.[0-9]+)?)\ W:([0-9]+(\.[0-9]+)?)\ Peak:([0-9]+(\.[0-9]+)?)\ IOPS ]]; then
            local storage="${BASH_REMATCH[6]}"
            local read_iops="${BASH_REMATCH[8]}"
            local write_iops="${BASH_REMATCH[10]}"
            local peak_iops="${BASH_REMATCH[12]}"
            total_target_storage=$(echo "${total_target_storage} + ${storage}" | bc -l)
            total_target_read_iops=$(echo "${total_target_read_iops} + ${read_iops}" | bc -l)
            total_target_write_iops=$(echo "${total_target_write_iops} + ${write_iops}" | bc -l)
            total_target_peak_iops=$(echo "${total_target_peak_iops} + ${peak_iops}" | bc -l)
            target_count=$((target_count + 1))
        elif [[ "${line}" =~ Successfully\ pulled\ chunked\ target\ image\.\ [0-9]+(\.[0-9]+)?s\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ [0-9]+(\.[0-9]+)?MB\ \|\ [0-9]+(\.[0-9]+)?%\ \|\ [0-9]+(\.[0-9]+)?%\ \|\ ([0-9]+(\.[0-9]+)?)\ MB\ \|\ R:([0-9]+(\.[0-9]+)?)\ W:([0-9]+(\.[0-9]+)?)\ Peak:([0-9]+(\.[0-9]+)?)\ IOPS ]]; then
            local storage="${BASH_REMATCH[6]}"
            local read_iops="${BASH_REMATCH[8]}"
            local write_iops="${BASH_REMATCH[10]}"
            local peak_iops="${BASH_REMATCH[12]}"
            total_chunked_target_storage=$(echo "${total_chunked_target_storage} + ${storage}" | bc -l)
            total_chunked_target_read_iops=$(echo "${total_chunked_target_read_iops} + ${read_iops}" | bc -l)
            total_chunked_target_write_iops=$(echo "${total_chunked_target_write_iops} + ${write_iops}" | bc -l)
            total_chunked_target_peak_iops=$(echo "${total_chunked_target_peak_iops} + ${peak_iops}" | bc -l)
            chunked_target_count=$((chunked_target_count + 1))
        fi
    done < <(cat "${OUTPUT_FILE}" | sed 's/\x1b\[[0-9;]*[a-zA-Z]//g' | grep -E "(Successfully pulled target image|Successfully pulled chunked target image)")

    local -r storage_difference=$(echo "scale=2; ${total_target_storage} - ${total_chunked_target_storage}" | bc -l)
    local -r percentage_improvement=$(echo "scale=2; (${storage_difference} * 100) / ${total_target_storage}" | bc -l)

    echo "  💾 Image Storage Usage:"
    echo "    • Non-chunked: ${total_target_storage}MB"
    echo "    • Chunked:     ${total_chunked_target_storage}MB"
    echo "    • Difference:  ${storage_difference}MB (${percentage_improvement}% improvement)"

    if [[ ${target_count} -gt 0 && ${chunked_target_count} -gt 0 ]]; then
        local -r avg_target_read_iops=$(echo "scale=1; ${total_target_read_iops} / ${target_count}" | bc -l)
        local -r avg_target_write_iops=$(echo "scale=1; ${total_target_write_iops} / ${target_count}" | bc -l)
        local -r avg_target_peak_iops=$(echo "scale=1; ${total_target_peak_iops} / ${target_count}" | bc -l)
        local -r avg_chunked_target_read_iops=$(echo "scale=1; ${total_chunked_target_read_iops} / ${chunked_target_count}" | bc -l)
        local -r avg_chunked_target_write_iops=$(echo "scale=1; ${total_chunked_target_write_iops} / ${chunked_target_count}" | bc -l)
        local -r avg_chunked_target_peak_iops=$(echo "scale=1; ${total_chunked_target_peak_iops} / ${chunked_target_count}" | bc -l)

        local -r read_iops_difference=$(echo "scale=1; ${avg_target_read_iops} - ${avg_chunked_target_read_iops}" | bc -l)
        local -r write_iops_difference=$(echo "scale=1; ${avg_target_write_iops} - ${avg_chunked_target_write_iops}" | bc -l)
        local -r peak_iops_difference=$(echo "scale=1; ${avg_target_peak_iops} - ${avg_chunked_target_peak_iops}" | bc -l)
        local -r read_percentage_improvement=$(echo "scale=2; (${read_iops_difference} * 100) / ${avg_target_read_iops}" | bc -l)
        local -r write_percentage_improvement=$(echo "scale=2; (${write_iops_difference} * 100) / ${avg_target_write_iops}" | bc -l)
        local -r peak_percentage_improvement=$(echo "scale=2; (${peak_iops_difference} * 100) / ${avg_target_peak_iops}" | bc -l)

        echo "  📊 Image IOPS Usage:"
        echo "    • Non-chunked R:${avg_target_read_iops} W:${avg_target_write_iops} Peak:${avg_target_peak_iops} IOPS"
        echo "    • Chunked     R:${avg_chunked_target_read_iops} W:${avg_chunked_target_write_iops} Peak:${avg_chunked_target_peak_iops} IOPS"
        echo "    • Difference  R:${read_iops_difference} (${read_percentage_improvement}%) W:${write_iops_difference} (${write_percentage_improvement}%) Peak:${peak_iops_difference} (${peak_percentage_improvement}%) IOPS"
    fi
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
