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
    local -r kc="/var/lib/microshift/resources/kubeadmin/kubeconfig"
    run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig ${kc} wait --for=condition=Available deployment/lvms-operator -n openshift-storage --timeout=300s"
    run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig ${kc} wait --for=jsonpath='{.status.numberReady}'=1 daemonset/vg-manager -n openshift-storage --timeout=300s"
    # Wait for topolvm-node daemonset to be created by the operator and become ready
    local attempt
    for attempt in $(seq 30); do
        if run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig ${kc} get daemonset/topolvm-node -n openshift-storage 2>/dev/null" | grep -q topolvm-node; then
            break
        fi
        echo "Waiting for topolvm-node daemonset to be created (attempt ${attempt}/30)"
        sleep 10
    done
    run_command_on_vm host1 "sudo /usr/bin/oc --kubeconfig ${kc} wait --for=jsonpath='{.status.numberReady}'=1 daemonset/topolvm-node -n openshift-storage --timeout=300s"
}

scenario_run_tests() {
    run_tests host1 \
        suites/storage/
}
