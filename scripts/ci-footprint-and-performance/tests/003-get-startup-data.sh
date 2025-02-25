#!/bin/bash

set -xeuo pipefail

dst_dir="${ARTIFACTS_DIR}/startup-data"
mkdir -p "${dst_dir}"

dst_file="${dst_dir}/${TEST_TIME}.json"

sudo journalctl -u microshift | grep 'Startup data' | grep -oP '\{.*\}' | head -n 1 > "${dst_file}"