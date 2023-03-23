#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'
set -x

SCRIPT_PATH="$(readlink -f "$(dirname "${BASH_SOURCE[0]}")")"
source "${SCRIPT_PATH}/../helpers.sh"

cleanup() {
  oc delete service hello-microshift
  oc delete -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
}
trap cleanup EXIT

gcloud compute firewall-rules update "${INSTANCE_PREFIX}" --allow tcp:22,icmp,tcp:5678
oc create -f "${SCRIPT_PATH}/assets/hello-microshift.yaml"
oc create service loadbalancer hello-microshift --tcp=5678:8080
oc wait pods -l app=hello-microshift --for condition=Ready --timeout=60s

set +x
retries=3
backoff=3s
for try in $(seq 1 "${retries}"); do
  echo "Attempt: ${try}"
  echo "Running: curl -vk ${IP}:5678"
  RESPONSE=$(curl -I ${IP}:5678 2>&1)
  RESULT=$?
  echo "Exit code: ${RESULT}"
  echo -e "Response: \n${RESPONSE}\n\n"
  if [ $RESULT -eq 0 ] && echo "${RESPONSE}" | grep -q -E "HTTP.*200"; then
    echo "Request fulfilled conditions to be successful (exit code = 0, response contains 'HTTP.*200')"
    exit 0
  fi
  echo -e "Waiting ${backoff} before next retry\n\n"
  sleep "${backoff}"
done
set -x

exit 1
