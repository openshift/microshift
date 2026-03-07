#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

GENERATOR=swaggerdocs ${SCRIPT_ROOT}/hack/update-codegen.sh
