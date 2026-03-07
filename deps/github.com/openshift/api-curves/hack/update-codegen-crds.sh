#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

GENERATOR=empty-partial-schemas ${SCRIPT_ROOT}/hack/update-codegen.sh
GENERATOR=schemapatch ${SCRIPT_ROOT}/hack/update-codegen.sh
GENERATOR=crd-manifest-merge ${SCRIPT_ROOT}/hack/update-codegen.sh
