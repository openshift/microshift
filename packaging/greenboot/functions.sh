#!/bin/bash
#
# Functions used by MicroShift in Greenboot health check procedures.
#

MICROSHIFT_GREENBOOT_FAIL_MARKER=/run/microshift-greenboot-healthcheck-failed

# Note about the output
# This file runs as part of a systemd unit, greenboot-healthcheck. All of the
# output is captured by journald, and in order to link it to the unit it
# belongs to, it needs to be printed in a certain way. Any foreground `echo`
# command will automatically get picked up. External commands, such as `cat`
# or running in the background requires special care. In order for journald
# to take the output as part of the unit it needs to be channeled through
# systemd-cat, which will propagate all the required configuration for it.
# To keep it runable without systemd, it is also printed to regular stdout.

# Print GRUB boot, Greenboot variables and ostree / bootc status affecting the
# script behavior. This information is important for troubleshooting rollback
# issues.
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
    grub_vars=$(grub2-editenv - list | grep ^boot_ || true)
    boot_vars=$(set | grep -E '^GREENBOOT_|^MICROSHIFT_' || true)

    [ -z "${grub_vars}" ] && grub_vars="None"
    [ -z "${boot_vars}" ] && boot_vars="None"

    # Assume RPM system installation type
    local system_type
    local system_stat
    system_type="RPM"
    system_stat="Not an ostree / bootc system"

    # Check ostree and bootc status. Note that on bootc systems, ostree status
    # command may also return valid output, so it needs to be overriden by the
    # bootc status command output.
    if which ostree &>/dev/null ; then
        local -r ostree_stat=$(ostree admin status 2>/dev/null || true)
        if [ -n "${ostree_stat}" ] ; then
            system_type="ostree"
            system_stat="${ostree_stat}"
        fi
    fi
    if which bootc &>/dev/null ; then
        local -r bootc_stat=$(bootc status --booted --json 2>/dev/null | jq -r .status.type || true)
        if [ "${bootc_stat}" == "bootcHost" ] ; then
            system_type="bootc"
            system_stat="${bootc_stat}"
        fi
    fi

    echo -e "GRUB boot variables:\n${grub_vars}"
    echo -e "Greenboot variables:\n${boot_vars}"
    echo -e "System installation type:\n${system_type}"
    echo -e "System installation status:\n${system_stat}"
}

# Get the recommended wait timeout to be used for running health check operations.
# The returned timeout is a product of a base value and a boot attempt counter, so
# that the timeout increases after every boot attempt.
#
# The base value for the timeout and the maximum boot attempts can be defined in the
# /etc/greenboot/greenboot.conf file using the MICROSHIFT_WAIT_TIMEOUT_SEC and
# GREENBOOT_MAX_BOOT_ATTEMPTS settings. These values can be in the [60..9999] range
# for MICROSHIFT_WAIT_TIMEOUT_SEC and the [1..9] range for GREENBOOT_MAX_BOOT_ATTEMPTS.
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
    local max_boots=${GREENBOOT_MAX_BOOT_ATTEMPTS:-3}
    local reBoots='^[1-9]{1}$'
    if [[ ! ${max_boots} =~ ${reBoots} ]] ; then
        max_boots=3
        >&2 echo "GREENBOOT_MAX_BOOT_ATTEMPTS value '${GREENBOOT_MAX_BOOT_ATTEMPTS}' is not in the [1..9] range: using '${max_boots}' instead"
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

# Exit if previous MicroShift healthcheck scripts failed.
function exit_if_fail_marker_exists() {
    if [ -f "${MICROSHIFT_GREENBOOT_FAIL_MARKER}" ]; then
        >&2 echo "'${MICROSHIFT_GREENBOOT_FAIL_MARKER}' file exists, exiting with an error"
        exit 1
    fi
}

# Create fail marker and exit
function create_fail_marker_and_exit() {
    >&2 echo "Creating '${MICROSHIFT_GREENBOOT_FAIL_MARKER}' file and exiting with an error."
    touch "${MICROSHIFT_GREENBOOT_FAIL_MARKER}"
    exit 1
}

# Clear fail marker
function clear_fail_marker() {
    if [ -f "${MICROSHIFT_GREENBOOT_FAIL_MARKER}" ]; then
        >&2 echo "'${MICROSHIFT_GREENBOOT_FAIL_MARKER}' file exists - removing"
        rm -f "${MICROSHIFT_GREENBOOT_FAIL_MARKER}"
    fi
}
