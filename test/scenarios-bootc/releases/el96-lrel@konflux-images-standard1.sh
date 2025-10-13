#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for published MicroShift images.
# These are ec / rc / z-stream, thus guaranteed to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

scenario_create_vms() {
    local -r bootc_spec="$(curl -s "${KONFLUX_LREL_RELEASE_IMAGE_URL}")"
    if [ -z "${bootc_spec}" ] || [[ "${bootc_spec}" != quay.io/openshift* ]] ; then
        echo "ERROR: Failed to retrieve a bootc pull spec from '${KONFLUX_LREL_RELEASE_IMAGE_URL}'"
        exit 1
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template "${bootc_spec}"
    launch_vm --boot_blueprint rhel96-bootc

    # Open the firewall ports. Other scenarios get this behavior by embedding
    # settings in the blueprint, but we cannot open firewall ports in published
    # images. We need to do this step before running the RF suite so that suite
    # can assume it can reach all of the same ports as for any other test.
    configure_vm_firewall host1
}

scenario_remove_vms() {
    does_vm_exists host1

    remove_vm host1
}

scenario_run_tests() {
    does_vm_exists host1

    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
