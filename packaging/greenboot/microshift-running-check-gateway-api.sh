#!/bin/bash
#
# MicroShift Gateway API-specific functionality used in Greenboot health check procedures.
#
# If 'microshift-gateway-api' RPM is installed, health check needs to include resources
# from the 'openshift-gateway-api' namespace.
#
set -eu -o pipefail

SCRIPT_NAME=$(basename "$0")
SCRIPT_PID=$$
CHECK_DEPLOY_NS="openshift-gateway-api"
LOG_POD_EVENTS=false

# Source the MicroShift health check functions library
# shellcheck source=packaging/greenboot/functions.sh
source /usr/share/microshift/functions/greenboot.sh

# Set the term handler to convert exit code to 1
trap 'forced_termination' TERM SIGINT

# Set the exit handler to log the exit status
trap 'log_script_exit' EXIT

# Handler that will be called when the script is terminated by sending TERM or
# INT signals. To override default exit codes it forces returning 1 like the
# rest of the error conditions throughout the health check.
function forced_termination() {
    echo "Signal received, terminating."
    exit 1
}

#
# Main
#

# Exit if the current user is not 'root'
if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

exit_if_fail_marker_exists

echo "STARTED"

# Print the boot variable status
print_boot_status

# Exit if the MicroShift service is not enabled
if [ "$(systemctl is-enabled microshift.service 2>/dev/null)" != "enabled" ] ; then
    echo "MicroShift service is not enabled. Exiting..."
    exit 0
fi

# Set the wait timeout for the current check based on the boot counter
WAIT_TIMEOUT_SECS=$(get_wait_timeout)

# Starting pod-specific checks
# Log list of pods and their events on failure
LOG_POD_EVENTS=true

# Wait for the Deployments to be ready
echo "Waiting ${WAIT_TIMEOUT_SECS}s for '${CHECK_DEPLOY_NS}' Deployments to be ready"
if ! wait_for "${WAIT_TIMEOUT_SECS}" namespace_deployment_ready ; then
    echo "Error: Timed out waiting for '${CHECK_DEPLOY_NS}' Deployments to be ready"
    create_fail_marker_and_exit
fi
