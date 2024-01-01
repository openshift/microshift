#!/bin/bash
set -xuo pipefail
#
# Use this script to avoid intermittent DNF errors due to local cache
# inconsistencies or RPM CDN errors.
#
# The following commands can be used to download the script when the
# full MicroShift repository is not available.
#
# DNF_RETRY=$(mktemp /tmp/dnf_retry.XXXXXXXX.sh)
# curl -s https://raw.githubusercontent.com/openshift/microshift/main/scripts/dnf_retry.sh -o "${DNF_RETRY}"
# chmod 755 "${DNF_RETRY}"
#
if [ $# -ne 1 ] && [ $# -ne 2 ] ; then
    echo "Usage: $(basename "$0") <dnf_mode> [packages_to_install]"
    exit 1
fi

DNF_MODE=$1
DNF_PACK=""
[ $# -eq 2 ] && DNF_PACK=$2

rc=0
for _ in $(seq 3) ; do
    # shellcheck disable=SC2086
    sudo dnf "${DNF_MODE}" -y ${DNF_PACK} && exit 0
    # If unsuccessful, save the return code for exit
    rc=$?
    # Clean cache and retry
    sudo dnf clean -y all
done

exit "${rc}"
