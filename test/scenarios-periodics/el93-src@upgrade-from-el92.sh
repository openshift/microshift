#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Scenario tests upgrade from RHEL9.2 to RHEL9.3 without upgrading MicroShift.

start_commit=rhel-9.2-microshift-source
target_commit="rhel-9.3-microshift-source"

scenario_create_vms() {
	prepare_kickstart host1 kickstart.ks.template "${start_commit}"
	launch_vm host1 rhel-9.2
}

scenario_remove_vms() {
	remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TARGET_REF:${target_commit}" \
        suites/upgrade/upgrade-successful.robot
}
