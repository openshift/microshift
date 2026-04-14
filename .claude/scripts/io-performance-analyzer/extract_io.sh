#!/bin/bash
# Extract disk IO data from PCP archive using pcp2json with 15-second intervals.
# Outputs JSON with arrays: timestamps, bi (reads/s), bo (writes/s), await (disk await ms)
#
# Usage: ./extract_io.sh <pcp-archive-dir> [output-json] [timezone]

set -euo pipefail

DATA_DIR="${1:?Usage: $0 <pcp-archive-dir> [output-json] [timezone]}"
OUTPUT="${2:-/data/io_data.json}"
TIMEZONE="${3:-UTC}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Using archive directory: ${DATA_DIR}"

# Extract all metrics in a single pcp2json call
TMPFILE=$(mktemp)
trap 'rm -f "${TMPFILE}"' EXIT

(cd "${DATA_DIR}" && pcp2json -a . -t 15sec \
    disk.dev.read disk.dev.write disk.dev.await) \
    > "${TMPFILE}" 2>/dev/null || true

# Parse pcp2json output into clean plot-ready JSON
python3 "${SCRIPT_DIR}/parse_pcp.py" --timezone "${TIMEZONE}" \
    "${TMPFILE}" "${OUTPUT}"

echo "Wrote $(python3 -c "import json; d=json.load(open('${OUTPUT}')); print(len(d['timestamps']))" 2>/dev/null || echo 0) data points to ${OUTPUT}"
