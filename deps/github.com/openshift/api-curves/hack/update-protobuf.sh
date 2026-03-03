#!/usr/bin/env bash

source "$(dirname "${BASH_SOURCE}")/lib/init.sh"

[[ -n "${PROTO_OPTIONAL:-}" ]] && exit 0

if [ -z "${GOPATH:-}" ]; then
  echo "Generating protobuf requires GOPATH to be set. Please set GOPATH.
To skip protobuf generation, set \$PROTO_OPTIONAL."
  exit 1
fi

if [[ "${GOPATH}/src/github.com/openshift/api" != "${SCRIPT_ROOT}" ]]; then
  echo "Generating protobuf requires the repository to be checked out within the GOPATH."
  echo "The repository must be checked out at ${GOPATH}/src/github.com/openshift/api."
  exit 1
fi

if [[ "$(protoc --version)" != "libprotoc 23."* ]]; then
  echo "Generating protobuf requires protoc 23.x. Please download and
install the platform appropriate Protobuf package for your OS:

  https://github.com/google/protobuf/releases

To skip protobuf generation, set \$PROTO_OPTIONAL."
  exit 1
fi

# Build go-to-protobuf when it's not present and not overridden for a specific file.
if [ -z "${GO_TO_PROTOBUF:-}" ]; then
  ${TOOLS_MAKE} go-to-protobuf
  GO_TO_PROTOBUF="${TOOLS_OUTPUT}/go-to-protobuf"
fi

# Build protoc-gen-gogo when it's not present and not overridden for a specific file.
if [ -z "${PROTOC_GEN_GOGO:-}" ]; then
  ${TOOLS_MAKE} protoc-gen-gogo
  PROTOC_GEN_GOGO="${TOOLS_OUTPUT}/protoc-gen-gogo"
fi

# We need to make sure that protoc-gen-gogo is on the path for go-to-protobuf to run correctly.
protoc_bin_dir=$(dirname "${PROTOC_GEN_GOGO}")

PATH="$PATH:${protoc_bin_dir}" "${GO_TO_PROTOBUF}" \
  --output-dir="${GOPATH}/src" \
  --apimachinery-packages='-k8s.io/apimachinery/pkg/util/intstr,-k8s.io/apimachinery/pkg/api/resource,-k8s.io/apimachinery/pkg/runtime/schema,-k8s.io/apimachinery/pkg/runtime,-k8s.io/apimachinery/pkg/apis/meta/v1,-k8s.io/apimachinery/pkg/apis/meta/v1beta1,-k8s.io/api/core/v1,-k8s.io/api/rbac/v1' \
  --go-header-file=${SCRIPT_ROOT}/hack/empty.txt \
  --proto-import=${SCRIPT_ROOT}/third_party/protobuf \
  --proto-import=${SCRIPT_ROOT}/vendor \
  --proto-import=${SCRIPT_ROOT}/tools/vendor \
  --packages="${API_PACKAGES}"
