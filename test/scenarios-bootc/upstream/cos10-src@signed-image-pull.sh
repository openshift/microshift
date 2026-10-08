#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template cos10-bootc-source
    launch_vm centos10-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

apply_temporary_signature_config_workaround() {
    local -r helper=/tmp/temporary_container_signature_config_workaround.sh

    copy_file_to_vm \
        host1 \
        "${ROOTDIR}/test/suites/container-signature/assets/temporary_container_signature_config_workaround.sh" \
        "${helper}"
    run_command_on_vm host1 \
        "sudo bash ${helper} && rm -f ${helper} && sudo systemctl restart crio.service"
}

scenario_run_tests() {
    # TEMPORARY until OCPBUGS-129494 is resolved. The COS 9 image already
    # exposes its vendor signature configuration where CRI-O expects it.
    apply_temporary_signature_config_workaround
    run_tests host1 suites/container-signature/signed-image-pull.robot
}
