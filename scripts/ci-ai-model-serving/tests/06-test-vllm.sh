#!/usr/bin/env bash

set -xeuo pipefail

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

# Create namespace
oc create ns test-vllm

# Get the vLLM image from the release info to: 
# 1. pre-fetch it
# 2. update the vllm ServingRuntime CR
VLLM_IMAGE="$(jq -r '.images | with_entries(select(.key == "vllm-image")) | .[]' /usr/share/microshift/release/release-ai-model-serving-"$(uname -i)".json)"

# Pull the images before creating K8s objects to make the test more
# deterministic (both images are several GiB in size, so they take a while to download).
pull_image "${VLLM_IMAGE}"
pull_image quay.io/microshift/ai-testing-model:vllm-granite-3b-code-base-2k

# Create ServingRuntime
cp /usr/lib/microshift/manifests.d/001-microshift-ai-model-serving/runtimes/vllm.yaml /tmp/vllm.yaml
sed -i "s,image: vllm-image,image: ${VLLM_IMAGE}," /tmp/vllm.yaml
oc apply -n test-vllm -f /tmp/vllm.yaml

# Create InferenceService
# --dtype=half will be passed through to the deployment and to the vLLM model server.
# It's needed because GPU on g4dn.xlarge isn't powerful enough for Bfloat16 but it's enough for this test.
cat <<'EOF' | oc apply -n test-vllm -f -
apiVersion: serving.kserve.io/v1beta1
kind: InferenceService
metadata:
  name: granite-3b-code-base-2k
spec:
  predictor:
    model:
      modelFormat:
        name: vLLM
      storageUri: "oci://quay.io/microshift/ai-testing-model:vllm-granite-3b-code-base-2k"
      args:
      - --dtype=half
      resources:
        limits:
          nvidia.com/gpu: 1
        requests:
          nvidia.com/gpu: 1
EOF

# Create Route for the InferenceService
cat <<'EOF' | oc apply -n test-vllm -f -
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: granite
spec:
  host: granite-predictor.apps.example.com
  port:
    targetPort: 8080
  to:
    kind: Service
    name: granite-3b-code-base-2k-predictor
    weight: 100
  wildcardPolicy: None
EOF

# Wait for the Deployment that kserve created based on the InferenceService created earlier.
# 10m of timeout because it might take a while to unpack the model from the OCI.
sudo microshift healthcheck \
    -v 2 \
    --timeout 10m0s \
    --namespace test-vllm \
    --deployments granite-3b-code-base-2k-predictor

# Ask the model "At what temperature does water boil?"
# Thanks to temperature=0 the answer is constant.
resp=$(curl -X POST \
    granite-predictor.apps.example.com/v1/completions \
    --connect-to "granite-predictor.apps.example.com::$(hostname -i):" \
    -H "Content-Type: application/json" \
    --data '{
        "model": "granite-3b-code-base-2k",
        "prompt": "At what temperature does water boil?",
        "max_tokens": 16,
        "temperature": 0}')

res=0
if ! echo "${resp}" | jq -r '.choices[0].text' | grep -q "The boiling point of water is 100°C."; then
    echo "Unexpected answer"
    res=1
fi

# Query the model again - just for fun and double checking.
# Because of temperature=0.5, the responses vary between calls.
resp=$(curl -X POST \
    granite-predictor.apps.example.com/v1/completions \
    --connect-to "granite-predictor.apps.example.com::$(hostname -i):" \
    -H "Content-Type: application/json" \
    --data '{
        "model": "granite-3b-code-base-2k",
        "prompt": "Once upon a time,",
        "max_tokens": 256,
        "temperature": 0.5}')

text=$(echo "${resp}" | jq -r '.choices[0].text')
if [[ "${text}" == "null" ]]; then
    echo "Text does not exist"
    res=1
fi

if (( "${#text}" == 0 )); then 
    echo "Text is empty"
    res=1
fi

oc delete -n test-vllm route granite
oc delete -n test-vllm InferenceService granite-3b-code-base-2k
oc delete -n test-vllm ServingRuntime vllm-runtime
oc delete ns test-vllm

exit "${res}"
