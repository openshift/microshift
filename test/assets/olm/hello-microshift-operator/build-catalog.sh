#!/usr/bin/env bash
#
# Build a file-based catalog image for the hello-microshift operator.
#
# The catalog contains a single operator (hello-microshift-operator v0.1.0)
# that deploys a minimal HTTP server using quay.io/microshift/busybox:1.36.
#
# Usage:
#   ./build-catalog.sh                                    # build only
#   ./build-catalog.sh --push                             # build and push
#   ./build-catalog.sh --image quay.io/myrepo/catalog:v1  # custom image name

set -euo pipefail

SCRIPT_DIR="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
ROOT_DIR="$(realpath "${SCRIPT_DIR}/../../../..")"

DEFAULT_IMAGE="quay.io/microshift/hello-microshift-catalog:latest"
IMAGE="${DEFAULT_IMAGE}"
PUSH=false

while [[ $# -gt 0 ]]; do
    case "$1" in
        --image)
            IMAGE="$2"
            shift 2
            ;;
        --push)
            PUSH=true
            shift
            ;;
        *)
            echo "Unknown argument: $1" >&2
            exit 1
            ;;
    esac
done

OPM="${OPM:-}"
if [[ -z "${OPM}" ]]; then
    OPM="${ROOT_DIR}/_output/bin/opm"
    if [[ ! -x "${OPM}" ]]; then
        echo "opm not found at ${OPM}, fetching..."
        DEST_DIR="${ROOT_DIR}/_output/bin" "${ROOT_DIR}/scripts/fetch_tools.sh" opm
    fi
fi

WORK_DIR=$(mktemp -d)
trap 'rm -rf "${WORK_DIR}"' EXIT

CATALOG_DIR="${WORK_DIR}/catalog/hello-microshift-operator"
mkdir -p "${CATALOG_DIR}"

echo "Rendering bundle from ${SCRIPT_DIR}..."
"${OPM}" render "${SCRIPT_DIR}" -o json > "${CATALOG_DIR}/catalog.json"

echo "Appending package and channel entries..."
yq -o json -I 0 "${SCRIPT_DIR}/catalog-extra.yaml" >> "${CATALOG_DIR}/catalog.json"

echo "Validating file-based catalog..."
"${OPM}" validate "${WORK_DIR}/catalog/"

cat > "${WORK_DIR}/catalog.Dockerfile" <<'DOCKERFILE'
FROM quay.io/operator-framework/opm:latest
ENTRYPOINT ["/bin/opm"]
CMD ["serve", "/configs", "--cache-dir=/tmp/cache", "--cache-enforce-integrity=false"]
COPY catalog/hello-microshift-operator /configs/hello-microshift-operator
LABEL operators.operatorframework.io.index.configs.v1=/configs
DOCKERFILE

ARCHES=(amd64 arm64)

echo "Creating manifest ${IMAGE}..."
podman manifest rm "${IMAGE}" 2>/dev/null || true
podman manifest create "${IMAGE}"

for arch in "${ARCHES[@]}"; do
    echo "Building linux/${arch}..."
    podman build --platform "linux/${arch}" \
        --manifest "${IMAGE}" \
        -f "${WORK_DIR}/catalog.Dockerfile" "${WORK_DIR}"
done

if ${PUSH}; then
    echo "Pushing ${IMAGE}..."
    podman manifest push --all "${IMAGE}" "docker://${IMAGE}"
fi

echo "Done: ${IMAGE}"
