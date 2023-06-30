#!/bin/bash

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
RPMBUILD_DIR="$(git rev-parse --show-toplevel)/_output/rpmbuild/"

mkdir -p "${RPMBUILD_DIR}"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
mkdir -p "${RPMBUILD_DIR}/SOURCES/test-agent"
cp "${SCRIPT_DIR}"/microshift-test-agent.{service,sh} "${RPMBUILD_DIR}/SOURCES/test-agent"

rpmbuild --define "_topdir ${RPMBUILD_DIR}" -bb "${SCRIPT_DIR}/microshift-test-agent.spec"
