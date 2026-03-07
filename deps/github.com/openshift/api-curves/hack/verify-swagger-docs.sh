#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

GENERATOR=swaggerdocs EXTRA_ARGS=--verify ${SCRIPT_ROOT}/hack/update-codegen.sh

