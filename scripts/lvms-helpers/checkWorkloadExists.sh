#!/bin/bash

set -eux

# Set KUBECONFIG path - use dynamic IP-based path
KUBECONFIG=${KUBECONFIG:-/var/lib/microshift/resources/kubeadmin/kubeconfig}

ns="test-lvms"
appLabel="app-lvms"

echo "INFO:" "Check if deployment Pod is 'Running'......."
iter=24
period=5
podName=$(sudo oc --kubeconfig "${KUBECONFIG}" get pod -n ${ns} -l app=${appLabel} --no-headers | awk '{print $1}')
if [ "${podName}" == "" ]; then
  echo "ERROR:" "Deployment Pod not found."
  exit 1
fi

echo "INFO:" "Waiting for Pod to become ready(max 2-minutes)....."
result=""
while [[ "${result}" != "Running" && ${iter} -gt 0 ]]; do
  #shellcheck disable=SC1083
  result=$(sudo oc --kubeconfig "${KUBECONFIG}" get pod "${podName}" -n "${ns}" -o=jsonpath={.status.phase})
  (( iter -- ))
  sleep ${period}
done
if [ "${result}" == "Running" ]; then
  echo "INFO:" "Deployment Pod is Running."
else
  echo "ERROR:" "Deployment Pod is not in 'Running' state."
  sudo oc --kubeconfig "${KUBECONFIG}" -n "${ns}" describe pod "${podName}"
  exit 1
fi

echo "INFO:" "Check if previously written data exists in Pod mounted volume......."
data=$(sudo oc --kubeconfig "${KUBECONFIG}" exec -n "${ns}" "${podName}" -- /bin/sh -c 'cat /mnt/storage/testfile')
if [[ ${data} =~ "Storage_Test" ]]; then
  echo "SUCCESS:" "Data written before MicroShift upgrade still exists in the Pod volume"
else
  echo "ERROR:" "Data written before MicroShift upgrade not found in the Pod volume"
  exit 1
fi

echo "INFO:" "Check if new data can be writen/read into previously created Pod mounted volume......."
#shellcheck disable=SC2016
sudo oc --kubeconfig "${KUBECONFIG}" exec -n "${ns}" "${podName}" -- /bin/sh -c 'echo Storage_Test $(date) > /mnt/storage/testfile2'
data=$(sudo oc --kubeconfig "${KUBECONFIG}" exec -n "${ns}" "${podName}" -- /bin/sh -c 'cat /mnt/storage/testfile2')
if [[ ${data} =~ "Storage_Test" ]]; then
  echo "SUCCESS:" "Data successfully written into the previously created Pod mounted volume"
else
  echo "ERROR:" "Failed to write data into the previously created Pod mounted volume"
  exit 1
fi
