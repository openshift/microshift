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
        suites/configuration2/apiserver-readiness.robot \
        suites/configuration2/audit-log.robot \
        suites/configuration2/data-dir.robot \
        suites/configuration2/drop-in-config.robot \
        suites/configuration2/kustomize-sources.robot \
        suites/configuration2/logging.robot
}
