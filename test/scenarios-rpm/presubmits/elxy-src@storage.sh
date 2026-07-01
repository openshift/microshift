#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}"
}
scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift

    # Wait for LVMS operator and vg-manager to be ready before running storage tests.
    local -r kc="/var/lib/microshift/resources/kubeadmin/kubeconfig"
    run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig ${kc} wait --for=condition=Available deployment/lvms-operator -n openshift-storage --timeout=300s"
    run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig ${kc} wait --for=jsonpath='{.status.numberReady}'=1 daemonset/vg-manager -n openshift-storage --timeout=300s"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # Only run RPM-compatible storage tests.
    # Excluded: reboot.robot (needs greenboot), snapshot.robot (needs LVM thin pool setup).
    run_tests host1 \
        suites/storage/pvc-resize.robot \
        suites/storage/storage-version-migration.robot
}
