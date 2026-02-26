#!/bin/bash

DEFAULT_EXPECTED_PODS=6
DEFAULT_ALL_PODS=10

# Validate arguments
if [[ $# -eq 1 ]]; then
  echo "Error: When providing arguments, you must specify both EXPECTED_PODS and ALL_PODS"
  echo "Usage: $0 [EXPECTED_PODS ALL_PODS]"
  exit 1
fi

if [[ $# -eq 2 ]]; then
  # Validate that both arguments are numbers
  if ! [[ $1 =~ ^[0-9]+$ ]]; then
    echo "Error: EXPECTED_PODS must be a number"
    echo "Usage: $0 [EXPECTED_PODS ALL_PODS]"
    exit 1
  fi

  if ! [[ $2 =~ ^[0-9]+$ ]]; then
    echo "Error: ALL_PODS must be a number"
    echo "Usage: $0 [EXPECTED_PODS ALL_PODS]"
    exit 1
  fi

  # Validate that EXPECTED_PODS is less than or equal to ALL_PODS
  if [[ $1 -gt $2 ]]; then
    echo "Error: EXPECTED_PODS ($1) cannot be greater than ALL_PODS ($2)"
    exit 1
  fi
fi

EXPECTED_PODS=${1:-${DEFAULT_EXPECTED_PODS}}
ALL_PODS=${2:-${DEFAULT_ALL_PODS}}

# Timeout in seconds (10 minutes) - fail fast to get useful output before CI kills the job
TIMEOUT=600

# Define the location of the microshift kubeconfig
KUBECONFIG="/var/lib/microshift/resources/kubeadmin/kubeconfig"
USER_KUBECONFIG="${HOME}/.kube"

echo "Waiting for MicroShift to start with ${EXPECTED_PODS} expected non-storage pods and ${ALL_PODS} total pods"

# Start the timer
START_TIME=$(date +%s)

# Start the microshift service
sudo systemctl start microshift.service

# Check to see how long microshift.service took to start
SYSTEMD_BLAME=$(systemd-analyze blame | grep microshift.service)
echo "${SYSTEMD_BLAME}" | awk '{print $2 ": " $1}' | sed 's/s$/ seconds/'

# Loop until the kubeconfig exists, checking service health
while ! sudo [ -e "${KUBECONFIG}" ]; do
  # Check if service has failed
  if systemctl is-failed --quiet microshift.service; then
    echo "ERROR: microshift.service failed to start"
    echo "Service status:"
    systemctl status microshift.service --no-pager
    echo ""
    echo "Recent journal logs:"
    journalctl -u microshift.service --no-pager -n 50
    exit 1
  fi

  # Check for timeout
  elapsed=$(( $(date +%s) - START_TIME ))
  if [[ ${elapsed} -gt ${TIMEOUT} ]]; then
    echo "ERROR: Timed out after ${elapsed}s waiting for kubeconfig"
    systemctl status microshift.service --no-pager
    exit 1
  fi

  sleep 1
done

# Calculate the time it took for kubeconfig to exist
KUBE_TIME=$(date +%s)
DURATION=$((KUBE_TIME - START_TIME))

echo "Kubeconfig: ${DURATION} seconds"

# Copy the kubeconfig to the user homedir
mkdir -p "${USER_KUBECONFIG}"
sudo cp "${KUBECONFIG}" "${USER_KUBECONFIG}"/config
sudo chown "${USER}":"${USER}" "${USER_KUBECONFIG}"/config

count_ready_nostorage() {
  oc get po -A --no-headers | grep -vE "csi|storage" | grep -c Running
}

count_ready_all() {
  oc get pods -A -o "jsonpath={..status.conditions[?(@.type==\"Ready\")].status}" \
    | grep -o -w "True" \
    | wc -l
}

wait_for_ready() {
  local label=$1
  local expected=$2
  local count_fn=$3
  local result_var=$4
  local prev_ready=-1

  while true; do
    # Check for timeout
    local elapsed=$(( $(date +%s) - START_TIME ))
    if [[ ${elapsed} -gt ${TIMEOUT} ]]; then
      echo "ERROR: Timed out after ${elapsed}s waiting for ${label} (${prev_ready}/${expected} ready)"
      echo "Final pod status:"
      oc get pods -A
      echo ""
      echo "Recent events:"
      oc get events -A --sort-by='.lastTimestamp' | tail -20
      exit 1
    fi

    local ready
    ready=$(${count_fn})

    # Print progress when pod count changes
    if [[ ${ready} -ne ${prev_ready} ]]; then
      elapsed=$(( $(date +%s) - START_TIME ))
      echo "[${elapsed}s] ${label}: ${ready}/${expected} ready"
      # Show which pods are not yet ready
      oc get pods -A -o json \
        | jq -r '.items[] | select(.status.conditions[]? | select(.type=="Ready" and .status!="True")) | "  NOT READY: \(.metadata.namespace)/\(.metadata.name)"' \
        2>/dev/null || true
      prev_ready=${ready}
    fi

    # Wait until all pods report ready
    if [[ ${ready} -ge ${expected} ]]; then
      break
    fi
    sleep 1
  done

  oc get po -A

  # Calculate the time it took for all pods to be running
  local end_time
  local duration
  end_time=$(date +%s)
  duration=$((end_time - START_TIME))

  echo "Ready (${label}): ${duration} seconds (${expected} pods)"
  if [[ -n ${result_var} ]]; then
    printf -v "${result_var}" "%s" "${duration}"
  fi
}

READY_SECONDS_NON_STORAGE=""
READY_SECONDS_ALL=""

wait_for_ready "Non-storage pods" "${EXPECTED_PODS}" count_ready_nostorage READY_SECONDS_NON_STORAGE
wait_for_ready "All pods" "${ALL_PODS}" count_ready_all READY_SECONDS_ALL

END_TIME=$(date +%s)
echo "{\"ready_seconds_non_storage\":${READY_SECONDS_NON_STORAGE},\"ready_seconds_all\":${READY_SECONDS_ALL},\"start_epoch\":${START_TIME},\"end_epoch\":${END_TIME}}"
