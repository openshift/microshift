#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}" --vm_vcpus 6
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift

    # Install low-latency RPM and wait for tuned reboot
    run_command_on_vm host1 "sudo dnf install -y microshift-low-latency"
    run_command_on_vm host1 "sudo systemctl restart microshift.service"

    local -r start_time=$(date +%s)
    while true; do
        boot_num=$(run_command_on_vm host1 "sudo journalctl --list-boots --quiet | wc -l" || true)
        boot_num="${boot_num%$'\r'*}"
        if [[ "${boot_num}" -ge 2 ]]; then
            break
        fi
        if [ $(( $(date +%s) - start_time )) -gt 120 ]; then
            echo "Timed out waiting for tuned reboot"
            exit 1
        fi
        sleep 5
    done

    wait_for_microshift_endpoint /readyz
    wait_for_microshift_endpoint /livez
}

scenario_run_tests() {
    run_tests host1 \
        --exitonfailure \
        suites/tuned/profile.robot \
        suites/tuned/microshift-tuned.robot \
        suites/tuned/workload-partitioning.robot \
        suites/tuned/uncore-cache.robot
}
