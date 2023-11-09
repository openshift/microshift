#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Scenario tests upgrade from RHEL9.2 to RHEL9.3 (including MicroShift).

start_commit="rhel-9.2-microshift-4.$(previous_minor_version)"
target_commit=rhel-9.3-microshift-crel

scenario_create_vms() {
	if ! does_commit_exist "${target_commit}"; then
		echo "Commit '${target_commit}' not found in ostree repo - skipping test"
		return 0
	fi
	prepare_kickstart host1 kickstart.ks.template "${start_commit}"
	launch_vm host1
}

scenario_remove_vms() {
	remove_vm host1
}

scenario_run_tests() {
	if ! does_commit_exist "${target_commit}"; then
		echo "Commit '${target_commit}' not found in ostree repo - skipping test"
		return 0
	fi
	run_tests host1 \
		--variable "TARGET_REF:${target_commit}" \
		suites/upgrade/upgrade-successful.robot
}
