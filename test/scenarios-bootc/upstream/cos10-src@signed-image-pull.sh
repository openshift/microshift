#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template cos10-bootc-source
    launch_vm centos10-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

copy_signed_image_pull_helper() {
    copy_file_to_vm \
        host1 \
        "${ROOTDIR}/test/suites/container-signature/assets/temporary_container_config_workaround.py" \
        /tmp/temporary_container_config_workaround.py
}

scenario_run_tests() {
    copy_signed_image_pull_helper
    run_tests host1 suites/container-signature/signed-image-pull.robot
}
