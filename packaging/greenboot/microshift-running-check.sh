#!/bin/bash
set -e

SCRIPT_NAME=$(basename $0)
SCRIPT_PID=$$
OCGET_CMD="oc get --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig"
OCGET_OPT="--no-headers"
PODS_NS_LIST=(openshift-ovn-kubernetes openshift-service-ca openshift-ingress openshift-dns openshift-storage)
PODS_CT_LIST=(2                        1                    1                 2             2)

# Source Greenboot configuration file if it exists
GREENBOOT_CONF_FILE=/etc/greenboot/greenboot.conf
[ -f "${GREENBOOT_CONF_FILE}" ] && source ${GREENBOOT_CONF_FILE}
WAIT_TIMEOUT_SECS_BASE=${MICROSHIFT_WAIT_TIMEOUT_SEC:-300}

# Set the exit handler to log the exit status
trap 'script_exit' EXIT

# The script exit handler logging the FAILURE or FINISHED message depending
# on the exit status of the last command
#
# args: None
# return: None
function script_exit() {
    [ "$?" -ne 0 ] && status=FAILURE || status=FINISHED
    echo $status
}

# Run a command with a second delay until it returns a zero exit status
#
# arg1: Time in seconds to wait for a command to succeed
# argN: Command to run with optional arguments
# return: 0 if a command ran successfully within the wait period, or 1 otherwise
function wait_for() {
    local timeout=$1
    shift 1

    local start=$(date +%s)
    until ("$@"); do
        sleep 1
        
        local now=$(date +%s)
        [ $(( now - start )) -ge $timeout ] && return 1
    done

    return 0
}

# Check the microshift.service systemd unit activity, terminating the script
# with the SIGTERM signal if the unit reports a failed state
#
# args: None
# return: 0 if the systemd unit is active, or 1 otherwise
function microshift_service_active() {
    local is_failed=$(systemctl is-failed microshift.service)
    local is_active=$(systemctl is-active microshift.service)

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
    local check_rd=$(${OCGET_CMD} --raw='/readyz?verbose' | awk '$2 != "ok"')
    local check_lv=$(${OCGET_CMD} --raw='/livez?verbose'  | awk '$2 != "ok"')

    [ "${check_rd}" != "readyz check passed" ] && return 1
    [ "${check_lv}" != "livez check passed"  ] && return 1
    return 0
}

# Check if any MicroShift pods are in the 'Running' status
# 
# args: None
# return: 0 if any pods are in the 'Running' status, or 1 otherwise
function any_pods_running() {
    local count=$(${OCGET_CMD} pods ${OCGET_OPT} -A 2>/dev/null | awk '$4~/Running/' | wc -l)

    [ "${count}" -gt 0 ] && return 0
    return 1
}

# Check if all the MicroShift pod images in a given namespace are downloaded.
#
# args: None
# env1: 'CHECK_PODS_NS' environment variable for the namespace to check
# return: 0 if all the images in a given namespace are downloaded, or 1 otherwise
function namespace_images_downloaded() {
    local ns=${CHECK_PODS_NS}

    local images=$(${OCGET_CMD} pods ${OCGET_OPT} -n ${ns} -o jsonpath="{.items[*].spec.containers[*].image}" 2>/dev/null)
    for i in ${images} ; do
        # Return an error on the first missing image
        local cimage=$(crictl image -q ${i})
        [ -z "${cimage}" ] && return 1
    done

    return 0
}

# Check if a given number of MicroShift pods in a given namespace are in the 'Ready' status,
# terminating the script with the SIGTERM signal if more pods are ready than expected.
#
# args: None
# env1: 'CHECK_PODS_NS' environment variable for the namespace to check
# env2: 'CHECK_PODS_CT' environment variable for the pod count to check
# return: 0 if the expected number of pods are ready, or 1 otherwise
function namespace_pods_ready() {
    local ns=${CHECK_PODS_NS}
    local ct=${CHECK_PODS_CT}

    local status=$(${OCGET_CMD} pods ${OCGET_OPT} -n ${ns} -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)
    local tcount=$(echo $status | grep -o True  | wc -l)
    local fcount=$(echo $status | grep -o False | wc -l)

    # Terminate the script in case more pods are ready than expected - nothing to wait for
    if [ "${tcount}" -gt "${ct}" ] ; then
        echo "The number of ready pods in the '${ns}' namespace is greater than the expected '${ct}' count. Terminating..."
        kill -TERM ${SCRIPT_PID}
    fi
    # Exit with error if any pods are not ready yet
    [ "${fcount}" -gt 0 ] && return 1
    # Check the ready pod count    
    [ "${tcount}" -eq "${ct}" ] && return 0
    return 1
}

# Check if MicroShift pods in a given namespace started and verify they are not restarting by sampling
# the pod restart count 10 times every 5 seconds and comparing the current sample with the previous one.
# The pods are considered restarting if the number of 'pod-restarting' samples is greater than the
# number of 'pod-not-restarting' ones.
#
# arg1: Name of the namespace to check
# return: 0 if pods are not restarting, or 1 otherwise
function namespace_pods_not_restarting() {
    local ns=$1
    local restarts=0

    local count1=$(${OCGET_CMD} pods ${OCGET_OPT} -n ${ns} -o 'jsonpath={..status.containerStatuses[].restartCount}' 2>/dev/null)
    for i in $(seq 10) ; do
        sleep 5
        local countS=$(${OCGET_CMD} pods ${OCGET_OPT} -n ${ns} -o 'jsonpath={..status.containerStatuses[].started}' 2>/dev/null | grep -vc false)
        local count2=$(${OCGET_CMD} pods ${OCGET_OPT} -n ${ns} -o 'jsonpath={..status.containerStatuses[].restartCount}' 2>/dev/null)

        # If pods started, a restart is detected by comparing the count string between the checks.
        # The number of pod restarts is incremented when a restart is detected, or decremented otherwise.
        if [ "${countS}" -ne 0 ] && [ "${count1}" = "${count2}" ] ; then
            restarts=$(( restarts - 1 ))
        else
            restarts=$(( restarts + 1 ))
            count1=${count2}
        fi
    done

    [ "${restarts}" -lt 0 ] && return 0
    return 1
}

#
# Main
#

# Exit if the current user is not 'root'
if [ $(id -u) -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 0
fi

echo "STARTED"

# Dump GRUB boot and Greenboot variables if any
grub_vars=$(grub2-editenv - list | grep ^boot_ || true)
boot_vars=$(set | egrep '^GREENBOOT_|^MICROSHIFT_' || true)

[ -z "${grub_vars}" ] && grub_vars=None
[ -z "${boot_vars}" ] && boot_vars=None

echo -e "GRUB boot variables:\n${grub_vars}"
echo -e "Greenboot variables:\n${boot_vars}"
echo -e "The ostree status:\n$(ostree admin status || true)"

# Exit if the MicroShift service is not enabled
if [ $(systemctl is-enabled microshift.service 2>/dev/null) != "enabled" ] ; then
    echo "MicroShift service is not enabled. Exiting..."
    exit 0
fi

# Update the wait timeout according to the boot counter.
# The new wait timeout is a product of the timeout base and the number of boot attempts.
MAX_BOOT_ATTEMPTS=${GREENBOOT_MAX_BOOT_ATTEMPTS:-3}
BOOT_COUNTER=$(grub2-editenv - list | grep ^boot_counter= | awk -F= '{print $2}')
[ -z "${BOOT_COUNTER}" ] && BOOT_COUNTER=$(( $MAX_BOOT_ATTEMPTS - 1 ))

WAIT_TIMEOUT_SECS=$(( $WAIT_TIMEOUT_SECS_BASE * ( $MAX_BOOT_ATTEMPTS - $BOOT_COUNTER ) ))
[ ${WAIT_TIMEOUT_SECS} -le 0 ] && WAIT_TIMEOUT_SECS=${WAIT_TIMEOUT_SECS_BASE}

# Wait for MicroShift service to be active (failed status terminates the script)
echo "Waiting ${WAIT_TIMEOUT_SECS}s for MicroShift service to be active and not failed"
wait_for ${WAIT_TIMEOUT_SECS} microshift_service_active

# Wait for MicroShift API health endpoints to be OK
echo "Waiting ${WAIT_TIMEOUT_SECS}s for MicroShift API health endpoints to be OK"
wait_for ${WAIT_TIMEOUT_SECS} microshift_health_endpoints_ok

# Wait for any pods to enter running state
echo "Waiting ${WAIT_TIMEOUT_SECS}s for any pods to be running"
wait_for ${WAIT_TIMEOUT_SECS} any_pods_running

# Wait for MicroShift core pod images to be downloaded
for i in ${!PODS_NS_LIST[@]}; do
    CHECK_PODS_NS=${PODS_NS_LIST[$i]}    

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for pod image(s) from the '${CHECK_PODS_NS}' namespace to be downloaded"
    wait_for ${WAIT_TIMEOUT_SECS} namespace_images_downloaded
done

# Wait for MicroShift core pods to enter ready state
for i in ${!PODS_NS_LIST[@]}; do
    CHECK_PODS_NS=${PODS_NS_LIST[$i]}
    CHECK_PODS_CT=${PODS_CT_LIST[$i]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for ${CHECK_PODS_CT} pod(s) from the '${CHECK_PODS_NS}' namespace to be in 'Ready' state"
    wait_for ${WAIT_TIMEOUT_SECS} namespace_pods_ready
done

# Verify that MicroShift core pods are not restarting
for i in ${!PODS_NS_LIST[@]}; do
    CHECK_PODS_NS=${PODS_NS_LIST[$i]}

    echo "Checking pod restart count in the '${CHECK_PODS_NS}' namespace"
    namespace_pods_not_restarting ${CHECK_PODS_NS}
done
