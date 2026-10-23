#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template cos10-bootc-source
    launch_vm centos10-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

apply_crio_config() {
    local -r helper=/tmp/crio-config.sh

    copy_file_to_vm \
        host1 \
        "${ROOTDIR}/test/assets/crio-config.sh" \
        "${helper}"
    run_command_on_vm host1 \
        "sudo bash ${helper} && rm -f ${helper} && sudo systemctl restart crio.service"
}

scenario_run_tests() {
    # TEMPORARY until OCPBUGS-129494 is resolved. The COS 9 image already
    # exposes its vendor signature configuration where CRI-O expects it.
    apply_crio_config
    run_tests host1 suites/crio-config/signed-image-pull.robot
}
