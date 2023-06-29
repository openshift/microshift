#!/bin/bash
#
# This script should be run on the hypervisor to set up port forwarding.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

usage() {
    cat - <<EOF
${BASH_SOURCE[0]} (remote|cleanup|local) [options]

  -h           Show this help.

remote: Set up iptables and connection settings for VMs for remote
        access.

  remote options;

  -a <port>    The base port number for API server port forwarding.

  -l <port>    The base port number for load balancer port forwarding for
               services running on the cluster.

  -s <port>    The base port number for ssh port forwarding.

cleanup: Remove iptables rules created by 'remote' command.

local: Set up connection settings for VMs for local access from the
       hypervisor.

EOF
}

add_iptables_rules() {
    local vm_ip="${1}"
    local api_port="${2}"
    local ssh_port="${3}"
    local lb_port="${4}"

    # Setup external access with port forwarding to allow running commands and tests from the CI container.

    # MicroShift API server
    sudo /sbin/iptables -I FORWARD -o virbr0 -p tcp -d "${vm_ip}" --dport 6443 -j ACCEPT
    sudo /sbin/iptables -t nat -I PREROUTING -p tcp --dport "${api_port}" -j DNAT --to "${vm_ip}:6443"

    # SSH to the VM
    sudo /sbin/iptables -I FORWARD -o virbr0 -p tcp -d "${vm_ip}" --dport 22 -j ACCEPT
    sudo /sbin/iptables -t nat -I PREROUTING -p tcp --dport "${ssh_port}" -j DNAT --to "${vm_ip}:22"

    # Port usable for services running on MicroShift
    sudo /sbin/iptables -I FORWARD -o virbr0 -p tcp -d "${vm_ip}" --dport 5678 -j ACCEPT
    sudo /sbin/iptables -t nat -I PREROUTING -p tcp --dport "${lb_port}" -j DNAT --to "${vm_ip}:5678"

    # Show the rules we just added
    sudo /sbin/iptables -L FORWARD | grep "${vm_ip}"
    sudo /sbin/iptables -t nat -L PREROUTING | grep "${vm_ip}"
}

delete_iptables_rules() {
    local vm_ip="${1}"

    # If there are no rules grep exits with an error, so do the search
    # with an expression that always returns true. Sort the results in
    # reverse order so we remove rules from the end of the list and do
    # not change the index values of the other rules that need to be
    # removed.

    echo "Looking for rules using ${vm_ip}"
    # shellcheck disable=SC2162  # read without -r
    (sudo /sbin/iptables -L FORWARD --line-numbers | grep "${vm_ip}" | sort -n -r || true) | while read num rule; do
        echo "Removing rule ${rule}"
        sudo /sbin/iptables -D FORWARD "${num}"
    done

    # shellcheck disable=SC2162  # read without -r
    (sudo /sbin/iptables -t nat -L PREROUTING --line-numbers | grep "${vm_ip}" | sort -n -r || true) | while read num rule; do
        echo "Removing rule ${rule}"
        sudo /sbin/iptables -t nat -D PREROUTING "${num}"
    done
}

action_remote() {

    local api_base=""
    local ssh_base=""
    local lb_base=""
    while getopts "a:hl:s:" opt; do
        case "${opt}" in
            a)
                api_base="${OPTARG}"
                ;;
            l)
                lb_base="${OPTARG}"
                ;;
            s)
                ssh_base="${OPTARG}"
                ;;
            h)
                usage
                exit 0
                ;;
            *)
                usage
                exit 1
                ;;
        esac
    done

    if [ -z "${api_base}" ]; then
        usage
        error "Specify API base port with -a option"
        exit 1
    fi
    local api_port="${api_base}"

    if [ -z "${lb_base}" ]; then
        usage
        error "Specify LB base port with -l option"
        exit 1
    fi
    local lb_port="${lb_base}"

    if [ -z "${ssh_base}" ]; then
        usage
        error "Specify ssh base port with -s option"
        exit 1
    fi
    local ssh_port="${ssh_base}"

    cd "${SCENARIO_INFO_DIR}"

    local scenario
    local vm_name
    local vm_ip
    for scenario in *; do
        pushd "${scenario}" >/dev/null
        for vm in vms/*; do
            vm_name=$(basename "${vm}")
            vm_ip=$(cat "${vm}/ip")

            echo "${scenario}: Configuring ${vm_name} at ${vm_ip} with API port ${api_port}, ssh port ${ssh_port}, and LB port ${lb_port}"
            add_iptables_rules "${vm_ip}" "${api_port}" "${ssh_port}" "${lb_port}"

            # Record the ports used for the VM
            echo "${api_port}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/api_port"
            echo "${ssh_port}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/ssh_port"
            echo "${lb_port}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/lb_port"

            # Increment the ports so they are unique
            api_port=$((api_port+1))
            ssh_port=$((ssh_port+1))
            lb_port=$((lb_port+1))
        done
        popd >/dev/null
    done
}

action_cleanup() {
    cd "${SCENARIO_INFO_DIR}"

    local scenario
    local vm_name
    local vm_ip
    for scenario in *; do
        pushd "${scenario}" >/dev/null
        for vm in vms/*; do
            vm_name=$(basename "${vm}")
            vm_ip=$(cat "${vm}/ip")

            echo "${scenario}: Cleaning up iptables rules for ${vm_name}"
            delete_iptables_rules "${vm_ip}"
        done
        popd >/dev/null
    done
}

action_local() {
    cd "${SCENARIO_INFO_DIR}"

    local api_port=6443
    local ssh_port=22
    local lb_port=5678

    local scenario
    local vm_name
    local vm_ip
    for scenario in *; do
        pushd "${scenario}" >/dev/null
        for vm in vms/*; do
            vm_name=$(basename "${vm}")
            vm_ip=$(cat "${vm}/ip")

            echo "${scenario}: Configuring ${vm_name} at ${vm_ip} with API port ${api_port} and ssh port ${ssh_port}"

            # Record the ports used for the VM
            echo "${api_port}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/api_port"
            echo "${ssh_port}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/ssh_port"
            echo "${lb_port}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/lb_port"
            echo "${vm_ip}" > "${SCENARIO_INFO_DIR}/${scenario}/vms/${vm_name}/public_ip"
        done
        popd >/dev/null
    done
}

if [ $# -eq 0 ]; then
    usage
    exit 1
fi
action="${1}"
shift

case "${action}" in
    remote|cleanup|local)
        "action_${action}" "$@"
        ;;
    -h)
        usage
        exit 0
        ;;
    *)
        usage
        exit 1
        ;;
esac
