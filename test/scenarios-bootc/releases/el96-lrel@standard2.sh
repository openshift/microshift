#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    remove_vm host1
}

scenario_run_tests() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    run_tests host1 suites/standard2
}
