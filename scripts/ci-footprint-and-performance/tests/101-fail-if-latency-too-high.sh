#!/bin/bash

set -xeuo pipefail

# For log purposes: dividing maxes by 1000 because oslat with --bucket-width 500 outputs max latency like "4123" instead of "4.123".
echo "Max latency on each core: $(jq -c '[.thread[].max/1000]' "${LOW_LAT_ARTIFACTS}/oslat.json")"

# Not dividing by 1000 and comparing against 10000 (10us) because of floating point number.
max_lat=$(jq '[.thread[].max] | max' "${LOW_LAT_ARTIFACTS}/oslat.json")
if [[ "${max_lat}" -gt 10000 ]]; then
    exit 1
fi
