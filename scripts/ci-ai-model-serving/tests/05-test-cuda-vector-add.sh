#!/usr/bin/env bash

set -xeuo pipefail

oc create namespace cuda-test

cat << EOF | oc apply -f -
apiVersion: batch/v1
kind: Job
metadata:
  name: test-cuda-vector-add
  namespace: cuda-test
spec:
  template:
    spec:
      restartPolicy: Never
      containers:
      - name: cuda-vector-add
        image: "nvcr.io/nvidia/k8s/cuda-sample:vectoradd-cuda12.5.0-ubi8"
        resources:
          limits:
            nvidia.com/gpu: 1
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            drop: ["ALL"]
          runAsNonRoot: true
          seccompProfile:
            type: "RuntimeDefault"
EOF

# Wait till complete
oc wait --for=condition=complete --timeout=120s -n cuda-test job/test-cuda-vector-add

# Get the logs
pod=$(oc get pods -n cuda-test --selector=batch.kubernetes.io/job-name=test-cuda-vector-add --output=jsonpath='{.items[*].metadata.name}')
logfile=$(mktemp)
oc logs -n cuda-test "${pod}" > "${logfile}"

oc delete job -n cuda-test test-cuda-vector-add
oc delete ns cuda-test

if ! grep -q -E '^Test PASSED$' "${logfile}"; then
    echo "CUDA vector-add test failed"
    exit 1
fi
