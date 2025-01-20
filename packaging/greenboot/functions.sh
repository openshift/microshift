#!/bin/bash
#
# Functions used by MicroShift in Greenboot health check procedures.
# This library may also be used for user workload health check verification.
#
SCRIPT_NAME=$(basename "$0")
SCRIPT_PID=$$

OCCONFIG_OPT="--kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig"
OCGET_OPT="--no-headers"
OCGET_CMD="oc get ${OCCONFIG_OPT}"
OCROLLOUT_CMD="oc rollout ${OCCONFIG_OPT}"

# Note about the output
# This file runs as part of a systemd unit, greenboot-healthcheck. All of the
# output is captured by journald, and in order to link it to the unit it
# belongs to, it needs to be printed in a certain way. Any foreground `echo`
# command will automatically get picked up. External commands, such as `cat`
# or running in the background requires special care. In order for journald
# to take the output as part of the unit it needs to be channeled through
# systemd-cat, which will propagate all the required configuration for it.
# To keep it runable without systemd, it is also printed to regular stdout.

# Space separated list of log file locations to be printed out in case of
# a health check failure
LOG_FAILURE_FILES=()

# Print GRUB boot, Greenboot variables and ostree status affecting the script
# behavior. This information is important for troubleshooting rollback issues.
#
# args: None
# return: Print the GRUB boot variables, /etc/greenboot/greenboot.conf settings
# and ostree status to stdout
function print_boot_status() {
    # Source Greenboot configuration file if it exists
    local conf_file=/etc/greenboot/greenboot.conf
    # shellcheck source=/dev/null
    [ -f "${conf_file}" ] && source ${conf_file}

    local grub_vars
    local boot_vars
    local ostr_stat
    grub_vars=$(grub2-editenv - list | grep ^boot_ || true)
    boot_vars=$(set | grep -E '^GREENBOOT_|^MICROSHIFT_' || true)
    ostr_stat=$(ostree admin status 2>/dev/null || true)

    [ -z "${grub_vars}" ] && grub_vars="None"
    [ -z "${boot_vars}" ] && boot_vars="None"
    [ -z "${ostr_stat}" ] && ostr_stat="Not an ostree system"

    echo -e "GRUB boot variables:\n${grub_vars}"
    echo -e "Greenboot variables:\n${boot_vars}"
    echo -e "The ostree status:\n${ostr_stat}"
}

# Get the recommended wait timeout to be used for running health check operations.
# The returned timeout is a product of a base value and a boot attempt counter, so
# that the timeout increases after every boot attempt.
#
# The base value for the timeout and the maximum boot attempts can be defined in
# the /etc/greenboot/greenboot.conf file using the MICROSHIFT_WAIT_TIMEOUT_SEC
# and GREENBOOT_MAX_BOOTS settings. These values can be in the [60..9999] range
# for MICROSHIFT_WAIT_TIMEOUT_SEC and the [1..9] range for GREENBOOT_MAX_BOOTS.
#
# args: None
# return: Print the recommended timeout value to stdout. If the values are not
# in range, errors are printed to stderr.
function get_wait_timeout() {
    # Source Greenboot configuration file if it exists
    local conf_file=/etc/greenboot/greenboot.conf
    # shellcheck source=/dev/null
    [ -f "${conf_file}" ] && source ${conf_file}

    # Read and verify the wait timeout value, allowing for the [60..9999] range
    local base_timeout=${MICROSHIFT_WAIT_TIMEOUT_SEC:-600}
    local reSecs='^[1-9]{1}[0-9]{0,3}$'
    if [[ ! ${base_timeout} =~ ${reSecs} ]] ; then
        base_timeout=600
        >&2 echo "Could not parse MICROSHIFT_WAIT_TIMEOUT_SEC value '${MICROSHIFT_WAIT_TIMEOUT_SEC}': using '${base_timeout}' instead"
    fi
    if [[ ${base_timeout} -lt 60 ]] ; then
        base_timeout=60
        >&2 echo "MICROSHIFT_WAIT_TIMEOUT_SEC value '${MICROSHIFT_WAIT_TIMEOUT_SEC}' is less than 60: using '${base_timeout}' instead"
    fi

    # Read and verify the max boots value, allowing for the [1..9] range
    local max_boots=${GREENBOOT_MAX_BOOTS:-3}
    local reBoots='^[1-9]{1}$'
    if [[ ! ${max_boots} =~ ${reBoots} ]] ; then
        max_boots=3
        >&2 echo "GREENBOOT_MAX_BOOTS value '${GREENBOOT_MAX_BOOTS}' is not in the [1..9] range: using '${max_boots}' instead"
    fi

    # Update the wait timeout according to the boot counter.
    # The new wait timeout is a product of the timeout base and the number of boot attempts.
    local boot_counter
    boot_counter=$(grub2-editenv - list | grep ^boot_counter= | awk -F= '{print $2}')
    [ -z "${boot_counter}" ] && boot_counter=$(( max_boots - 1 ))

    local wait_timeout=$(( base_timeout * ( max_boots - boot_counter ) ))
    [ ${wait_timeout} -le 0 ] && wait_timeout=${base_timeout}

    echo "${wait_timeout}"
}

# Run a command with a second delay until it returns a zero exit status
#
# arg1: Time in seconds to wait for a command to succeed
# argN: Command to run with optional arguments
# return: 0 if a command ran successfully within the wait period, or 1 otherwise
function wait_for() {
    local timeout=$1
    shift 1

    local -r start=$(date +%s)
    until ("$@"); do
        sleep 1

        local now
        now=$(date +%s)
        [ $(( now - start )) -ge "${timeout}" ] && return 1
    done

    return 0
}

# Check if all the pod images in a given namespace are downloaded.
#
# args: None
# env1: 'CHECK_PODS_NS' environment variable for the namespace to check
# return: 0 if all the images in a given namespace are downloaded, or 1 otherwise
function namespace_images_downloaded() {
    local -r ns=${CHECK_PODS_NS}

    local -r images=$(${OCGET_CMD} pods ${OCGET_OPT} -n "${ns}" -o jsonpath="{.items[*].spec.containers[*].image}" 2>/dev/null)
    for i in ${images} ; do
        # Return an error on the first missing image
        local cimage
        cimage=$(crictl image -q "${i}")
        [ -z "${cimage}" ] && return 1
    done

    return 0
}

# Check the deployment rollout status in a given namespace is successful.
#
# args: None
# env: 'CHECK_DEPLOY_NS' environment variable for the namespace to check
# return: 0 if deployment rollout is successful, or 1 otherwise
function namespace_deployment_ready() {
    local -r ns="${CHECK_DEPLOY_NS}"

    if ${OCROLLOUT_CMD} status deployment -n "${ns}" \
            --watch=false --timeout=1s &>/dev/null ; then
        return 0
    fi

    return 1
}

# Check if a given number of pods in a given namespace are in the 'Ready' status,
# terminating the script with the SIGTERM signal if more pods are ready than expected.
#
# args: None
# env1: 'CHECK_PODS_NS' environment variable for the namespace to check
# env2: 'CHECK_PODS_CT' environment variable for the pod count to check
# return: 0 if the expected number of pods are ready, or 1 otherwise
function namespace_pods_ready() {
    local ns=${CHECK_PODS_NS}
    local ct=${CHECK_PODS_CT}

    local -r status=$(${OCGET_CMD} pods ${OCGET_OPT} -n "${ns}" -o 'jsonpath={..status.conditions[?(@.type=="Ready")].status}' 2>/dev/null)
    local -r tcount=$(echo "${status}" | grep -o True  | wc -l)
    local -r fcount=$(echo "${status}" | grep -o False | wc -l)

    # Terminate the script in case more pods are ready than expected - nothing to wait for
    if [ "${tcount}" -gt "${ct}" ] ; then
        echo "The number of ready pods in the '${ns}' namespace is greater than the expected '${ct}' count. Terminating..." | tee >(systemd-cat -t "${SCRIPT_NAME}")
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

    local count1
    count1=$(${OCGET_CMD} pods ${OCGET_OPT} -n "${ns}" -o 'jsonpath={..status.containerStatuses[].restartCount}' 2>/dev/null)
    for i in $(seq 10) ; do
        sleep 5
        local countS
        local count2
        countS=$({ ${OCGET_CMD} pods ${OCGET_OPT} -n "${ns}" -o 'jsonpath={..status.containerStatuses[].started}' 2>/dev/null | grep -vc false; } || true)
        count2=$(${OCGET_CMD} pods ${OCGET_OPT} -n "${ns}" -o 'jsonpath={..status.containerStatuses[].restartCount}' 2>/dev/null)

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

# Print the contents of files from the 'LOG_FAILURE_FILES' array.
# If a file does not exist, it is skipped and an error is logged.
#
# args: None
# return: None
function print_failure_logs() {
    for file in "${LOG_FAILURE_FILES[@]}"; do
        echo "======"
        if [ -f "${file}" ]; then
            echo "Failure log in: '${file}'"
            echo "------"
            echo "$(<"${file}")"
            echo "------"
        else
            echo "Info: Log file '${file}' does not exist"
        fi
    done
}

# Run a command specified in the arguments, redirect its output to a temporary
# file and add this file to 'LOG_FAILURE_FILES' setting so that is it printed
# in the logs if the script exits with failure.
#
# All the command output including stdout and stderr is redirected to its log file.
#
# arg1: A name to be used when creating "/tmp/${name}.XXXXXXXXXX" temporary files
# arg2: A command to be run
# return: None
function log_failure_cmd() {
    local -r logName="$1"
    local -r logCmd="$2"
    local -r logFile=$(mktemp "/tmp/${logName}.XXXXXXXXXX")

    # Run the command ignoring errors and log its output
    (${logCmd}) &> "${logFile}" || true
    # Save the log file name in the list to be printed
    LOG_FAILURE_FILES+=("${logFile}")
}

# The script exit handler logging the FAILURE or FINISHED message depending
# on the exit status of the last command. If the 'LOG_POD_EVENTS' environment
# variable is set, also log the list of pods and their events on failure.
#
# args: None
# env: 'LOG_POD_EVENTS' environment variable to enable pod-specific logging
# return: None
function log_script_exit() {
    if [ "$?" -ne 0 ] ; then
        if ${LOG_POD_EVENTS}; then
            log_failure_cmd "pod-list" "${OCGET_CMD} pods -A -o wide"
            log_failure_cmd "pod-events" "${OCGET_CMD} events -A --sort-by='.metadata.creationTimestamp'"
        fi
        print_failure_logs
        echo "FAILURE"
    else
        echo "FINISHED"
    fi
}
