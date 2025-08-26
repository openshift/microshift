#!/usr/bin/env bash

set -xeuo pipefail

NS=test-triton
TRITON_IMAGE="nvcr.io/nvidia/tritonserver:25.06-py3"
MODEL_IMAGE="quay.io/microshift/ai-testing-model:onnx-densenet-121"

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
oc create ns "${NS}"

# Pull the images before creating K8s objects to make the test more deterministic.
pull_image "${TRITON_IMAGE}"
pull_image "${MODEL_IMAGE}"

# Create ServingRuntime
# ServingRuntime definition from https://docs.redhat.com/en/documentation/red_hat_openshift_ai_self-managed/2.22/html-single/serving_models/index#adding-a-tested-and-verified-model-serving-runtime-for-the-single-model-serving-platform_serving-large-models
# Triton image is from https://catalog.ngc.nvidia.com/orgs/nvidia/containers/tritonserver
cat <<EOF | oc apply -n "${NS}" -f -
apiVersion: serving.kserve.io/v1alpha1
kind: ServingRuntime
metadata:
  name: triton-kserve-rest
  labels:
    opendatahub.io/dashboard: "true"
spec:
  annotations:
    prometheus.kserve.io/path: /metrics
    prometheus.kserve.io/port: "8002"
  containers:
    - args:
        - tritonserver
        - --model-store=/mnt/models
        - --grpc-port=9000
        - --http-port=8080
        - --allow-grpc=true
        - --allow-http=true
      image: ${TRITON_IMAGE}
      name: kserve-container
      resources:
        limits:
          cpu: "1"
          memory: 2Gi
        requests:
          cpu: "1"
          memory: 2Gi
      ports:
        - containerPort: 8080
          protocol: TCP
  protocolVersions:
    - v2
    - grpc-v2
  supportedModelFormats:
    - autoSelect: true
      name: tensorrt
      version: "8"
    - autoSelect: true
      name: tensorflow
      version: "1"
    - autoSelect: true
      name: tensorflow
      version: "2"
    - autoSelect: true
      name: onnx
      version: "1"
    - name: pytorch
      version: "1"
    - autoSelect: true
      name: triton
      version: "2"
    - autoSelect: true
      name: xgboost
      version: "1"
    - autoSelect: true
      name: python
      version: "1"
EOF

# Create InferenceService
cat <<EOF | oc apply -n "${NS}" -f -
apiVersion: "serving.kserve.io/v1beta1"
kind: "InferenceService"
metadata:
  name: onnx-triton
spec:
  predictor:
    model:
      protocolVersion: v2
      runtime: triton-kserve-rest
      modelFormat:
        name: onnx
      storageUri: oci://${MODEL_IMAGE}
      resources:
        limits:
          nvidia.com/gpu: 1
        requests:
          nvidia.com/gpu: 1
EOF

# Create Route for the InferenceService
cat <<EOF | oc apply -n "${NS}" -f -
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: onnx-triton
spec:
  host: onnx-triton.apps.example.com
  port:
    targetPort: 8080
  to:
    kind: Service
    name: onnx-triton-predictor
    weight: 100
  wildcardPolicy: None
EOF

# Wait for the Deployment that kserve created based on the InferenceService created earlier.
# 10m of timeout because it might take a while to unpack the model from the OCI.
sudo microshift healthcheck \
    -v 2 \
    --timeout 10m0s \
    --namespace "${NS}" \
    --deployments onnx-triton-predictor

# Following verification procedure is based on https://github.com/triton-inference-server/tutorials/tree/17331012af74eab68ad7c86d8a4ae494272ca4f7/Quick_Deploy/OpenVINO#deploying-an-onnx-model
temp_dir=$(mktemp -d /tmp/triton-test-XXXXXX)
pushd "${temp_dir}"

curl -o client.py "https://raw.githubusercontent.com/triton-inference-server/tutorials/17331012af74eab68ad7c86d8a4ae494272ca4f7/Quick_Deploy/ONNX/client.py"
sed -i 's,url="localhost:8000",url="onnx-triton.apps.example.com",' ./client.py

curl -o img1.jpg "https://www.hakaimagazine.com/wp-content/uploads/header-gulf-birds.jpg"

python3 -m venv ./venv/
./venv/bin/python -m pip install 'tritonclient[all]==2.59' torchvision==0.22

hosts="$(hostname -i) onnx-triton.apps.example.com"
if ! sudo grep -q "${hosts}" /etc/hosts; then
    echo "${hosts}" | sudo tee -a /etc/hosts
fi
cat /etc/hosts

output="$(./venv/bin/python client.py)"
echo "Query output: ${output}"

res=0
expected_output="['11.548589:92' '11.231404:14' '7.527273:95' '6.922710:17' '6.576274:88']"
if [[ "${output}" != "${expected_output}" ]]; then
    echo "Unexpected output"
    res=1
fi

popd

oc delete -n "${NS}" route onnx-triton
oc delete -n "${NS}" InferenceService onnx-triton
oc delete -n "${NS}" ServingRuntime triton-kserve-rest
oc delete ns "${NS}"

exit "${res}"
