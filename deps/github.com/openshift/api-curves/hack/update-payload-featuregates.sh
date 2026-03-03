#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

go run --mod=vendor -trimpath github.com/openshift/api/payload-command/cmd/write-available-featuresets --asset-output-dir=./payload-manifests/featuregates

# Build codegen-crds when it's not present and not overridden for a specific file.
if [ -z "${CODEGEN:-}" ];then
  ${TOOLS_MAKE} codegen
  CODEGEN="${TOOLS_OUTPUT}/codegen"
fi

"${CODEGEN}" featureset-markdown
