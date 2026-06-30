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
    configure_microshift_mirror "${PREVIOUS_RELEASE_REPO}"
}

scenario_run_tests() {
    local -r reponame=$(basename "${LOCAL_REPO}")

    run_tests host1 \
        --exitonfailure \
        --variable "SOURCE_REPO_URL:${WEB_SERVER_URL}/rpm-repos/${reponame}" \
        --variable "TARGET_VERSION:$(local_rpm_version)" \
        --variable "PREVIOUS_MINOR_VERSION:${PREVIOUS_MINOR_VERSION}" \
        suites/rpm/upgrade-successful.robot
}
