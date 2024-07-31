#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/ci-footprint-and-performance/tests/functions.sh
source "${SCRIPTDIR}/functions.sh"

TEST_DURATION=$(( 30 * 60 ))
POD=oslat

# oslat is an application behaving similar to CPU-intensive DPDK application.
# It measures all the interruptions and disruptions to the busy loop that simulates CPU heavy data processing.
sudo tee "/tmp/oslat.yaml" << EOF
apiVersion: v1
kind: Pod
metadata:
  name: ${POD}
  annotations:
    cpu-load-balancing.crio.io: "disable"
    irq-load-balancing.crio.io: "disable"
    cpu-quota.crio.io: "disable"
spec:
  runtimeClassName: microshift-low-latency
  containers:
  - name: oslat
    image: quay.io/container-perf-tools/oslat
    imagePullPolicy: Always
    resources:
      requests:
        memory: "400Mi"
        cpu: "10"
      limits:
        memory: "400Mi"
        cpu: "10"
    env:
    - name: tool
      value: "oslat"
    - name: manual
      value: "n"
    - name: PRIO
      value: "1"
    - name: delay
      value: "0"
    - name: RUNTIME_SECONDS
      value: "${TEST_DURATION}"
    - name: TRACE_THRESHOLD
      value: ""
    - name: EXTRA_ARGS
      value: "--json=/tmp/oslat.json --bucket-width 500 --zero-omit"
    securityContext:
      privileged: true
      capabilities:
        add:
          - SYS_NICE
          - IPC_LOCK
EOF

create_and_wait_for_pod "/tmp/oslat.yaml" "${POD}"
wait_for_test_to_finish "${POD}" "${TEST_DURATION}" "Test completed."

oc describe pod "${POD}" > "${LOW_LAT_ARTIFACTS}/oslat-pod.txt"
oc logs "${POD}" > "${LOW_LAT_ARTIFACTS}/oslat.log"
oc exec "${POD}" -- cat /tmp/oslat.json > "${LOW_LAT_ARTIFACTS}/oslat.json"

oc delete -f /tmp/oslat.yaml

cat "${LOW_LAT_ARTIFACTS}/oslat-pod.txt"
cat "${LOW_LAT_ARTIFACTS}/oslat.log"
cat "${LOW_LAT_ARTIFACTS}/oslat.json"
