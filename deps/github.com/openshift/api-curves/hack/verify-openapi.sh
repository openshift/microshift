#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

OUTPUT_PATH=verify_openapi ${SCRIPT_ROOT}/hack/update-openapi.sh

diff verify_openapi/openapi.json openapi/openapi.json
diff verify_openapi/generated_openapi/zz_generated.openapi.go openapi/generated_openapi/zz_generated.openapi.go

# TODO figure out how to always delete this.
rm -rf verify_openapi
