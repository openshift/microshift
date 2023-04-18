#!/bin/bash

set -euo pipefail

ROOTDIR=$(git rev-parse --show-toplevel)
CONTROLLER_BIN="${ROOTDIR}/_output/bin/controller-gen"

generate_crd() {
    ${CONTROLLER_BIN} crd paths=../../hack/config-gen/configcrd output:stdout
}

echo "Generating packaging/microshift/config.yaml and cockpit-plugin/packaging/config-openapi-spec.json"
generate_crd | go run -mod vendor ../../hack/config-gen \
-a ../../cockpit-plugin/packaging/config-openapi-spec.json \
-o ../../packaging/microshift/config.yaml

echo "Updating docs/howto_config.md"
generate_crd | go run -mod vendor ../../hack/config-gen \
-o ../../docs/howto_config.md \
-t ../../docs/howto_config.md
