#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

GENERATOR=compatibility EXTRA_ARGS=--verify ${SCRIPT_ROOT}/hack/update-codegen.sh

