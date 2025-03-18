#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for published MicroShift images.
# These are ec / rc / z-stream, thus guaranteed to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

scenario_create_vms() {
    local bootc_spec
    if [[ "${CURRENT_RELEASE_REPO}" == http* ]] ; then
        # Discover a pre-release MicroShift bootc image reference on the mirror
        local -r mirror_url="$(dirname "${CURRENT_RELEASE_REPO}")/bootc-pullspec.txt"

        bootc_spec="$(curl -s "${mirror_url}")"
        if [ -z "${bootc_spec}" ] || [[ "${bootc_spec}" != quay.io/openshift* ]] ; then
            echo "ERROR: Failed to retrieve a bootc pull spec from '${mirror_url}'"
            exit 1
        fi
    else
        # Use the latest released MicroShift bootc image reference in public
        # registry for the current minor version
        bootc_spec="registry.redhat.io/openshift4/microshift-bootc-rhel9:v4.${MINOR_VERSION}"
    fi

    prepare_kickstart host1 kickstart-bootc.ks.template "${bootc_spec}"
    launch_vm --boot_blueprint rhel94-bootc --bootc

    # Open the firewall ports. Other scenarios get this behavior by embedding
    # settings in the blueprint, but we cannot open firewall ports in published
    # images. We need to do this step before running the RF suite so that suite
    # can assume it can reach all of the same ports as for any other test.
    configure_vm_firewall host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    if [[ "${CURRENT_RELEASE_REPO}" == "" ]] ; then
        # TODO: While 4.19-ec is not available, it needs to exit without an error.
        exit 0
    fi
    run_tests host1 \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard2/
}
