#!/bin/bash

# Sourced from cleanup_scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template ""
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    install_microshift_localrepo host1 "${YMINUS1_REPO}"
    uninstall_microshift_localrepo host1

    install_microshift_localrepo host1 "${LOCAL_REPO}"
    uninstall_microshift_localrepo host1
}
