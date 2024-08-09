#!/bin/bash
set -e

SCRIPT_NAME=$(basename "$0")
SCRIPT_PID=$$
PODS_NS_LIST=(openshift-ovn-kubernetes openshift-service-ca openshift-ingress openshift-dns)
PODS_CT_LIST=(2 1 1 2)
LOG_POD_EVENTS=false

# Source the MicroShift health check functions library
# shellcheck source=packaging/greenboot/functions.sh
source /usr/share/microshift/functions/greenboot.sh

# Set the term handler to convert exit code to 1
trap 'forced_termination' TERM SIGINT

# Set the exit handler to log the exit status
trap 'log_script_exit' EXIT

# Handler that will be called when the script is terminated by sending TERM or
# INT signals. To override default exit codes it forces returning 1 like the
# rest of the error conditions throughout the health check.
function forced_termination() {
    echo "Signal received, terminating."
    exit 1
}

# Check preconditions for existence of lvms deployment.
# Adapted from MicroShift code.
#
# args: None
# return: 0 if lvms readiness should be checked, 1 otherwise
function lvmsShouldBeDeployed() {
    if ! hash vgs 2>/dev/null; then
        return 1
    fi
    if [ -f /etc/microshift/lvmd.yaml ]; then
        return 0
    fi
    if ! lvmsDriverShouldExist; then
        return 1
    fi

    local -r volume_groups=$(vgs --readonly --options=name --noheadings)
    local -r volume_groups_count=$(echo "${volume_groups}" | wc -w)
    if [ "${volume_groups_count}" -eq 0 ]; then
        return 1
    elif [ "${volume_groups_count}" -eq 1 ]; then
        return 0
    elif echo "${volume_groups}" | grep -qw "microshift"; then
        return 0
    else
        return 1
    fi
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
# TODO: Remove when `microshift healthcheck` is complete.
if [ "$(systemctl is-enabled microshift.service 2>/dev/null)" != "enabled" ] ; then
    echo "MicroShift service is not enabled. Exiting..."
    exit 0
fi

# Set the wait timeout for the current check based on the boot counter
WAIT_TIMEOUT_SECS=$(get_wait_timeout)

# Always log potential MicroShift upgrade errors on failure
LOG_FAILURE_FILES+=("/var/lib/microshift-backups/prerun_failed.log")

/usr/bin/microshift healthcheck -v=2 --timeout="${WAIT_TIMEOUT_SECS}s"

# Wait for MicroShift API health endpoints to be OK
echo "Waiting ${WAIT_TIMEOUT_SECS}s for MicroShift API health endpoints to be OK"
if ! wait_for "${WAIT_TIMEOUT_SECS}" microshift_health_endpoints_ok ; then
    log_failure_cmd "health-readyz" "${OCGET_CMD} --raw=/readyz?verbose"
    log_failure_cmd "health-livez"  "${OCGET_CMD} --raw=/livez?verbose"

    echo "Error: Timed out waiting for MicroShift API health endpoints to be OK"
    exit 1
fi

if lvmsShouldBeDeployed; then
    PODS_NS_LIST+=(openshift-storage)
    PODS_CT_LIST+=(2)
fi
declare -a csi_components=('csi-snapshot-controller' 'csi-snapshot-webhook')
csi_pods_ct=0
for csi_c in "${csi_components[@]}"; do
    if csiComponentShouldBeDeployed "${csi_c}"; then
        (( csi_pods_ct += 1 ))
    fi
done
if [ ${csi_pods_ct} -gt 0 ]; then
    PODS_NS_LIST+=(kube-system)
    PODS_CT_LIST+=("${csi_pods_ct}")
fi

# Starting pod-specific checks
# Log list of pods and their events on failure
LOG_POD_EVENTS=true

# Wait for any pods to enter running state
echo "Waiting ${WAIT_TIMEOUT_SECS}s for any pods to be running"
if ! wait_for "${WAIT_TIMEOUT_SECS}" any_pods_running ; then
    echo "Error: Timed out waiting for any MicroShift pod to be running"
    exit 1
fi

# Wait for MicroShift core pod images to be downloaded
for i in "${!PODS_NS_LIST[@]}"; do
    CHECK_PODS_NS=${PODS_NS_LIST[${i}]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for pod image(s) from the '${CHECK_PODS_NS}' namespace to be downloaded"
    if ! wait_for "${WAIT_TIMEOUT_SECS}" namespace_images_downloaded; then
        echo "Error: Timed out waiting for pod image(s) from the '${CHECK_PODS_NS}' namespace to be downloaded"
        exit 1
    fi
done

# Wait for MicroShift core pods to enter ready state
for i in "${!PODS_NS_LIST[@]}"; do
    CHECK_PODS_NS=${PODS_NS_LIST[${i}]}
    CHECK_PODS_CT=${PODS_CT_LIST[${i}]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for ${CHECK_PODS_CT} pod(s) from the '${CHECK_PODS_NS}' namespace to be in 'Ready' state"
    if ! wait_for "${WAIT_TIMEOUT_SECS}" namespace_pods_ready; then
        echo "Error: Timed out waiting for ${CHECK_PODS_CT} pod(s) in the '${CHECK_PODS_NS}' namespace to be in 'Ready' state"
        exit 1
    fi
done

# Verify that MicroShift core pods are not restarting
declare -A pid2name
for i in "${!PODS_NS_LIST[@]}"; do
    CHECK_PODS_NS=${PODS_NS_LIST[${i}]}

    echo "Checking pod restart count in the '${CHECK_PODS_NS}' namespace"
    namespace_pods_not_restarting "${CHECK_PODS_NS}" &
    pid=$!

    pid2name["${pid}"]="${CHECK_PODS_NS}"
done

# Wait for the restart check functions to complete, printing errors in case of a failure
check_failed=false
for pid in "${!pid2name[@]}"; do
    if ! wait "${pid}" ; then
        check_failed=true

        name=${pid2name["${pid}"]}
        echo "Error: Pods are restarting too frequently in the '${name}' namespace"
    fi
done

# Exit with an error code if the pod restart check failed
if ${check_failed} ; then
    exit 1
fi
