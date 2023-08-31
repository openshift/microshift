#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm host1 "rhel-9.2-microshift-4.$(previous_minor_version)"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r reponame=$(basename "${LOCAL_REPO}")
    run_tests host1 \
        --variable "TARGET_REPO_URL:${WEB_SERVER_URL}/rpm-repos/${reponame}" \
        suites/upgrade/upgrade-rpm-successful.robot
}
