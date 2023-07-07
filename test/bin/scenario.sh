#!/bin/bash
#
# This script should be run on the hypervisor to manage the VMs needed
# for a scenario.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

DEFAULT_BOOT_BLUEPRINT="rhel-9.2"
LVM_SYSROOT_SIZE="10240"
WEB_SERVER_URL="http://${VM_BRIDGE_IP}:${WEB_SERVER_PORT}"
PULL_SECRET="${PULL_SECRET:-${HOME}/.pull-secret.json}"
PULL_SECRET_CONTENT="$(jq -c . "${PULL_SECRET}")"
PUBLIC_IP=${PUBLIC_IP:-""}  # may be overridden in global settings file
VM_BOOT_TIMEOUT=8m

full_vm_name() {
    local base="${1}"
    echo "${SCENARIO}-${base}"
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
}

# Show the IP address of the VM
function get_vm_ip {
    local vmname="${1}"
    sudo virsh domifaddr "${vmname}" \
        | grep vnet \
        | awk '{print $4}' \
        | cut -f1 -d/
}

# Try to login to the host via ssh until the connection is accepted
wait_for_ssh() {
    local ip="${1}"

    echo "Waiting ${VM_BOOT_TIMEOUT} for ssh access to ${ip}"
    timeout "${VM_BOOT_TIMEOUT}" bash -c "until ssh -oBatchMode=yes -oStrictHostKeyChecking=accept-new redhat@${ip} 'echo host is up'; do date; sleep 5; done"
}

# Add the microshift config file needed to allow remote access via IP address
enable_ip_access() {
    local vmname="${1}"
    local ip="${2}"

    echo "Waiting ${VM_BOOT_TIMEOUT} for ${full_vmname} to boot"
    wait_for_ssh "${ip}"

    echo "Adjusting VM host settings"
    scp "${SCRIPTDIR}/force_vm_settings.sh" "redhat@${ip}:/home/redhat/"
    ssh "redhat@${ip}" "chmod +x /home/redhat/force_vm_settings.sh && sudo /home/redhat/force_vm_settings.sh"
}

# Wait for greenboot health check to complete, without checking the results
wait_for_greenboot() {
    local vmname="${1}"
    local ip="${2}"

    echo "Waiting ${VM_BOOT_TIMEOUT} for greenboot on ${vmname} to complete"
    timeout "${VM_BOOT_TIMEOUT}" bash -c "until ssh redhat@${ip} \"sudo journalctl -n 5 -u greenboot-healthcheck; sudo systemctl status greenboot-healthcheck | grep 'active (exited)'\"; do date; sleep 10; done"
}

# Public function to start a VM.
#
# Creates a new VM using the scenario name and the vmname given to
# create a unique name. Uses the boot_blueprint argument to select the
# ISO from which to boot. If no boot_blueprint is specified, uses
# DEFAULT_BOOT_BLUEPRINT.
#
# Arguments
#  vmname -- The short name of the VM in the scenario (e.g., "host1").
#  boot_blueprint -- The image blueprint used to create the ISO that
#                    should be used to boot the VM. This is _not_
#                    necessarily the image to be installed (see
#                    prepare_kickstart).
launch_vm() {
    local vmname="$1"
    local boot_blueprint="${2:-${DEFAULT_BOOT_BLUEPRINT}}"

    local full_vmname
    local kickstart_url

    full_vmname="$(full_vm_name "${vmname}")"
    kickstart_url="${WEB_SERVER_URL}/scenario-info/${SCENARIO}/vms/${vmname}/kickstart.ks"

    # Checking web server for kickstart file
    if ! curl -o /dev/null "${kickstart_url}" >/dev/null 2>&1; then
        error "Failed to load kickstart file from ${kickstart_url}"
        exit 1
    fi

    echo "Creating ${full_vmname}"

    # FIXME: variable for vcpus?
    # FIXME: variable for memory?
    # FIXME: variable for ISO
    timeout "${VM_BOOT_TIMEOUT}" sudo virt-install \
         --noautoconsole \
         --name "${full_vmname}" \
         --vcpus 2 \
         --memory 4092 \
         --disk "path=${VM_DISK_DIR}/${full_vmname}.qcow2,size=30" \
         --network network=default,model=virtio \
         --events on_reboot=restart \
         --location "${VM_DISK_DIR}/${boot_blueprint}.iso" \
         --extra-args "inst.ks=${kickstart_url}" \
         --wait

    # Wait for an IP to be assigned
    ip=$(get_vm_ip "${full_vmname}")
    while [ -z "${ip}" ]; do
        echo "Waiting for VM ${full_vmname} to have an IP"
        sleep 30
        ip=$(get_vm_ip "${full_vmname}")
    done
    echo "VM ${full_vmname} has IP ${ip}"

    # Remove any previous key info for the host
    if [ -f "${HOME}/.ssh/known_hosts" ]; then
        echo "Clearing known_hosts entry for ${ip}"
        ssh-keygen -R "${ip}"
    fi

    # Record the IP of this VM so our caller can use it to configure
    # port forwarding and the firewall.
    mkdir -p "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}"
    echo "${ip}" > "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/ip"
    # Record the _public_ IP of the VM so the test suite can use it to
    # access the host. This is useful when the public IP is the
    # hypervisor forwarding connections. If we have no PUBLIC_IP, use
    # the VM IP and assume a local connection.
    if [ -n "${PUBLIC_IP}" ]; then
        echo "${PUBLIC_IP}" > "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/public_ip"
    else
        echo "${ip}" > "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/public_ip"
    fi

    enable_ip_access "${full_vmname}" "${ip}"
    wait_for_greenboot "${full_vmname}" "${ip}"

    echo "${full_vmname} is up and ready"
}

# Clean up the resources for one VM.
remove_vm() {
    local vmname="${1}"

    local full_vmname
    full_vmname="$(full_vm_name "${vmname}")"

    # Remove the actual VM and its storage
    if sudo virsh dumpxml "${full_vmname}" >/dev/null; then
        if ! sudo virsh dominfo "${full_vmname}" | grep '^State' | grep -q 'shut off'; then
            sudo virsh destroy "${full_vmname}"
        fi
        sudo virsh undefine "${full_vmname}"
    fi
    if sudo virsh vol-dumpxml "${full_vmname}.qcow2" vm-storage >/dev/null; then
        sudo virsh vol-delete "${full_vmname}.qcow2" vm-storage
    fi

    # Remove the info file so something processing the VMs does not
    # assume the file exists. This is most useful in a local setting.
    rm -rf "${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}"
}

# Run the tests for the current scenario
run_tests() {
    local vmname="${1}"
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


    local ssh_port_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/ssh_port"
    local api_port_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/api_port"
    local lb_port_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/lb_port"
    local public_ip_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/public_ip"
    local ip_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/ip"
    local api_port
    local ssh_port
    local public_ip
    local vm_ip
    local f
    for f in "${ssh_port_file}" "${api_port_file}" "${lb_port_file}" "${public_ip_file}" "${ip_file}"; do
        if [ ! -f "${f}" ]; then
            error "Cannot read ${f}"
            exit 1
        fi
    done
    ssh_port=$(cat "${ssh_port_file}")
    api_port=$(cat "${api_port_file}")
    lb_port=$(cat "${lb_port_file}")
    public_ip=$(cat "${public_ip_file}")
    vm_ip=$(cat "${ip_file}")

    local variable_file="${SCENARIO_INFO_DIR}/${SCENARIO}/variables.yaml"
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

    "${rf_binary}" \
        --name "${SCENARIO}" \
        --randomize all \
        --loglevel TRACE \
        --outputdir "${SCENARIO_INFO_DIR}/${SCENARIO}" \
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
    # shellcheck disable=SC1090  # cannot follow source using variable
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

    # shellcheck disable=SC1090  # cannot follow source using variable
    source "${SCENARIO_SCRIPT}"
}

action_create() {
    load_global_settings
    load_scenario_script
    scenario_create_vms
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
    local ssh_port_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/ssh_port"
    local ip_file="${SCENARIO_INFO_DIR}/${SCENARIO}/vms/${vmname}/ip"

    ssh_port=$(cat "${ssh_port_file}")
    ip=$(cat "${ip_file}")

    ssh "redhat@${ip}" -p "${ssh_port}"
}

action_run() {
    load_scenario_script
    scenario_run_tests
}

usage() {
    cat - <<EOF
scenario.sh (create|run|cleanup) scenario-script [args]

  create -- Set up the infrastructure for the test, such as VMs.

  run -- Run the scenario.

  cleanup -- Remove the VMs created for the scenario.

  login -- Login to a host for a scenario.

Settings

  The script looks for ${TESTDIR}/scenario_settings.sh for some global settings.

Login

  scenario.sh login <scenario-script> <host>
EOF
}

if [ $# -ne 2 ]; then
    usage
    exit 1
fi

action="$1"
shift
SCENARIO_SCRIPT="$1"
shift
SCENARIO=$(basename "${SCENARIO_SCRIPT}" .sh)

case "${action}" in
    create|run|cleanup|login)
        "action_${action}" "$@"
        ;;
    *)
        error "Unknown instruction ${action}"
        usage
        exit 1
esac
