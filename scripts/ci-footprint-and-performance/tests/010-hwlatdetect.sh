#!/bin/bash

set -xeuo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=scripts/ci-footprint-and-performance/tests/functions.sh
source "${SCRIPTDIR}/functions.sh"

TEST_DURATION=$(( 2 * 60 ))
POD=hwlatdetect

# Following is a container running utility `hwlatdetect`.
# It's used for verifying if hardware platform is suitable for RT operations.
# It looks for latencies introduced by the hardware or BIOS.
sudo tee /tmp/hwlatdetect.yaml << EOF
apiVersion: v1 
kind: Pod 
metadata:
  name: ${POD} 
spec:
  containers:
  - name: hwlatdetect 
    image: quay.io/container-perf-tools/hwlatdetect
    imagePullPolicy: Always 
    # Request and Limits are not required - hwlat detector is done in the kernel
    env:
    - name: tool
      value: "hwlatdetect"
    - name: RUNTIME_SECONDS 
      value: "${TEST_DURATION}"
    - name: THRESHOLD
      value: "2"
    - name: EXTRA_ARGS
      value: "--report /tmp/hwlatdetect.txt"
    securityContext:
      # Required for access to /dev/cpu_dma_latency and /sys/kernel/debug on the host
      privileged: true
EOF

create_and_wait_for_pod /tmp/hwlatdetect.yaml "${POD}"
wait_for_test_to_finish "${POD}" "${TEST_DURATION}" "test finished"

oc describe pod "${POD}" > "${LOW_LAT_ARTIFACTS}/hwlatdetect-pod.txt"
oc logs "${POD}" > "${LOW_LAT_ARTIFACTS}/hwlatdetect.log"
oc exec "${POD}" -- cat /tmp/hwlatdetect.txt > "${LOW_LAT_ARTIFACTS}/hwlatdetect-results.txt"

oc delete -f "/tmp/${POD}.yaml"

cat "${LOW_LAT_ARTIFACTS}/hwlatdetect-pod.txt"
cat "${LOW_LAT_ARTIFACTS}/hwlatdetect.log"
cat "${LOW_LAT_ARTIFACTS}/hwlatdetect-results.txt"
