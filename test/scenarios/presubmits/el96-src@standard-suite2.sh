#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.6-microshift-source-optionals
    launch_vm
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # TODO copejon: the test is killed by the scenario.sh script during the suite teardown. Increasing the timeout to 35m is a workaround.
    run_tests host1 suites/standard2/
}
