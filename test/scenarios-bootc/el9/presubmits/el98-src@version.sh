#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    cat <<EOF
min_vcpus=4
min_memory=4096
min_disksize=20
networks=
boot_image=rhel98-bootc-source
fips=false
fast=true
EOF
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source
    launch_vm rhel98-bootc --vm_vcpus 4
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.8" \
        suites/standard1/version.robot
}
