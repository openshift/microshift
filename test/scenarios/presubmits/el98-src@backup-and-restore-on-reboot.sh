#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.


# NOTE: prerun-data-management.robot is destructive (deletes MicroShift
# data, triggers greenboot rollback loops) and must run last.
export TEST_RANDOMIZATION=none

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.8-microshift-source
    launch_vm rhel-9.8
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        suites/backup/backup-restore-on-reboot.robot \
        suites/backup/prerun-data-management.robot
}
