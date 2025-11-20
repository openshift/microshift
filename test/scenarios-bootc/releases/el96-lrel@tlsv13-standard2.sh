#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    # Apply TLS v1.3 configuration via drop-in config
    run_command_on_vm host1 "sudo mkdir -p /etc/microshift/config.d"
    run_command_on_vm host1 "sudo tee /etc/microshift/config.d/10-tls.yaml > /dev/null << 'EOF'
apiServer:
  tls:
    minVersion: VersionTLS13
EOF"

    # Restart MicroShift to apply TLS configuration
    run_command_on_vm host1 "sudo systemctl restart microshift"

    # Wait for MicroShift to be ready
    local vmname="host1"
    local -r full_vmname="$(full_vm_name "${vmname}")"
    local -r vm_ip="$(get_vm_property "${vmname}" ip)"
    if ! wait_for_greenboot "${full_vmname}" "${vm_ip}"; then
            record_junit "${vmname}" "pre_test_greenboot_check" "FAILED"
            return 1
    fi
    record_junit "${vmname}" "pre_test_greenboot_check" "OK"

    # Run standard tests excluding tls-configuration.robot since TLS v1.3 is already configured
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        --exclude tls-configuration \
        suites/standard2/ suites/selinux/validate-selinux-policy.robot
}
