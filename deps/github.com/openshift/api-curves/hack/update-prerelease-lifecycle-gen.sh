#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# Build prerelease-lifecycle-gen when it's not present and not overridden for a specific file.
if [ -z "${PRERELEASE_LIFECYCLE_GEN:-}" ];then
  ${TOOLS_MAKE} prerelease-lifecycle-gen
  PRERELEASE_LIFECYCLE_GEN="${TOOLS_OUTPUT}/prerelease-lifecycle-gen"
fi

"${PRERELEASE_LIFECYCLE_GEN}" --output-file zz_prerelease_lifecycle_generated.go --logtostderr -v 3 --go-header-file $(dirname "${BASH_SOURCE}")/boilerplate.go.txt ${EXTRA_ARGS:-} $(echo "${API_PACKAGES}" | tr ',' ' ')
