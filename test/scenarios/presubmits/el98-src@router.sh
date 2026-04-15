#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel-9.8-microshift-source"

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    cat <<EOF
boot_image=${start_image}
slow=true
EOF
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm rhel-9.8
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/router/router.robot
}
