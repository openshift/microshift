#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-source
    # Using centos9 is necessary for getting the latest anaconda.
    # It is a temporary workaround until rhel-9.4.iso build is available.
    launch_vm host1 centos9 "" "" "" "" 2 "" "1"

    # Open the firewall ports. Other scenarios get this behavior by
    # embedding settings in the blueprint, but there is no blueprint
    # for this scenario. We need do this step before running the RF
    # suite so that suite can assume it can reach all of the same
    # ports as for any other test.
    configure_vm_firewall host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r vmname=$(full_vm_name host1)
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        suites/network/multi-nic.robot
}
