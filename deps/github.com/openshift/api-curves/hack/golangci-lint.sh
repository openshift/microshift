#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# Build codegen-crds when it's not present and not overridden for a specific file.
if [ -z "${GOLANGCI_LINT:-}" ];then
  ${TOOLS_MAKE} golangci-lint kube-api-linter
  GOLANGCI_LINT="${TOOLS_OUTPUT}/golangci-lint"
fi

# In CI, HOME is set to / and is not writable.
# Make sure golangci-lint can create its cache.
HOME=${HOME:-"/tmp"}
if [[ ${HOME} == "/" ]]; then
  HOME="/tmp"
fi

# We have two separate configs so that we can have different config for types that are go-validated vs openapi validated.
# Init sets errexit so we must unset that for the rest of this script to correctly propagate the exit code.
set +e

"${GOLANGCI_LINT}" $@
status=$?

"${GOLANGCI_LINT}" $@ --config=.golangci.go-validated.yaml && exit ${status}
