#!/bin/bash
#
# This script reports the version of the latest RPM in the local
# repository.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

release_info_rpm=$(find "${LOCAL_REPO}" -name 'microshift-release-info-*.rpm' | sort | tail -n 1)
if [ -z "${release_info_rpm}" ]; then
    error "Failed to find microshift-release-info RPM in ${LOCAL_REPO}"
    exit 1
fi
rpm -q --queryformat '%{version}' "${release_info_rpm}" 2>/dev/null
