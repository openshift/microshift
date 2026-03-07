#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

# OUTPUT_PATH allows the verify script to generate into a different folder.
output_path="${OUTPUT_PATH:-openapi}"
output_package="${SCRIPT_ROOT}/${output_path}"

GENERATOR=openapi EXTRA_ARGS=--openapi:output-package-path=${output_path}/generated_openapi ${SCRIPT_ROOT}/hack/update-codegen.sh

go build github.com/openshift/api/openapi/cmd/models-schema

./models-schema  | jq '.' > ${output_package}/openapi.json
