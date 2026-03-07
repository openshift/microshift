#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

GENERATOR=empty-partial-schemas EXTRA_ARGS=--verify ${SCRIPT_ROOT}/hack/update-codegen.sh
GENERATOR=schemapatch EXTRA_ARGS=--verify ${SCRIPT_ROOT}/hack/update-codegen.sh
GENERATOR=crd-manifest-merge EXTRA_ARGS=--verify ${SCRIPT_ROOT}/hack/update-codegen.sh
