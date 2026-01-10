#!/bin/bash

set -eux

# Set KUBECONFIG path - use dynamic IP-based path
KUBECONFIG=${KUBECONFIG:-/var/lib/microshift/resources/kubeadmin/kubeconfig}

ns="test-lvms"
appLabel="app-lvms"

echo "INFO:" "Waiting for deployment Pod to be created (max 2-minutes)....."
iter=24
period=5
podName=""
while [[ "${podName}" == "" && ${iter} -gt 0 ]]; do
  podName=$(sudo oc --kubeconfig "${KUBECONFIG}" get pod -n ${ns} -l app=${appLabel} --no-headers 2>/dev/null | awk '{print $1}')
  if [ "${podName}" == "" ]; then
    (( iter -- ))
    sleep ${period}
  fi
done

if [ "${podName}" == "" ]; then
  echo "ERROR:" "Deployment Pod not found after waiting."
  exit 1
fi
echo "INFO:" "Pod ${podName} found."

# Now wait for the pod to be ready (all containers initialized and ready)
echo "INFO:" "Waiting for Pod and all containers to be ready (max 3-minutes)....."
if sudo oc --kubeconfig "${KUBECONFIG}" wait --for=condition=Ready pod/"${podName}" -n "${ns}" --timeout=180s; then
  echo "INFO:" "Deployment Pod is Ready."
else
  echo "ERROR:" "Pod did not become ready within timeout."
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
