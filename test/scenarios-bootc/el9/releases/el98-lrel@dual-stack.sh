#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-bootc-brew-lrel-optional"

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    cat <<EOF
min_vcpus=4
min_memory=4096
min_disksize=20
networks=${VM_DUAL_STACK_NETWORK}
boot_image=${start_image}
fips=false
EOF
}

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm rhel98-bootc --network "${VM_DUAL_STACK_NETWORK}" --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"
    run_tests host1 suites/ipv6/dualstack.robot
}
