#!/bin/bash
# Build the PCP container and generate a Disk I/O performance graph.
#
# Usage: ./run_io_graph.sh [--timezone TZ] [--output-dir DIR] [pcp-data-dir]
#   --timezone TZ    : IANA timezone for timestamps (default: UTC)
#   --output-dir DIR : directory for output CSV and PNG (default: script directory)
#   pcp-data-dir     : directory containing PCP archive files (default: auto-detect)

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TIMEZONE="UTC"
OUTPUT_DIR=""

# Parse options
while [[ $# -gt 0 ]]; do
    case "$1" in
        --timezone)
            TIMEZONE="${2:?--timezone requires a value (e.g. UTC, US/Eastern)}"
            shift 2
            ;;
        --output-dir)
            OUTPUT_DIR="${2:?--output-dir requires a directory path}"
            shift 2
            ;;
        *)
            DATA_DIR="$1"
            shift
            ;;
    esac
done

OUTPUT_DIR="${OUTPUT_DIR:-${SCRIPT_DIR}}"
mkdir -p "${OUTPUT_DIR}"
OUTPUT_DIR="$(cd "${OUTPUT_DIR}" && pwd)"

DATA_DIR="${DATA_DIR:-$(find "${SCRIPT_DIR}" -name 'Latest' -exec dirname {} \; | head -1)}"

if [[ -z "${DATA_DIR}" || ! -d "${DATA_DIR}" ]]; then
    echo "ERROR: PCP data directory not found. Usage: $0 [--timezone TZ] [--output-dir DIR] <pcp-data-dir>" >&2
    exit 1
fi

DATA_DIR="$(cd "${DATA_DIR}" && pwd)"
IMAGE_NAME="pcp-io-graph"

echo "==> Building container image..."
podman build -t "${IMAGE_NAME}" -f "${SCRIPT_DIR}/Dockerfile" "${SCRIPT_DIR}"

echo "==> Extracting IO data and generating graph from ${DATA_DIR} (timezone: ${TIMEZONE})..."
podman run --rm \
    -v "${DATA_DIR}:/data/pcp:ro,Z" \
    -v "${SCRIPT_DIR}/extract_io.sh:/data/extract_io.sh:ro,Z" \
    -v "${SCRIPT_DIR}/parse_pcp.py:/data/parse_pcp.py:ro,Z" \
    -v "${SCRIPT_DIR}/plot_io.py:/data/plot_io.py:ro,Z" \
    -v "${OUTPUT_DIR}:/data/output:Z" \
    -e MPLCONFIGDIR=/tmp/matplotlib \
    "${IMAGE_NAME}" -c "
        /data/extract_io.sh /data/pcp /data/output/io_data.json ${TIMEZONE} && \
        python3 /data/plot_io.py /data/output/io_data.json -o /data/output/disk_io_performance.png --timezone ${TIMEZONE}
    "

echo "==> Done!"
echo "    JSON:  ${OUTPUT_DIR}/io_data.json"
echo "    Graph: ${OUTPUT_DIR}/disk_io_performance.png"
