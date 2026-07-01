#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

# Each optional suite restarts MicroShift with its own kustomizePaths config,
# adding ~10 minutes of restart overhead to the total execution time.
# shellcheck disable=SC2034  # used elsewhere
TEST_EXECUTION_TIMEOUT=60m

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}" --vm_disksize 25 --vm_vcpus 4
}
scenario_setup_vms() {
    rpm_configure_vm
    rpm_install_microshift

    # Install all optional RPMs needed by the compatible test suites.
    # cert-manager, gateway-api, olm are tested; the rest are installed
    # because they are pulled as dependencies or needed by the config.
    run_command_on_vm host1 "sudo dnf install -y ${MICROSHIFT_OPTIONAL_RPMS}"
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # Only run RPM-compatible optional tests.
    # NOT included (must stay on libvirt/bootc):
    #   - generic-device-plugin.robot: requires serialsim kernel module
    #   - multus.robot: requires multiple NICs on multus bridge, hypervisor L2 connectivity
    #   - observability.robot: requires external Prometheus + Loki on hypervisor
    #   - sriov.robot: requires SR-IOV physical NIC hardware
    #   - healthchecks-disabled-service.robot: requires greenboot + microshift-greenboot packages
    run_tests host1 \
        suites/optional/cert-manager.robot \
        suites/optional/gateway-api.robot \
        suites/optional/olm.robot \
        suites/optional/tls-scanner.robot
}
