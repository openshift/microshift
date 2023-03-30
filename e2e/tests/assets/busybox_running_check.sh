#!/bin/bash
set -e

SCRIPT_NAME=$(basename $0)
PODS_NS_LIST=(busybox)
PODS_CT_LIST=(1      )

# Source the MicroShift health check functions library
source /usr/share/microshift/functions/greenboot.sh

# Set the exit handler to log the exit status
trap 'script_exit' EXIT

# The script exit handler logging the FAILURE or FINISHED message depending
# on the exit status of the last command
#
# args: None
# return: None
function script_exit() {
    [ "$?" -ne 0 ] && status=FAILURE || status=FINISHED
    echo $status
}

#
# Main
#

# Exit if the current user is not 'root'
if [ $(id -u) -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

echo "STARTED"

# Exit if the MicroShift service is not enabled
if [ $(systemctl is-enabled microshift.service 2>/dev/null) != "enabled" ] ; then
    echo "MicroShift service is not enabled. Exiting..."
    exit 0
fi

# Set the wait timeout for the current check based on the boot counter
WAIT_TIMEOUT_SECS=$(get_wait_timeout)

# Wait for pod images to be downloaded
for i in ${!PODS_NS_LIST[@]}; do
    CHECK_PODS_NS=${PODS_NS_LIST[$i]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for pod image(s) from the '${CHECK_PODS_NS}' namespace to be downloaded"
    wait_for ${WAIT_TIMEOUT_SECS} namespace_images_downloaded
done

# Wait for pods to enter ready state
for i in ${!PODS_NS_LIST[@]}; do
    CHECK_PODS_NS=${PODS_NS_LIST[$i]}
    CHECK_PODS_CT=${PODS_CT_LIST[$i]}

    echo "Waiting ${WAIT_TIMEOUT_SECS}s for ${CHECK_PODS_CT} pod(s) from the '${CHECK_PODS_NS}' namespace to be in 'Ready' state"
    wait_for ${WAIT_TIMEOUT_SECS} namespace_pods_ready
done

# Verify that pods are not restarting
for i in ${!PODS_NS_LIST[@]}; do
    CHECK_PODS_NS=${PODS_NS_LIST[$i]}

    echo "Checking pod restart count in the '${CHECK_PODS_NS}' namespace"
    namespace_pods_not_restarting ${CHECK_PODS_NS}
done
