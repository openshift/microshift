#!/bin/bash
#
# This script runs on the hypervisor, from the iso-build step.

set -euo pipefail

SCRIPTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${SCRIPTDIR}/common.sh"

# Log output automatically
LOGDIR="${ROOTDIR}/_output/ci-logs"
LOGFILE="${LOGDIR}/$(basename "$0" .sh).log"
if [ ! -d "${LOGDIR}" ]; then
    mkdir -p "${LOGDIR}"
fi
echo "Logging to ${LOGFILE}"
# Set fd 1 and 2 to write to the log file
exec &> >(tee >(awk '{ print strftime("%Y-%m-%d %H:%M:%S"), $0; fflush() }' >"${LOGFILE}"))

(cd .. && make rpm)

cd "${TESTDIR}"

./bin/ci_phase_iso_build.sh

# Start the web server to host the kickstart files and ostree commit
# repository.
./bin/start_webserver.sh

echo "Set up complete. Run the scenarios you want by hand, then run"
echo "./bin/manage_vm_connections.sh local"
