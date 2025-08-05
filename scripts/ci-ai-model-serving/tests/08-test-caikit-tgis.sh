#!/usr/bin/env bash

set -xeuo pipefail

SCRIPTDIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

function pull_image() {
    local -r img="${1}"
    for i in 1 2 3; do
        GOMAXPROCS=8 sudo crictl pull "${img}" && break
        if [ "${i}" -eq 3 ]; then
            echo "ERROR: Failed to pull ${img} image after 3 attempts"
            exit 1
        fi
        echo "Attempt ${i} failed. Retrying in 5 seconds..." && sleep 5
    done
}

NS=test-caikit-tgis
MINIO_NS=minio

oc create ns "${NS}"

CAIKIT_TGIS_IMAGE="$(jq -r '.images | with_entries(select(.key == "caikit-tgis-image")) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -m)".json)"
TGIS_IMAGE="$(jq -r '.images | with_entries(select(.key == "tgis-image")) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -m)".json)"
pull_image "${CAIKIT_TGIS_IMAGE}"
pull_image "${TGIS_IMAGE}"
pull_image quay.io/opendatahub/modelmesh-minio-examples:caikit-flan-t5

cp /usr/lib/microshift/manifests.d/050-microshift-ai-model-serving-runtimes/caikit-tgis.yaml /tmp/caikit-tgis.yaml
sed -i "s,image: caikit-tgis-image,image: ${CAIKIT_TGIS_IMAGE}," /tmp/caikit-tgis.yaml
sed -i "s,image: tgis-image,image: ${TGIS_IMAGE}," /tmp/caikit-tgis.yaml
oc apply -n "${NS}" -f /tmp/caikit-tgis.yaml

#
# Following instructions are based on https://github.com/opendatahub-io/caikit-tgis-serving/blob/main/demo/kserve/deploy-remove.md
#

# Deploy Minio (self-hostable S3-compatible alternative) with preloaded flan-t5-small
oc create ns "${MINIO_NS}"
oc apply -n "${MINIO_NS}" -f "${SCRIPTDIR}/caikit-tgis/010-minio.yaml"

# Create ServiceAccount and Secret for the InferenceService to use Minio as S3 backend.
oc apply -n "${NS}" -f "${SCRIPTDIR}/caikit-tgis/011-minio-connection-secret.yaml"
oc apply -n "${NS}" -f "${SCRIPTDIR}/caikit-tgis/012-minio-sa.yaml"

# Create InferenceService using model in S3 (minio)
oc apply -n "${NS}" -f "${SCRIPTDIR}/caikit-tgis/020-inference-svc.yaml"

# Create Route for the InferenceService
oc apply -n "${NS}" -f "${SCRIPTDIR}/caikit-tgis/021-route.yaml"

sudo microshift healthcheck \
    -v 2 \
    --timeout 10m0s \
    --namespace "${NS}" \
    --deployments flan-t5-predictor

resp=$(curl -X POST \
    flan-t5-predictor.apps.example.com/api/v1/task/text-generation \
    --connect-to "flan-t5-predictor.apps.example.com::$(hostname -i):" \
    -H "Content-Type: application/json" \
    --data '{"model_id": "flan-t5-small-caikit", "inputs": "At what temperature does Nitrogen boil?"}')

echo "${resp}" | jq

# The answer is technically wrong (boiling point of nitrogen is -320.4°F), but at least we got a response.
# We don't test the model itself, just the integration.
res=0
if ! echo "${resp}" | jq -r '.generated_text' | grep -q "74 degrees F"; then
    echo "Unexpected answer"
    res=1
fi

oc delete ns "${NS}" ; oc delete ns "${MINIO_NS}"

exit "${res}"
