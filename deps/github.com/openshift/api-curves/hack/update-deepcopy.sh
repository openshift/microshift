#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

GENERATOR=deepcopy ${SCRIPT_ROOT}/hack/update-codegen.sh
