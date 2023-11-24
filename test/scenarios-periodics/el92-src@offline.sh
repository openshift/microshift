#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-offline.ks.template ""
    # Create a VM with embedded container images and 0 NICs
    launch_vm host1 rhel-9.2-microshift-source-isolated "" "" "" "" 0
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # FIXME: Run the tests when the MicroShift offline configuration is fixed
    echo run_tests host1 suites/standard/
}
