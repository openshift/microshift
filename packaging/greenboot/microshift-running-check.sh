#!/bin/bash
set -e

SCRIPT_NAME=$(basename "$0")
SCRIPT_PID=$$
PODS_NS_LIST=(openshift-ovn-kubernetes openshift-service-ca openshift-ingress openshift-dns openshift-storage kube-system)
PODS_CT_LIST=(2                        1                    1                 2             2                 2)

# Source the MicroShift health check functions library
# shellcheck source=packaging/greenboot/functions.sh
source /usr/share/microshift/functions/greenboot.sh

# Set the term handler to convert exit code to 1
trap 'return_failure' TERM

# Set the exit handler to log the exit status
trap 'script_exit' EXIT

# The term handler to override the default behavior and have a uniform and
# homogeneous exit code in all controlled situations.
function return_failure() {
    exit 1
}

# The script exit handler logging the FAILURE or FINISHED message depending
# on the exit status of the last command
#
# args: None
# return: None
function script_exit() {
    [ "$?" -ne 0 ] && echo "FAILURE" || echo "FINISHED"
}

# Check the microshift.service systemd unit activity, terminating the script
# with the SIGTERM signal if the unit reports a failed state
#
# args: None
# return: 0 if the systemd unit is active, or 1 otherwise
function microshift_service_active() {
    local -r is_failed=$(systemctl is-failed microshift.service)
    local -r is_active=$(systemctl is-active microshift.service)

    # Terminate the script in case of a failed service - nothing to wait for
    if [ "${is_failed}" = "failed" ] ; then
        echo "The microshift.service systemd unit is failed. Terminating..."
        kill -TERM ${SCRIPT_PID}
    fi
    # Check the service activity
    [ "${is_active}" = "active" ] && return 0
    return 1
}

# Check if MicroShift API 'readyz' and 'livez' health endpoints are OK
#
# args: None
# return: 0 if all API health endpoints are OK, or 1 otherwise
function microshift_health_endpoints_ok() {
    local -r check_rd=$(${OCGET_CMD} --raw='/readyz?verbose' | awk '$2 != "ok"')
    local -r check_lv=$(${OCGET_CMD} --raw='/livez?verbose'  | awk '$2 != "ok"')

    [ "${check_rd}" != "readyz check passed" ] && return 1
    [ "${check_lv}" != "livez check passed"  ] && return 1
    return 0
}

# Check if any MicroShift pods are in the 'Running' status
#
# args: None
# return: 0 if any pods are in the 'Running' status, or 1 otherwise
function any_pods_running() {
    local -r count=$(${OCGET_CMD} pods ${OCGET_OPT} -A 2>/dev/null | awk '$4~/Running/' | wc -l)

    [ "${count}" -gt 0 ] && return 0
    return 1
}

#
# Main
#

# Exit if the current user is not 'root'
if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "STARTED"

# Print the boot variable status
print_boot_status

# Exit if the MicroShift service is not enabled
if [ "$(systemctl is-enabled microshift.service 2>/dev/null)" != "enabled" ] ; then
    echo "MicroShift service is not enabled. Exiting..."
    exit 0
fi

# Set the wait timeout for the current check based on the boot counter
WAIT_TIMEOUT_SECS=$(get_wait_timeout)

# Wait for MicroShift service to be active (failed status terminates the script)
echo "Waiting ${WAIT_TIMEOUT_SECS}s for MicroShift service to be active and not failed"
wait_for "${WAIT_TIMEOUT_SECS}" microshift_service_active

# Wait for MicroShift API health endpoints to be OK
echo "Waiting ${WAIT_TIMEOUT_SECS}s for MicroShift API health endpoints to be OK"
wait_for "${WAIT_TIMEOUT_SECS}" microshift_health_endpoints_ok

# Wait for any pods to enter running state
echo "Waiting ${WAIT_TIMEOUT_SECS}s for any pods to be running"
wait_for "${WAIT_TIMEOUT_SECS}" any_pods_running

# Wait for MicroShift core pod images to be downloaded
for i in "${!PODS_NS_LIST[@]}"; do
    CHECK_PODS_NS=${PODS_NS_LIST[${i}]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for pod image(s) from the '${CHECK_PODS_NS}' namespace to be downloaded"
    wait_for "${WAIT_TIMEOUT_SECS}" namespace_images_downloaded
done

# Wait for MicroShift core pods to enter ready state
for i in "${!PODS_NS_LIST[@]}"; do
    CHECK_PODS_NS=${PODS_NS_LIST[${i}]}
    CHECK_PODS_CT=${PODS_CT_LIST[${i}]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for ${CHECK_PODS_CT} pod(s) from the '${CHECK_PODS_NS}' namespace to be in 'Ready' state"
    wait_for "${WAIT_TIMEOUT_SECS}" namespace_pods_ready
done

# Verify that MicroShift core pods are not restarting
pids=()
for i in "${!PODS_NS_LIST[@]}"; do
    CHECK_PODS_NS=${PODS_NS_LIST[${i}]}

    echo "Checking pod restart count in the '${CHECK_PODS_NS}' namespace"
    namespace_pods_not_restarting "${CHECK_PODS_NS}" &
    pids+=($!)
done

for pid in "${pids[@]}"; do
    wait "${pid}"
done
