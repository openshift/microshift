#!/bin/bash
# shellcheck source=test/scenarios-rpm/common-scenarios-rpm.sh
source "${TESTDIR}/scenarios-rpm/common-scenarios-rpm.sh"

scenario_create_vms() {
    prepare_kickstart host1 kickstart-liveimg.ks.template ""
    launch_vm "${RPM_INSTALLER_IMAGE}"
    configure_vm_firewall host1
    subscription_manager_register host1
    configure_rpm_repos
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r reponame=$(basename "${LOCAL_REPO}")
    install_microshift "${WEB_SERVER_URL}/rpm-repos/${reponame}" "$(local_rpm_version)"

    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:${RPM_RHEL_VERSION}" \
        suites/standard1/
}
