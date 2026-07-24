#!/bin/bash
#
# This script runs on the hypervisor to execute all scenario tests.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"
# shellcheck source=test/bin/brew_test_alignment.sh
source "${SCRIPTDIR}/brew_test_alignment.sh"
# Directory to crawl for scenarios when creating/running in batch mode.
# The CI system will override this depending on the job its running.
SCENARIO_SOURCES="${SCENARIO_SOURCES:-${TESTDIR}/scenarios}"
# Directory where all the scenarios will be copied for execution, preserving
# the original scenario type derived from its directory name.
SCENARIOS_TO_RUN="${OUTPUTDIR}/scenarios-$(get_scenario_type_from_path "${SCENARIO_SOURCES}")"

# Copy the scenario definition files to a temporary location from
# which they will be read. This allows filtering the tests from a
# specific $SCENARIO_SOURCES directory set by CI.
prepare_scenario_sources() {
    rm -rf "${SCENARIOS_TO_RUN}"
    mkdir -p "${SCENARIOS_TO_RUN}"
    cp "${SCENARIO_SOURCES}"/*.sh "${SCENARIOS_TO_RUN}"/
    if ${EXCLUDE_CNCF_CONFORMANCE}; then
        find "${SCENARIOS_TO_RUN}" -name "*cncf-conformance.sh" -delete
    fi
}

# Log output automatically
LOGDIR="${ROOTDIR}/_output/ci-logs"
LOGFILE="${LOGDIR}/$(basename "$0" .sh).log"
if [ ! -d "${LOGDIR}" ]; then
    mkdir -p "${LOGDIR}"
fi
echo "Logging to ${LOGFILE}"
# Set fd 1 and 2 to write to the log file
exec &> >(tee >(awk '{ print strftime("%Y-%m-%d %H:%M:%S"), $0; fflush() }' >"${LOGFILE}"))

cd "${ROOTDIR}"

# Make sure libvirtd is running. We do this here, because some of the
# other scripts use virsh.
bash -x ./scripts/devenv-builder/manage-vm.sh config

# Clean up the image builder cache to free disk for virtual machines
bash -x ./scripts/devenv-builder/cleanup-composer.sh -full

cd "${ROOTDIR}/test"

# Set up the hypervisor configuration for the tests and start webserver, prometheus and loki
bash -x ./bin/manage_hypervisor_config.sh create

# Setup a container registry and mirror images.
# Release jobs need to also mirror the images from the brew RPMs.
if [[ "${SCENARIO_SOURCES:-}" =~ .*releases.* ]]; then
    bash -x ./bin/mirror_registry.sh -ri "${BREW_RPM_SOURCE}"
else
    bash -x ./bin/mirror_registry.sh
fi

# Prepare all the scenarios that need to run into an output directory
# where all the relevant scenarios will be copied for execution
prepare_scenario_sources

# For release jobs, align test suites/resources with the brew RPM source
# commit to avoid false failures from HEAD tests against older brew RPMs.
if [[ "${SCENARIO_SOURCES:-}" =~ .*releases.* ]]; then
    # shellcheck source=test/bin/common_versions.sh
    source "${SCRIPTDIR}/common_versions.sh"
    get_brew_source_commit
    if [[ -n "${BREW_SOURCE_COMMIT:-}" ]]; then
        checkout_brew_aligned_tests || exit 1
        trap 'restore_head_tests' EXIT
    fi
fi

BOOT_TEST_JOB_LOG="${IMAGEDIR}/boot_test_jobs.txt"

cd "${TESTDIR}"

if [ ! -d "${RF_VENV}" ]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" robotframework
fi
"${ROOTDIR}/scripts/fetch_tools.sh" etcdctl
if [[ "${SCENARIO_SOURCES:-}" =~ .*releases.* ]]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" ginkgo
fi

# Tell scenario.sh to merge stderr into stdout
export SCENARIO_MERGE_OUTPUT_STREAMS=true

# Show the summary of the output of the parallel jobs.
if [ -t 0 ]; then
    progress="--progress"
else
    progress=""
fi

# Each C2CC scenario runs 3 VMs (2 vCPUs / 4GiB each)
# and the c7g.metal instance has 64 cores / 128GiB.
jobs_arg=""
scenario_action="create-and-run"
if [[ "${SCENARIO_SOURCES}" =~ c2cc ]]; then
    # Limit amount of parallel scenarios for the C2CC jobs
    # to avoid over-provisioning the resources.
    jobs_arg="-j 8"

    # Power off passed VMs so the job limit is an actual cap on running VMs.
    scenario_action="create-run-shutdown"

    # Each arch runs one RHEL version to halve the scenario count.
    # The assignment rotates per commit so both combos get coverage:
    #   flip=0: x86 → el98,  ARM → el102
    #   flip=1: x86 → el102, ARM → el98
    flip=$(( 16#$(git rev-parse HEAD | cut -c1) % 2 ))
    if [[ "$(uname -m)" == "x86_64" ]]; then
        delete_ver=(el102 el98)
    else
        delete_ver=(el98 el102)
    fi
    echo "C2CC arch split: deleting ${delete_ver[flip]}-* (flip=${flip})"
    find "${SCENARIOS_TO_RUN}" -name "${delete_ver[flip]}-*.sh" -delete
elif [[ "${SCENARIO_SOURCES}" =~ .*releases.* ]]; then
    # Release scenarios have grown (e.g. the router scenarios were split
    # from 1 into 5) and no longer fit if every scenario's VMs stay up for
    # the whole job. Shut down passed scenarios' VMs as they finish so the
    # hypervisor only ever holds the still-running scenarios' VMs.
    jobs_arg="-j 20"
    scenario_action="create-run-shutdown"
fi

TEST_OK=true
# shellcheck disable=SC2086
if ! parallel \
    ${jobs_arg} \
    ${progress} \
    --results "${SCENARIO_INFO_DIR}/{/.}/boot_and_run.log" \
    --joblog "${BOOT_TEST_JOB_LOG}" \
    --delay 5 \
    bash -x ./bin/scenario.sh "${scenario_action}" ::: "${SCENARIOS_TO_RUN}"/*.sh ; then
   TEST_OK=false
fi

cat "${BOOT_TEST_JOB_LOG}"

echo "Boot and test phase complete"
if ! "${TEST_OK}"; then
    echo "Some tests or VM boots failed"
    exit 1
fi
exit 0
