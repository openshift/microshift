#!/bin/bash

set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
CONTROLLER_BIN="${ROOTDIR}/_output/bin/controller-gen"

generate_crd() {
    ${CONTROLLER_BIN} crd paths=./cmd/generate-config/configcrd output:stdout
}

pushd "${ROOTDIR}" &>/dev/null

echo "Generating packaging/microshift/config.yaml and assets/config/config-openapi-spec.json"
generate_crd | go run -mod vendor ./cmd/generate-config \
-a ./assets/config/config-openapi-spec.json \
-o ./packaging/microshift/config.yaml

echo "Updating docs/howto_config.md"
generate_crd | go run -mod vendor ./cmd/generate-config \
-o ./docs/howto_config.md \
-t ./docs/howto_config.md
