#!/bin/bash
#
# This script should be run on the hypervisor to manage the VMs needed
# for a scenario.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

DEFAULT_BOOT_BLUEPRINT="rhel-9.2"
LVM_SYSROOT_SIZE="10240"
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"
PULL_SECRET="${PULL_SECRET:-${HOME}/.pull-secret.json}"
PULL_SECRET_CONTENT="$(jq -c . "${PULL_SECRET}")"
PUBLIC_IP=${PUBLIC_IP:-""}  # may be overridden in global settings file
VM_BOOT_TIMEOUT=900
SKIP_SOS=${SKIP_SOS:-false}  # may be overridden in global settings file
VNC_CONSOLE=${VNC_CONSOLE:-false}  # may be overridden in global settings file

full_vm_name() {
    local base="${1}"
    echo "${SCENARIO//@/-}-${base}"
}

vm_property_filename() {
    local -r vmname="$1"
    local -r property="$2"

    echo "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/${property}"
}

get_vm_property() {
    local -r vmname="$1"
    local -r property="$2"
    local -r property_file="$(vm_property_filename "${vmname}" "${property}")"
    cat "${property_file}"
}

set_vm_property() {
    local -r vmname="$1"
    local -r property="$2"
    local -r value="$3"
    local -r property_file="$(vm_property_filename "${vmname}" "${property}")"
    mkdir -p "$(dirname "${property_file}")"
    echo "${value}" > "${property_file}"
}

sos_report() {
    if "${SKIP_SOS}"; then
        echo "Skipping sos reports"
        return
    fi
    echo "Creating sos reports"
    for vmdir in "${SCENARIO_INFO_DIR}"/"${SCENARIO}"/vms/*; do
        if [ ! -d "${vmdir}" ]; then
            # skip log files, etc.
            continue
        fi
        ip=$(cat "${vmdir}/ip")
        if [ -z "${ip}" ] ; then
            # skip hosts without NICs
            # FIXME: use virsh to copy sos report files
            continue
        fi
        # Copy the sos helper for compatibility, it is only available in 4.14 RPMs
        scp "${ROOTDIR}/scripts/microshift-sos-report.sh" "redhat@${ip}":/tmp
        ssh "redhat@${ip}" \
            "[ -f /usr/bin/microshift-sos-report ] && sudo /usr/bin/microshift-sos-report || sudo PROFILES=network,security /tmp/microshift-sos-report.sh; sudo chmod +r /tmp/sosreport*"
        mkdir -p "${vmdir}/sos"
        scp "redhat@${ip}:/tmp/sosreport*.tar.xz" "${vmdir}/sos/"
    done
}

# Public function to render a unique kickstart from a template for a
# VM in a scenario.
#
# Arguments:
#  vmname -- The short name of the VM (e.g., "host1")
#  template -- The path to the kickstart template file, relative to
#              the scenario directory.
#  boot_commit_ref -- The reference to the image that should be booted
#                     first on the host. This usually matches an image
#                     blueprint name.
prepare_kickstart() {
    local vmname="$1"
    local template="$2"
    local boot_commit_ref="$3"

    local full_vmname
    local output_file
    local vm_hostname

    full_vmname="$(full_vm_name "${vmname}")"
    output_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/kickstart.ks"
    vm_hostname="${full_vmname/./-}"

    echo "Preparing kickstart file ${template} ${output_file}"
    if [ ! -f "${KICKSTART_TEMPLATE_DIR}/${template}" ]; then
        # FIXME: Perhaps we want a default kickstart to reduce duplication?
        error "No ${template} in ${KICKSTART_TEMPLATE_DIR}"
        record_junit "${vmname}" "prepare_kickstart" "no-template"
        exit 1
    fi
    mkdir -p "$(dirname "${output_file}")"
    # shellcheck disable=SC2002   # useless cat
    cat "${KICKSTART_TEMPLATE_DIR}/${template}" \
        | sed -e "s/REPLACE_LVM_SYSROOT_SIZE/${LVM_SYSROOT_SIZE}/g" \
              -e "s|REPLACE_OSTREE_SERVER_URL|${WEB_SERVER_URL}/repo|g" \
              -e "s|REPLACE_BOOT_COMMIT_REF|${boot_commit_ref}|g" \
              -e "s|REPLACE_PULL_SECRET|${PULL_SECRET_CONTENT}|g" \
              -e "s|REPLACE_HOST_NAME|${vm_hostname}|g" \
              -e "s|REPLACE_REDHAT_AUTHORIZED_KEYS|${REDHAT_AUTHORIZED_KEYS}|g" \
              -e "s|REPLACE_PUBLIC_IP|${PUBLIC_IP}|g" \
              > "${output_file}"
    record_junit "${vmname}" "prepare_kickstart" "OK"
}

# Checks if provided commit exists in local ostree repository
does_commit_exist() {
    local -r commit="${1}"

    if ostree refs --repo "${IMAGEDIR}/repo" | grep -q "${commit}"; then
        return 0
    else
        return 1
    fi
}

# Show the IP address of the VM
function get_vm_ip {
    local -r vmname="${1}"
    local -r start=$(date +%s)
    local ip
    ip=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    while [ "${ip}" = "" ]; do
        now=$(date +%s)
        if [ $(( now - start )) -ge ${VM_BOOT_TIMEOUT} ]; then
            echo "Timed out while waiting for IP retrieval"
            exit 1
        fi
        sleep 1
        ip=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    done
    echo "${ip}"
}

# Try to login to the host via ssh until the connection is accepted
wait_for_ssh() {
    local -r ip="${1}"

    echo "Waiting ${VM_BOOT_TIMEOUT} for ssh access to ${ip}"

    local -r start_time=$(date +%s)
    while [ $(( $(date +%s) - start_time )) -lt "${VM_BOOT_TIMEOUT}" ] ; do
        if ssh -oConnectTimeout=10 -oBatchMode=yes -oStrictHostKeyChecking=accept-new "redhat@${ip}" "echo host is up" ; then
            return 0
        fi
        date
        sleep 5
    done
    # Return an error if non of the ssh attempts succeeded
    return 1
}

# Wait for greenboot health check to complete, without checking the results
wait_for_greenboot() {
    local -r vmname="${1}"
    local -r ip="${2}"

    echo "Waiting ${VM_BOOT_TIMEOUT} for greenboot on ${vmname} to complete"

    local -r start_time=$(date +%s)
    while [ $(( $(date +%s) - start_time )) -lt "${VM_BOOT_TIMEOUT}" ] ; do
        if ssh -oConnectTimeout=10 -oBatchMode=yes -oStrictHostKeyChecking=accept-new "redhat@${ip}" \
                "sudo journalctl -n 5 -u greenboot-healthcheck; \
                 systemctl show --property=SubState --value greenboot-healthcheck | grep -w exited" ; then
            return 0
        fi
        date
        sleep 10
    done
    # Return an error if non of the ssh attempts succeeded
    return 1
}

start_junit() {
    local outputfile="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/junit.xml"
    mkdir -p "$(dirname "${outputfile}")"

    echo "Creating ${outputfile}"

    cat - >"${outputfile}" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<testsuite name="infrastructure for ${SCENARIO}" timestamp="$(date --iso-8601=ns)">
EOF
}

close_junit() {
    local outputfile="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/junit.xml"

    echo '</testsuite>' >>"${outputfile}"
}

record_junit() {
    local vmname="$1"
    local step="$2"
    local results="$3"

    local outputfile="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/junit.xml"

    cat - >>"${outputfile}" <<EOF
<testcase classname="${SCENARIO} ${vmname}" name="${step}">
EOF

    case "${results}" in
        OK)
        ;;
        SKIP*)
        cat - >>"${outputfile}" <<EOF
<skipped message="${results}" type="${step}-skipped" />
EOF
        ;;
        *)
        cat - >>"${outputfile}" <<EOF
<failure message="${results}" type="${step}-failure" />
EOF
    esac

    cat - >>"${outputfile}" <<EOF
</testcase>
EOF
}


# Public function to start a VM.
#
# Creates a new VM using the scenario name and the vmname given to
# create a unique name. Uses the boot_blueprint and network_name
# arguments to select the ISO and network from which to boot.
# If no boot_blueprint is specified, uses DEFAULT_BOOT_BLUEPRINT.
# If no network_name is specified, uses the "default" network.
#
# Arguments
#  vmname -- The short name of the VM in the scenario (e.g., "host1").
#  boot_blueprint -- The image blueprint used to create the ISO that
#                    should be used to boot the VM. This is _not_
#                    necessarily the image to be installed (see
#                    prepare_kickstart).
#  network_name -- The name of the network used when creating the VM.
#  vm_vcpus -- Number of vCPUs for the VM.
#  vm_memory -- Size of RAM in MB for the VM.
#  vm_disksize -- Size of disk in GB for the VM.
#  vm_nics -- Number of network interfaces for the VM.
launch_vm() {
    local -r vmname="$1"
    local -r boot_blueprint="${2:-${DEFAULT_BOOT_BLUEPRINT}}"
    local -r network_name="${3:-default}"
    local -r vm_vcpus="${4:-2}"
    local -r vm_memory="${5:-4096}"
    local -r vm_disksize="${6:-20}"
    local -r vm_nics="${7:-1}"

    local -r full_vmname="$(full_vm_name "${vmname}")"
    local -r kickstart_url="${WEB_SERVER_URL}/scenario-info/${SCENARIO}/vms/${vmname}/kickstart.ks"

    # Checking web server for kickstart file
    if ! curl -o /dev/null "${kickstart_url}" >/dev/null 2>&1; then
        error "Failed to load kickstart file from ${kickstart_url}"
        exit 1
    fi

    echo "Creating ${full_vmname}"
    local -r vm_wait_timeout=$(( VM_BOOT_TIMEOUT / 60 ))
    local -r vm_pool_name="${VM_POOL_BASENAME}-${SCENARIO}"
    local -r vm_pool_dir="${VM_DISK_BASEDIR}/${vm_pool_name}"

    # Create the pool if it does not exist
    if [ ! -d "${vm_pool_dir}" ] ; then
        mkdir -p "${vm_pool_dir}"
    fi
    if ! sudo virsh pool-info "${vm_pool_name}" &>/dev/null; then
        sudo virsh pool-define-as "${vm_pool_name}" dir --target "${vm_pool_dir}"
        sudo virsh pool-build "${vm_pool_name}"
        sudo virsh pool-start "${vm_pool_name}"
        sudo virsh pool-autostart "${vm_pool_name}"
    fi

    # Prepare network and extra arguments for the VM creation depending on
    # the number of requested NICs
    local vm_network_args
    local vm_extra_args
    local vm_initrd_inject
    vm_network_args=""
    vm_extra_args="console=tty0 console=ttyS0,115200n8 inst.notmux"
    vm_initrd_inject=""

    for _ in $(seq "${vm_nics}") ; do
        vm_network_args+="--network network=${network_name},model=virtio "
    done
    if [ -z "${vm_network_args}" ] ; then
        vm_network_args="--network none"

        # Kickstart should be downloaded and injected into the ISO
        local -r kickstart_file=$(mktemp /tmp/kickstart.XXXXXXXX.ks)
        curl -s "${kickstart_url}" > "${kickstart_file}"

        vm_extra_args+=" inst.ks=file:/$(basename "${kickstart_file}")"
        vm_initrd_inject="--initrd-inject ${kickstart_file}"
    else
        vm_extra_args+=" inst.ks=${kickstart_url}"
    fi

    # Implement retries on VM creation until the problem is fixed
    # See https://github.com/virt-manager/virt-manager/issues/498
    local vm_created=false
    for attempt in $(seq 5) ; do
        local vm_create_start
        vm_create_start=$(date +%s)

        local graphics_args
        graphics_args="none"
        if "${VNC_CONSOLE}"; then
            graphics_args="vnc,listen=0.0.0.0"
        fi

        # When bash creates a background job (using `&`),
        # the bg job does not get its own TTY.
        # If the TTY is not provided, virt-install refuses
        # to attach to the console. `unbuffer` provides the TTY.
        # shellcheck disable=SC2086
        if ! sudo unbuffer virt-install \
            --autoconsole text \
            --graphics "${graphics_args}" \
            --name "${full_vmname}" \
            --vcpus "${vm_vcpus}" \
            --memory "${vm_memory}" \
            --disk "pool=${vm_pool_name},size=${vm_disksize}" \
            ${vm_network_args} \
            --events on_reboot=restart \
            --noreboot \
            --location "${VM_DISK_BASEDIR}/${boot_blueprint}.iso" \
            --extra-args "${vm_extra_args}" \
            ${vm_initrd_inject} \
            --wait ${vm_wait_timeout} ; then

            # Check if the command exited within 15s due to a failure
            local vm_create_end
            vm_create_end=$(date +%s)
            if [ $(( vm_create_end -  vm_create_start )) -lt 15 ] ; then
                local backoff=$(( attempt * 5 ))
                echo "Error running virt-install on attempt ${attempt}: retrying in ${backoff}s"
                sleep "${backoff}"
                continue
            fi
            # Stop retrying on timeout error
            break
        fi
        # Stop retrying when VM is created successfully
        vm_created=true
        break
    done

    if ${vm_created} ; then
        record_junit "${vmname}" "install_vm" "OK"
    else
        record_junit "${vmname}" "install_vm" "FAILED"
        return 1
    fi
    sudo virsh start "${full_vmname}"

    # If there is at least 1 NIC attached, wait for an IP to be assigned and poll for SSH access
    if  [ "${vm_nics}" -gt 0 ]; then
        # Wait for an IP to be assigned
        echo "Waiting for VM ${full_vmname} to have an IP"
        local -r ip=$(get_vm_ip "${full_vmname}")
        echo "VM ${full_vmname} has IP ${ip}"
        record_junit "${vmname}" "ip-assignment" "OK"

        # Remove any previous key info for the host
        if [ -f "${HOME}/.ssh/known_hosts" ]; then
            echo "Clearing known_hosts entry for ${ip}"
            ssh-keygen -R "${ip}"
        fi

        # Record the IP of this VM so our caller can use it to configure
        # port forwarding and the firewall.
        set_vm_property "${vmname}" "ip" "${ip}"
        # Record the _public_ IP of the VM so the test suite can use it to
        # access the host. This is useful when the public IP is the
        # hypervisor forwarding connections. If we have no PUBLIC_IP, use
        # the VM IP and assume a local connection.
        if [ -n "${PUBLIC_IP}" ]; then
            set_vm_property "${vmname}" "public_ip" "${PUBLIC_IP}"
        else
            set_vm_property "${vmname}" "public_ip" "${ip}"
            # Set the defaults for the various ports so that connections
            # from the hypervisor to the VM work.
            set_vm_property "${vmname}" "ssh_port" "22"
            set_vm_property "${vmname}" "api_port" "6443"
            set_vm_property "${vmname}" "lb_port" "5678"
        fi

        if wait_for_ssh "${ip}"; then
            record_junit "${vmname}" "ssh-access" "OK"
        else
            record_junit "${vmname}" "ssh-access" "FAILED"
            return 1
        fi
    else
        # Record no-IP for offline VMs to signal special sos report collection technique
        mkdir -p "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}"
        touch "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/ip"

        echo "VM ${full_vmname} has no NICs, skipping IP assignment and ssh polling"
        # Anything other than "OK" status is reported as an error
        record_junit "${vmname}" "ip-assignment" "OK"
        record_junit "${vmname}" "ssh-access" "OK"
    fi

    echo "${full_vmname} is up and ready"
}

# Clean up the resources for one VM.
remove_vm() {
    local -r vmname="${1}"
    local -r full_vmname="$(full_vm_name "${vmname}")"

    # Remove the actual VM
    if sudo virsh dumpxml "${full_vmname}" >/dev/null; then
        if ! sudo virsh dominfo "${full_vmname}" | grep '^State' | grep -q 'shut off'; then
            sudo virsh destroy "${full_vmname}"
        fi
        sudo virsh undefine "${full_vmname}"
    fi

    # Remove the VM storage pool
    local -r vm_pool_name="${VM_POOL_BASENAME}-${SCENARIO}"
    if sudo virsh pool-info "${vm_pool_name}" &>/dev/null; then
        sudo virsh pool-destroy "${vm_pool_name}"
        sudo virsh pool-undefine "${vm_pool_name}"
    fi

    # Remove the pool directory
    # ShellCheck: Using "${var:?}" to ensure this never expands to '/*'
    rm -rf "${VM_DISK_BASEDIR:?}/${vm_pool_name}"

    # Remove the info file so something processing the VMs does not
    # assume the file exists. This is most useful in a local setting.
    rm -rf "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}"
}

# Function to report the full current version, e.g. "4.13.5"
current_version() {
    "${SCRIPTDIR}/get_latest_rpm_version.sh"
}

# Function to report only the minor portion of the current version,
# e.g. from "4.13.5" reports "13"
current_minor_version() {
    current_version | cut -f2 -d.
}

# Function to report the *previous* minor version. If the current
# version is "4.13.5", reports "12".
previous_minor_version() {
    echo $(( $(current_minor_version) - 1 ))
}

# Function to report the *next* minor version. If the current
# version is "4.14.5", reports "15".
next_minor_version() {
    echo $(( $(current_minor_version) + 1 ))
}

# Run the tests for the current scenario
run_tests() {
    local -r vmname="${1}"
    local -r full_vmname="$(full_vm_name "${vmname}")"
    shift

    echo "Running tests with $# args" "$@"

    if [ ! -d "${RF_VENV}" ]; then
        error "RF_VENV (${RF_VENV}) does not exist, create it with: ${ROOTDIR}/scripts/fetch_tools.sh robotframework"
        exit 1
    fi
    local rf_binary="${RF_VENV}/bin/robot"
    if [ ! -f "${rf_binary}" ]; then
        error "robot is not installed to ${rf_binary}"
        exit 1
    fi

    local f
    for p in "ssh_port" "api_port" "lb_port" "public_ip" "ip"; do
        f="$(vm_property_filename "${vmname}" "${p}")"
        if [ ! -f "${f}" ]; then
            error "Cannot read ${f}"
            exit 1
        fi
    done
    local -r ssh_port=$(get_vm_property "${vmname}" "ssh_port")
    local -r api_port=$(get_vm_property "${vmname}" "api_port")
    local -r lb_port=$(get_vm_property "${vmname}" "lb_port")
    local -r public_ip=$(get_vm_property "${vmname}" "public_ip")
    local -r vm_ip=$(get_vm_property "${vmname}" "ip")

    local -r variable_file="${SCENARIO_INFO_DIR}/${SCENARIO}/variables.yaml"
    echo "Writing variables to ${variable_file}"
    mkdir -p "$(dirname "${variable_file}")"
    cat - <<EOF | tee "${variable_file}"
VM_IP: ${vm_ip}
API_PORT: ${api_port}
LB_PORT: ${lb_port}
USHIFT_HOST: ${public_ip}
USHIFT_USER: redhat
SSH_PRIV_KEY: "${SSH_PRIVATE_KEY:-}"
SSH_PORT: ${ssh_port}
EOF

    if ! wait_for_greenboot "${full_vmname}" "${vm_ip}"; then
        return 1
    fi

    "${rf_binary}" \
        --name "${SCENARIO}" \
        --randomize all \
        --loglevel TRACE \
        --outputdir "${SCENARIO_INFO_DIR}/${SCENARIO}" \
        --debugfile "${SCENARIO_INFO_DIR}/${SCENARIO}/rf-debug.log" \
        -x junit.xml \
        -V "${variable_file}" \
        "$@"
}

load_global_settings() {
    local filename="${TESTDIR}/scenario_settings.sh"
    if [ ! -f "${filename}" ]; then
        error "No ${filename}"
        exit 1
    fi

    # shellcheck source=/dev/null
    source "${filename}"

    if [ -z "${SSH_PUBLIC_KEY}" ]; then
        error "Set SSH_PUBLIC_KEY in ${filename}"
        exit 1
    fi

    REDHAT_AUTHORIZED_KEYS="$(cat "${SSH_PUBLIC_KEY}")"
}

## High-level action functions from command line arguments

load_scenario_script() {
    if [ ! -f "${SCENARIO_SCRIPT}" ]; then
        error "No scenario at ${SCENARIO_SCRIPT}"
        exit 1
    fi

    # shellcheck source=/dev/null
    source "${SCENARIO_SCRIPT}"
}

action_create() {
    start_junit
    trap close_junit RETURN

    if ! load_global_settings; then
        record_junit "setup" "load_global_settings" "FAILED"
        return 1
    fi
    record_junit "setup" "load_global_settings" "OK"

    if ! load_scenario_script; then
        record_junit "setup" "load_scenario_script" "FAILED"
        return 1
    fi
    record_junit "setup" "load_scenario_script" "OK"

    trap "sos_report" EXIT

    if ! scenario_create_vms; then
        record_junit "setup" "scenario_create_vms" "FAILED"
        return 1
    fi
    record_junit "setup" "scenario_create_vms" "OK"
}

action_cleanup() {
    load_global_settings
    load_scenario_script
    scenario_remove_vms
}

action_login() {
    load_global_settings
    local vmname
    if [ $# -eq 0 ]; then
        vmname="host1"
    else
        vmname="$1"
    fi

    ssh_port=$(get_vm_property "${vmname}" "ssh_port")
    ip=$(get_vm_property "${vmname}" "ip")

    ssh "redhat@${ip}" -p "${ssh_port}"
}

action_run() {
    load_global_settings
    load_scenario_script
    trap "sos_report" EXIT
    scenario_run_tests
}

usage() {
    cat - <<EOF
scenario.sh (create|boot|run|cleanup|rerun|login) scenario-script [args]

  create|boot -- Set up the infrastructure for the test, such as VMs.

  run -- Run the scenario.

  rerun -- cleanup, create, run for the same scenario.

  cleanup -- Remove the VMs created for the scenario.

  login -- Login to a host for a scenario.

Settings

  The script looks for ${TESTDIR}/scenario_settings.sh for some global settings.

Login

  scenario.sh login <scenario-script> [<host>]
EOF
}

if [ $# -lt 2 ]; then
    usage
    exit 1
fi

action="$1"
shift
SCENARIO_SCRIPT="$(realpath "$1")"
shift
SCENARIO=$(basename "${SCENARIO_SCRIPT}" .sh)

# Change directory to the test root
cd "${SCRIPTDIR}/.."

case "${action}" in
    create|run|cleanup|login)
        "action_${action}" "$@"
        ;;
    boot)
        action_create "$@"
        ;;
    rerun)
        action_cleanup "$@"
        action_create "$@"
        action_run "$@"
        ;;
    *)
        error "Unknown instruction ${action}"
        usage
        exit 1
esac
