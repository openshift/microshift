#!/bin/bash

# Define our check command
COMMAND="oc get pods -A -o 'jsonpath={..status.conditions[?(@.type==\"Ready\")].status}'"
# Define the specific output we are waiting for
EXPECTED_PODS=6
ALL_PODS=10
# Define the location of the microshift kubeconfig
KUBECONFIG="/var/lib/microshift/resources/kubeadmin/kubeconfig"
USER_KUBECONFIG="${HOME}/.kube"

# Start the timer
START_TIME=$(date +%s)

# Start the microshift service
sudo systemctl start microshift.service

# Check to see how long microshift.service took to start
SYSTEMD_BLAME=$(systemd-analyze blame | grep microshift.service)
echo "${SYSTEMD_BLAME}" | awk '{print $2 ": " $1}' | sed 's/s$/ seconds/'

# Loop until the kubeconfig exists
while ! sudo [ -e "${KUBECONFIG}" ]; do
  sleep 1
done

# Calculate the time it took for kubeconfig to exist
KUBE_TIME=$(date +%s)
DURATION=$((KUBE_TIME - START_TIME))

echo "Kubeconfig: ${DURATION} seconds"

# Copy the kubeconfig to the user homedir
[ ! -d "${USER_KUBECONFIG}" ] && mkdir "${USER_KUBECONFIG}"
sudo cp ${KUBECONFIG} "${USER_KUBECONFIG}"/config
sudo chown "${USER}":"${USER}" "${USER_KUBECONFIG}"/config

# podcheck_nostorage fn waits for an expected number of non-storage pods to be in Ready state
podcheck_nostorage() {
  expected=$1
  while true; do
      OUTPUT=$(eval "oc get po -A --no-headers")
      PODS_READY=$(echo "${OUTPUT}" | grep -vE "csi|lvm" | grep -c Running)

      # Wait until all pods report ready
      if [[ ${PODS_READY} -ge ${expected} ]]; then
          break
      fi
      sleep 1
  done

  oc get po -A

  # Calculate the time it took for all pods to be running
  END_TIME=$(date +%s)
  DURATION=$((END_TIME - START_TIME))

  echo "Boot: ${DURATION} seconds (${expected} pods)"
}

# podcheck fn waits for an expected number of pods to be in Ready state
podcheck() {
  expected=$1
  while true; do
      OUTPUT=$(eval "${COMMAND}")
      PODS_READY=$(echo "${OUTPUT}" | grep -o -w "True" | wc -l)

      # Wait until all pods report ready
      if [[ ${PODS_READY} -ge ${expected} ]]; then
          break
      fi
      sleep 1
  done

  oc get po -A

  # Calculate the time it took for all pods to be running
  END_TIME=$(date +%s)
  DURATION=$((END_TIME - START_TIME))

  echo "Boot: ${DURATION} seconds (${expected} pods)"
}

podcheck_nostorage ${EXPECTED_PODS}
podcheck ${ALL_PODS}
