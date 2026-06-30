#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift

    # Wait for LVMS pods — greenboot normally handles this, but SKIP_GREENBOOT=true in RPM mode
    run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig /var/lib/microshift/resources/kubeadmin/kubeconfig wait --for=condition=Available deployment/lvms-operator -n openshift-storage --timeout=300s"
}

scenario_run_tests() {
    run_tests host1 \
        suites/storage/
}
