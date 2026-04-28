#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck disable=SC2034  # used elsewhere
# Increase greenboot timeout for optional packages (more services to start)
GREENBOOT_TIMEOUT=1200

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    cat <<EOF
min_vcpus=4
min_memory=4096
min_disksize=30
networks=default
boot_image=rhel98-bootc-source-optionals
fips=false
vm_greenboot_timeout="${GREENBOOT_TIMEOUT}"
EOF
}

# Currently, RHOAI is only available for x86_64
check_platform() {
    local -r record_junit=${1:-false}

    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        if "${record_junit}"; then
            record_junit "setup" "scenario_create_vms" "SKIPPED"
        fi
        exit 0
    fi
}

scenario_create_vms() {
    check_platform true

    # Increased disk size because of the additional embedded images (especially OVMS which is ~3.5GiB)
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source-optionals
    launch_vm rhel98-bootc --vm_disksize 30 --vm_vcpus 4
}

scenario_remove_vms() {
    check_platform

    remove_vm host1
}

scenario_run_tests() {
    check_platform

    run_tests host1 \
        suites/ai-model-serving/ai-model-serving-online.robot
}
