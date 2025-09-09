#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-source
    launch_vm host1 rhel94-bootc "" "" "" "" "" "" "1"
}

scenario_remove_vms() {
    remove_vm host1
}

# Skip SELinux policy validation as it is only supported
# for bootc images in later releases
scenario_run_tests() {
    run_tests host1 \
        --exclude standard-selinux-policy \
        suites/standard1
}
