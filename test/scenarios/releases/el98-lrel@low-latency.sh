#!/bin/bash

# DISABLED (OCPBUGS-123681 / RHEL-261119): renamed with a .disabled suffix so
# the test harness (which only picks up *.sh scenarios) skips it entirely.
#
# tuned 2.28.0-1.el9fdp regressed the [bootloader] plugin so that, on
# traditionally-deployed rpm-ostree hosts, the microshift-baseline kernel
# arguments are no longer persisted to the BLS boot entry (tuned-adm returns 0
# but /proc/cmdline is unchanged after reboot). This is a real, shipping defect:
# the low-latency profile is broken in the field for anyone on tuned 2.28.0, so
# there is nothing valid to gate PRs on until it is fixed. The composer image
# cannot be pinned to an older tuned (the edge-commit image-type package set
# requests tuned unversioned, so a blueprint version pin only produces a
# depsolve "conflicting requests" error).
#
# Re-enable by renaming this file back to el98-lrel@low-latency.sh once the fixed
# tuned build lands in el9fdp.

export SKIP_GREENBOOT=true
export TEST_RANDOMIZATION=none

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-brew-lrel-tuned"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm rhel-9.8 --vm_vcpus 6
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    # Should not be run immediately after creating VM because of
    # microshift-tuned rebooting the node to activate the profile.
    local -r start_time=$(date +%s)
    while true; do
        boot_num=$(run_command_on_vm host1 "sudo journalctl --list-boots --quiet | wc -l" || true)
        boot_num="${boot_num%$'\r'*}"
        if [[ "${boot_num}" -ge 2 ]]; then
            break
        fi
        if [ $(( $(date +%s) - start_time )) -gt 60 ]; then
            echo "Timed out waiting for VM having 2 boots"
            exit 1
        fi
        sleep 5
    done

    # --exitonfailure because tests within suites are meant to be ordered,
    # so don't advance to next test if current failed.

    run_tests host1 \
        --exitonfailure \
        suites/tuned/profile.robot \
        suites/tuned/microshift-tuned.robot \
        suites/tuned/workload-partitioning.robot \
        suites/tuned/uncore-cache.robot
}
