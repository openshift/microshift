#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template "rhel-9.4-microshift-4.${PREVIOUS_MINOR_VERSION}"
    launch_vm
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
              --variable "TARGET_REF:rhel-9.6-microshift-source" \
              suites/upgrade/upgrade-successful.robot
}
