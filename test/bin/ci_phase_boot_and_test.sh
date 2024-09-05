#!/bin/bash
#
# This script runs on the hypervisor to execute all scenario tests.

set -xeuo pipefail
export PS4='+ $(date "+%T.%N") ${BASH_SOURCE#$HOME/}:$LINENO \011'

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# shellcheck source=test/bin/common.sh
source "${SCRIPTDIR}/common.sh"

# Copy the scenario definition files to a temporary location from
# which they will be read. This allows filtering the tests from a
# specific $SCENARIO_SOURCES directory.
prepare_scenario_sources() {
    rm -rf "${SCENARIOS_TO_RUN}"
    mkdir -p "${SCENARIOS_TO_RUN}"
    cp "${SCENARIO_SOURCES}"/*.sh "${SCENARIOS_TO_RUN}"/
    if ${EXCLUDE_CNCF_CONFORMANCE}; then
        find "${SCENARIOS_TO_RUN}" -name "*cncf-conformance.sh" -delete
    fi
}

ENABLE_REGISTRY_MIRROR=${ENABLE_REGISTRY_MIRROR:-true}
export ENABLE_REGISTRY_MIRROR

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
bash -x ./scripts/image-builder/cleanup.sh -full

cd "${ROOTDIR}/test"

# Set up the hypervisor configuration for the tests and start webserver
bash -x ./bin/manage_hypervisor_config.sh create

# Setup a container registry and mirror images.
if ${ENABLE_REGISTRY_MIRROR}; then
    bash -x ./bin/mirror_registry.sh
fi

# Prepare all the scenarios that need to run into a special directory
prepare_scenario_sources

BOOT_TEST_JOB_LOG="${IMAGEDIR}/boot_test_jobs.txt"

cd "${TESTDIR}"

if [ ! -d "${RF_VENV}" ]; then
    "${ROOTDIR}/scripts/fetch_tools.sh" robotframework
fi

# Tell scenario.sh to merge stderr into stdout
export SCENARIO_MERGE_OUTPUT_STREAMS=true

# Show the summary of the output of the parallel jobs.
if [ -t 0 ]; then
    progress="--progress"
else
    progress=""
fi

TEST_OK=true
if ! parallel \
    ${progress} \
    --results "${SCENARIO_INFO_DIR}/{/.}/boot_and_run.log" \
    --joblog "${BOOT_TEST_JOB_LOG}" \
    --delay 5 \
    bash -x ./bin/scenario.sh create-and-run ::: "${SCENARIOS_TO_RUN}"/*.sh ; then
   TEST_OK=false
fi

cat "${BOOT_TEST_JOB_LOG}"

echo "Boot and test phase complete"
if ! "${TEST_OK}"; then
    echo "Some tests or VM boots failed"
    exit 1
fi
exit 0