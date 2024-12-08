#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure the host is in a good state at the start of each test. We
# could have separated them and run them as separate scenarios, but
# did not want to spend the resources on a new VM.
export TEST_RANDOMIZATION=none

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-container.ks.template ""
    launch_container --image rhel95-bootc-source
}

scenario_remove_vms() {
    remove_container
}

scenario_run_tests() {
    run_tests host1 suites/logscan/
}
