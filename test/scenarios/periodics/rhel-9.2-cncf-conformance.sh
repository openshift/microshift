#!/bin/bash

# Sourced from cleanup_scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.2-microshift-source
    prepare_kickstart host2 kickstart.ks.template rhel-9.2-microshift-source
    launch_vm host1
    launch_vm host2
}

scenario_remove_vms() {
    remove_vm host1
    remove_vm host2
}
