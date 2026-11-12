#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck disable=SC2034  # used elsewhere
TEST_EXECUTION_TIMEOUT=45m

# shellcheck disable=SC2034  # used elsewhere
# Increase greenboot timeout for optional packages (more services to start)
GREENBOOT_TIMEOUT=1200

scenario_create_vms() {
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart.ks.template rhel-9.8-microshift-source-optionals
    launch_vm rhel-9.8 --vm_disksize 25 --vm_vcpus 4
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local skip_args=""
    if [[ "${UNAME_M}" =~ aarch64 ]]; then
        skip_args="--skip tls-scanner"
    fi
    # Generic Device Plugin suite is excluded because getting serialsim for ostree would require:
    # - getting the version of the kernel of ostree image,
    # - installing kernel-devel of that version on the hypervisor,
    # - building serialsim
    # - packaging serialsim as an RPM
    # - including the RPM in the ostree blueprint
    # GDP suite is tested with bootc images instead.
    #
    # shellcheck disable=SC2086
    run_tests host1 \
        --exclude generic-device-plugin \
        ${skip_args} \
        suites/optional2/
}
