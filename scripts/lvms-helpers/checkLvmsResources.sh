#!/bin/bash

set -eux

# Set KUBECONFIG path - use dynamic IP-based path
KUBECONFIG=${KUBECONFIG:-/var/lib/microshift/resources/kubeadmin/kubeconfig}

ns="openshift-storage"

echo "INFO:" "Check if LVMS resource pods are 'Running'......."
iter=24
period=5
result=""
for pod in $(sudo oc --kubeconfig "${KUBECONFIG}" get pods -n ${ns} --no-headers | awk '{print $1}'); do
  while [[ "${result}" != "Running" && ${iter} -gt 0 ]]; do
    #shellcheck disable=SC1083
    result=$(sudo oc --kubeconfig "${KUBECONFIG}" get pod "${pod}" -n ${ns} -o=jsonpath={.status.phase})
    (( iter -- ))
    sleep ${period}
  done
  if [ "${result}" == "Running" ]; then
    echo "INFO:" "LVMS resource Pod: ${pod} is Running as expected."
  else
    echo "ERROR:" "LVMS resource Pod: ${pod} is not in 'Running' state."
    sudo oc --kubeconfig "${KUBECONFIG}" -n ${ns} describe pod "${pod}"
    exit 1
  fi
done

echo "INFO:" "Check if LVMCluster resource is in 'Ready' state......."
iter=24
period=5
result=""
lvmClusterName=$(sudo oc --kubeconfig "${KUBECONFIG}" get lvmcluster -n ${ns} --no-headers | awk '{print $1}')
if [ "${lvmClusterName}" == "" ]; then
  echo "ERROR:" "LVMCluster resource not found."
  exit 1
fi
while [[ "${result}" != "Ready" && ${iter} -gt 0 ]]; do
  #shellcheck disable=SC1083
  result=$(sudo oc --kubeconfig "${KUBECONFIG}" get lvmcluster "${lvmClusterName}" -n ${ns} -o=jsonpath={.status.state})
  (( iter -- ))
  sleep ${period}
done
if [ "${result}" == "Ready" ]; then
  echo "SUCCESS:" "LVMSCluster resource: ${lvmClusterName} is in Ready state."
else
  echo "ERROR:" "LVMSCluster resource: ${lvmClusterName} is not in 'Ready' state."
  sudo oc --kubeconfig "${KUBECONFIG}" -n ${ns} describe lvmcluster "${lvmClusterName}"
  exit 1
fi
