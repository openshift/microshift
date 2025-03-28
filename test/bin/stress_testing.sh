#!/bin/bash

# This script enable Stress conditions at OS level in local or remote host
# limiting resources (latency, bandwidth, packet loss, memory, disk...)

set -xuo pipefail

SCRIPT_NAME=$(basename "$0")
WONDERSHAPER="/usr/local/bin/wondershaper"
SPEEDTEST_CLI="/usr/local/bin/speedtest-cli"

# run commands
function run {
    if [ "${SSH_PORT:-}" ]; then
        SSH_PORT_PARAM="-p ${SSH_PORT}"
    fi

    if [ "${SSH_PKEY:-}" ]; then
        SSH_PKEY_PARAM="-i ${SSH_PKEY}"
    fi

    if [ "${SSH_USER:-}" ]; then
        SSH_USER_PARAM="${SSH_USER}@"
    fi

    if [ "${SSH_HOSTNAME:-}" ]; then
        # shellcheck disable=SC2086
        ssh -o StrictHostKeyChecking=no -o LogLevel=ERROR ${SSH_PORT_PARAM:-} ${SSH_PKEY_PARAM:-} ${SSH_USER_PARAM:-}${SSH_HOSTNAME} "$@"
    else
        "$@"
    fi
}

# reboot and wait
function reboot_and_wait {
    run sudo reboot &
    wait_for_vm
    wait_for_k8s
    wait_for_pods
}

function wait_for_vm {
    local -r PREV_BOOT_ID=$(run cat /proc/sys/kernel/random/boot_id)
    local -r TIMEOUT_LIMIT=$(( SECONDS + 300 ))
    while true; do
        local NEW_BOOT_ID
        NEW_BOOT_ID=$(run cat /proc/sys/kernel/random/boot_id)
        if [ "${NEW_BOOT_ID:-}" ] && [ "${PREV_BOOT_ID}" != "${NEW_BOOT_ID}" ]; then
            echo "VM booted with a different boot id"
            break
        fi
        if (( SECONDS > TIMEOUT_LIMIT )); then
            echo "ERROR: Timed out waiting for the VM to boot"
            exit 1
        fi
        sleep 10
    done
}

function wait_for_k8s {
    local -r TIMEOUT_LIMIT=$(( SECONDS + 300 ))
    while true; do
        HEALTHZ_CHECK="$(run 'sudo oc get --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig --raw=/livez')"
        if [ "${HEALTHZ_CHECK:-}" == "ok" ]; then
            echo "k8s is up and running"
            break
        fi
        if (( SECONDS > TIMEOUT_LIMIT )); then
            echo "ERROR: Timed out waiting for k8s to be ready"
            exit 1
        fi
        sleep 10
    done
}

function wait_for_pods {
    run 'sudo oc wait --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig -A --all --for=condition=Ready pods --timeout 120s'
}

# install dependecies
function install_kernel-modules-extra {
    if ! run rpm-ostree status &> /dev/null ; then
        if ! run rpm -q kernel-modules-extra &> /dev/null ; then
            run sudo dnf -y --nogpgcheck install kernel-modules-extra &> /dev/null
            reboot_and_wait
        fi
        check_command "rpm -q kernel-modules-extra"
    fi
    check_command "tc -V"
}

function install_wondershaper {
    if ! run sudo ${WONDERSHAPER} -v  &> /dev/null ; then
        run sudo curl -Lo ${WONDERSHAPER} https://raw.githubusercontent.com/magnific0/wondershaper/master/wondershaper &> /dev/null
        run sudo chmod +x ${WONDERSHAPER} &> /dev/null
    fi
    check_command "sudo ${WONDERSHAPER} -v"
}

function install_speedtest-cli {
    if ! run ${SPEEDTEST_CLI} --version &> /dev/null ; then
        run sudo curl -Lo ${SPEEDTEST_CLI} https://raw.githubusercontent.com/sivel/speedtest-cli/master/speedtest.py &> /dev/null
        run sudo chmod +x ${SPEEDTEST_CLI} &> /dev/null
    fi
    check_command "${SPEEDTEST_CLI} --version"
}

function install_stress-ng {
    if ! run stress-ng --version &> /dev/null ; then
        if ! run rpm -q stress-ng &> /dev/null ; then
            run sudo dnf -y --nogpgcheck install stress-ng &> /dev/null
        fi
    fi
    check_command "stress-ng --version"
}

# bandwidth condition
function bandwidth_condition {
    bandwidth_condition_ready
    if [ "${ACTION}" == "enable" ]; then
        enable_bandwidth
    elif [ "${ACTION}" == "disable" ]; then
        disable_bandwidth
    elif [ "${ACTION}" == "get" ]; then
        get_bandwidth
    fi
}

function bandwidth_condition_ready { 
    install_kernel-modules-extra
    install_wondershaper
    install_speedtest-cli
}

function enable_bandwidth {
    echo -e "Bandwidth condition enabled: max download and upload rate is ${VALUE} Kbps on ${INTERFACE} interface"
    run sudo ${WONDERSHAPER} -a "${INTERFACE}" -d "${VALUE}" -u "${VALUE}"
}

function disable_bandwidth {
    echo -e "Bandwidth condition disabled on ${INTERFACE} interface"
    run sudo ${WONDERSHAPER} -a "${INTERFACE}" -c || true
}

function get_bandwidth {
    run ${SPEEDTEST_CLI} --simple | tail -n2
}

# latency condition
function latency_condition {
    latency_condition_ready
    if [ "${ACTION}" == "enable" ]; then
        enable_latency
    elif [ "${ACTION}" == "disable" ]; then
        disable_latency
    elif [ "${ACTION}" == "get" ]; then
        get_latency
    fi
}

function latency_condition_ready {
    install_kernel-modules-extra
}

function enable_latency {
    echo -e "Latency condition enabled: min latency is ${VALUE} ms on ${INTERFACE} interface"
    run sudo tc qdisc replace dev "${INTERFACE}" root netem delay "${VALUE}ms"
}

function disable_latency {
    echo -e "Latency condition disabled on ${INTERFACE} interface"
    run sudo tc qdisc delete dev "${INTERFACE}" root
}

function get_latency {
    AVG_LATENCY=$(run ping -c 3 8.8.8.8 | awk '/avg/{print $4}' | awk -F"/" '{printf "%d\n",$2}')
    echo "${AVG_LATENCY}ms"
}

function tbd {
    echo -e "This action is not implemented yet"
}

function check_command {
    if ! run $1 &> /dev/null ; then
        echo -e "ERROR: $1: command not found"
        exit 1
    fi
}

# network outage
function network_outage {
    network_outage_ready
    if [ "${ACTION}" == "enable" ]; then
        enable_network_outage
    elif [ "${ACTION}" == "disable" ]; then
        disable_network_outage
    elif [ "${ACTION}" == "get" ]; then
        get_network_outage
    fi
}

function network_outage_ready {
    check_command "iptables --version"
}

function enable_network_outage {
    if [[ "$(get_network_outage)" == "Network outage is active" ]]; then
        echo -e "Network outage is already active. Exiting."
        exit 0
    fi
    run sudo iptables -I INPUT -p tcp ! --dport 22 -j DROP
    run sudo iptables -I OUTPUT -p tcp ! --sport 22 -j DROP
    echo -e "Network outage condition enabled on all interfaces except SSH (port 22)"
}

function disable_network_outage {
    if [[ "$(get_network_outage)" == "Network outage is not active" ]]; then
        echo -e "Network outage is not active. Exiting."
        exit 0
    fi
    run sudo iptables -D INPUT -p tcp ! --dport 22 -j DROP
    run sudo iptables -D OUTPUT -p tcp ! --sport 22 -j DROP
    echo -e "Network outage condition disabled"
}

function get_network_outage {
    if run sudo iptables --list-rules | grep -E "\-A (INPUT|OUTPUT) \-p.*22.*DROP"; then
        echo -e "Network outage is active"
    else
        echo -e "Network outage is not active"
    fi
}

# memory pressure condition
function memory_pressure_condition {
    memory_pressure_condition_ready
    if [ "${ACTION}" == "enable" ]; then
        enable_memory_pressure
    elif [ "${ACTION}" == "disable" ]; then
        disable_memory_pressure
    elif [ "${ACTION}" == "get" ]; then
        get_memory_pressure
    fi
}

function memory_pressure_condition_ready {
    install_stress-ng
}

function enable_memory_pressure {
    echo -e "Memory pressure condition enabled: min memory usage is ${VALUE}K"
    run stress-ng --vm 10 --vm-bytes "${VALUE}"M --verify -v
}

function disable_memory_pressure {
    echo -e "Memory pressure condition disabled"
    run sudo killall -TERM stress-ng
}

function get_memory_pressure {
    run echo "Current free memory: $(free -h | awk '/Mem/{print $3}')"
}

function tbd {
    echo -e "This action is not implemented yet"
}

function check_command {
    if ! run $1 &> /dev/null ; then
        echo -e "ERROR: $1: command not found"
        return 1
    fi
}

# set interface
function set_network_interface {
    if [ -z "${INTERFACE:-}" ]; then
        IP=$(run hostname -I | awk '{print $1}')
        INTERFACE=$(run ip route | grep default | awk -v ip="${IP}" '$0~ip{print $5}')
    fi
}

# check args
function pre_check {
    # empty or help
    if [ $# -eq 0 ] || [[ "$*" == *"--help"* ]]; then 
        help
        exit 0
    fi
}

function post_check {
    # check action is set  
    if [ -z "${ACTION}" ]; then
        echo -e "ERROR: action must be set"
        exit 1
    fi

    # check disable and get actions must not have a value
    if [ "${ACTION}" == "disable" ] || [ "${ACTION}" == "get" ] && [ ! "${VALUE:-}" == "" ]; then
        echo -e "ERROR: value param must be removed"
        exit 1
    fi
}

function check_value {
    if [ "${ACTION}" == "enable" ] && [ -z "${VALUE:-}" ]; then
        echo -e "ERROR: value param is missing"
        exit 1
    elif [ "${ACTION}" == "enable" ] && [ "${VALUE:-}" ]; then
        # check if value is a integer
        if ! [[ "${VALUE:-}" =~ ^[0-9]+$ ]]; then
            echo "ERROR: value param must be an integer" 
            exit 1
        fi
    fi 
}

# usage
function help {
    cat - <<EOF
USAGE: 
    ${SCRIPT_NAME} -e CONDITION -v value [-i interface] [-h hostname [-u ssh_user] [-p ssh_port]]
    ${SCRIPT_NAME} -d CONDITION [-i interface] [-h hostname [-u ssh_user] [-p ssh_port]]
    ${SCRIPT_NAME} -g CONDITION [-i interface] [-h hostname [-u ssh_user] [-p ssh_port]]
    ${SCRIPT_NAME} --help\n
PARAMS:
    -e,     Enable condition
    -d,     Disable condition
    -g,     Returns the current real value for the condition
    -v,     Target value when enable a condition
    -i,     Network interface for network conditions, default one is set if ommited
    -h,     Hostname to perform action in a remote host
    -u,     ssh user to perform action in a remote host
    -p,     ssh port to perform in a remote host
    -k,     ssh private key path
    --help, Shows this help\n
CONDITION: { cpu | memory | disk | bandwidth | packet_loss | latency }
    cpu: limit the usage cpu % to the value
    memory: limit the available memory to the value
    disk: limit the free space to the value in M
    bandwidth: limit the max download and upload bandwidth to the value in Kbps
    packet_loss: set packet loss % to the value
    latency: limit minimun latency to the value in ms\n
    network_outage: disable network access to the host except SSH (port 22)
EXAMPLE: 
    ${SCRIPT_NAME} --help
    ${SCRIPT_NAME} -e latency 100 -h 192.168.1.1 -u root -p 22 -k ~/key.pem
    ${SCRIPT_NAME} -d latency -h 192.168.1.1 -u root -p 22 -k ~/key.pem
    ${SCRIPT_NAME} -g latency
EOF
}

pre_check "$@"

# main loop
while getopts ":e:d:g:v:i:h:u:p:k:" param; do
    case ${param} in
        e)  ACTION="enable" CONDITION="${OPTARG}";;
        d)  ACTION="disable" CONDITION="${OPTARG}";;
        g)  ACTION="get" CONDITION="${OPTARG}";;
        v)  VALUE="${OPTARG}";;
        i)  INTERFACE="${OPTARG}";;
        h)  SSH_HOSTNAME="${OPTARG}";;
        u)  SSH_USER="${OPTARG}";;
        p)  SSH_PORT="${OPTARG}";;
        k)  SSH_PKEY="${OPTARG}";;
        *)   echo -e "ERROR: invalid option '$*'" && exit 1;;
    esac
done
shift $((OPTIND-1))

post_check

# condition loop
case ${CONDITION:-} in
    c  | cpu)            tbd ;;
    m  | memory)         check_value; memory_pressure_condition ;;
    d  | disk)           tbd ;;
    b  | bandwidth)      check_value; set_network_interface; bandwidth_condition ;;
    p  | packet_loss)    tbd ;;
    l  | latency)        check_value; set_network_interface; latency_condition;;
    no | network_outage) network_outage ;;
    "")               echo -e "ERROR: condition is missing"; exit 1;;
    *)                echo -e "ERROR: condition is not valid '${CONDITION}'"; exit 1;;
esac
