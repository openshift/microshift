#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.8-microshift-source
    launch_vm rhel-9.8
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        suites/configuration/tls-configuration.robot \
        suites/configuration/drop-in-config.robot \
        suites/configuration/show-config.robot \
        suites/configuration/logging.robot \
        suites/configuration/data-dir.robot \
        suites/configuration/apiserver-readiness.robot \
        suites/configuration/audit-log.robot
}
