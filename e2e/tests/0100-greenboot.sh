#!/usr/bin/env bash

# reprovision_after_test=true

set -euo pipefail
IFS=$'\n\t'
PS4='+ $(date "+%T.%N")\011 '
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"

scp "${SCRIPT_PATH}/../../docs/config/busybox_running_check.sh" "${SCRIPT_PATH}/assets/greenboot-test.sh" "${USHIFT_USER}@${USHIFT_IP}":/tmp/
ssh -q "${USHIFT_USER}@${USHIFT_IP}" "chmod +x /tmp/greenboot-test.sh /tmp/busybox_running_check.sh && sudo /tmp/greenboot-test.sh"
