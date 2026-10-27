#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Suites that reconfigure MicroShift without optional components, split from
# the optional1 scenario to keep both under the robot execution timeout.

# shellcheck disable=SC2034  # used elsewhere
# Increase greenboot timeout for optional packages (more services to start)
GREENBOOT_TIMEOUT=1200

scenario_create_vms() {
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart-bootc.ks.template cos10-bootc-source-optionals
    launch_vm centos10-bootc --vm_disksize 25 --vm_vcpus 4
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local skip_args=""
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        skip_args="--skip tls-scanner"
    fi
    # Skip generic device plugin on RHEL 10 until we can get the correct kernel-devel package.
    skip_args+=" --skip generic-device-plugin"
    # shellcheck disable=SC2086
    run_tests host1 \
        ${skip_args} \
        suites/optional2/
}
