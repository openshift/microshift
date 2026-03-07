#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

VERIFY_DIR=$(mktemp -d -t featuregates-verify-XXXXXX)

go run --mod=vendor -trimpath github.com/openshift/api/payload-command/cmd/write-available-featuresets --asset-output-dir="${VERIFY_DIR}"

diff -r "${VERIFY_DIR}" ./payload-manifests/featuregates

rm -rf "${VERIFY_DIR}"

# Build codegen-crds when it's not present and not overridden for a specific file.
if [ -z "${CODEGEN:-}" ];then
  ${TOOLS_MAKE} codegen
  CODEGEN="${TOOLS_OUTPUT}/codegen"
fi

"${CODEGEN}" featureset-markdown --verify
