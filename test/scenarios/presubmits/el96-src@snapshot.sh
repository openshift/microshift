#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    cat <<EOF
min_vcpus=2
min_memory=4096
min_disksize=20
networks=
boot_image=rhel-9.6-microshift-source
fips=false
EOF
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.6-microshift-source
    launch_vm 
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/storage/snapshot.robot
}
