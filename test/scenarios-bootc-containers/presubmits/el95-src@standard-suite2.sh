#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-container.ks.template ""
    launch_container --image rhel95-bootc-source
}

scenario_remove_vms() {
    remove_container
}

scenario_run_tests() {
    run_tests host1 \
        suites/standard2/etcd.robot \
        suites/standard2/kustomize.robot \
        suites/standard2/validate-custom-certificates.robot

    # suites/standard2/hostname.robot - cannot change hostname inside a container
    # suites/standard2/validate-certificate-rotation.robot - gets stuck on `Sleep 5s` in Teardown
}
