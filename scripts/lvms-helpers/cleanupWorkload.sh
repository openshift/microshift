#!/bin/bash

set -ux

# Set KUBECONFIG path - use dynamic IP-based path
KUBECONFIG=${KUBECONFIG:-/var/lib/microshift/resources/kubeadmin/kubeconfig}

result=""
ns="test-lvms"

echo "INFO:" "Delete deployment storage resource......."
result=$(sudo oc --kubeconfig "${KUBECONFIG}" delete deployment mydep-lvms -n ${ns})
if [[ ${result} =~ "deleted" ]]; then
  echo "SUCCESS:" "Deployment resource deleted successfully."
else
  echo "ERROR:" "Deployment resource not found."
fi

echo "INFO:" "Delete PVC storage resource......."
result=$(sudo oc --kubeconfig "${KUBECONFIG}" delete pvc mypvc-lvms -n ${ns})
if [[ ${result} =~ "deleted" ]]; then
  echo "SUCCESS:" "PVC resource deleted successfully."
else
  echo "ERROR:" "PVC resource not found."
fi

echo "INFO:" "Delete test namepsace......."
result=$(sudo oc --kubeconfig "${KUBECONFIG}" delete ns ${ns})
if [[ ${result} =~ "deleted" ]]; then
  echo "SUCCESS:" "Namespace deleted successfully."
else
  echo "ERROR:" "Namespace resource not found."
fi
