#!/bin/bash

set -euo pipefail

SCRIPT_NAME=$(basename "$0")
FULL_CLEAN=false
KEEP_IMAGE=false
OVN_CLEAN=false
CERT_CLEAN=false
# Flags used for generating a user summary message
SERVICE_STOPPED=false
SERVICE_DISABLED=false

function usage() {
    if [ $# -gt 0 ] ; then
        echo "Error: $*"
    else
        echo "Stop all MicroShift services, also cleaning their data"
    fi
    echo ""
    echo "Usage: ${SCRIPT_NAME} <--all [--keep-images] | --ovn | --cert>"
    echo "   --all         Clean all MicroShift and OVN data"
    echo "   --keep-images Keep container images when cleaning all data"
    echo "   --ovn         Clean OVN data only"
    echo "   --cert        Clean certificates only"
    exit 1
}

function stop_disable_services() {
    echo Stopping MicroShift services
    for service in microshift microshift-etcd ; do
        systemctl stop --now   ${service} 2>/dev/null || true
        systemctl reset-failed ${service} 2>/dev/null || true
        SERVICE_STOPPED=true
    done

    if ${FULL_CLEAN} ; then
        echo Disabling MicroShift services
        systemctl disable microshift 2>/dev/null || true
        SERVICE_DISABLED=true
    fi

    # Killing the processes is the last resort after stopping the services with the 'systemctl stop' command.
    # Usually, this is only necessary in development environment.
    pkill -9 --exact microshift      || true
    pkill -9 --exact microshift-etcd || true
}

function stop_clean_pods() {
    # It is necessary to remove the pods (OVN-related last) to allow for further termination
    # of processes (i.e. conmon, etc.) that use the files under /var/run/ovn.
    # The cleanup of OVN data only works if the files under /var/run/ovn are not in use.
    if [ ! -e /var/run/crio/crio.sock ]; then
        echo "crio.sock is not present, not attempting to clean up pods"
        return 0
    fi

    echo Removing MicroShift pods
    # The 'openshift-ovn-kubernetes' namespace must be in the end of the list
    # to allow for pod deletion when MicroShift is stopped
    local namespaces
    namespaces=$(crictl pods --output json | jq -r '.items[].metadata.namespace' | sort -u)
    if [ -n "${namespaces}" ] ; then
        namespaces=$(grep -vw openshift-ovn-kubernetes <<<"${namespaces}" || true)
        namespaces+=" openshift-ovn-kubernetes"
    fi

    for ns in ${namespaces} ; do
        retries=5
        while [ ${retries} -gt 0 ] ; do
            local ocp_pods
            ocp_pods=$(crictl pods --namespace "${ns}" -q)
            if [ -z "${ocp_pods}" ] ; then
                break
            fi
            # shellcheck disable=SC2086
            crictl rmp -f ${ocp_pods} &>/dev/null || true
            retries=$(( retries - 1 ))
            sleep 1
        done
    done

    # When full clean is requested, remove image storage
    # after the pods are shut down
    if ${FULL_CLEAN} && ! ${KEEP_IMAGE} ; then
        echo Removing crio image storage
        crictl rmi -a &>/dev/null || true
    fi
}

function clean_processes() {
    if ${FULL_CLEAN} ; then
        # This operation can only be run in the full clean mode as the removal
        # of host interfaces is not detected by kubelet and pods cannot be
        # recovered following the service restart
        if ovs-vsctl list-ifaces br-int &>/dev/null ; then
            echo Deleting the 'br-int' interface
            ovs-vsctl del-br br-int
        else
            echo The 'br-int' bridge interface does not exist
        fi
    fi

    if ${FULL_CLEAN} || ${OVN_CLEAN} ; then
        echo Killing conmon, pause and OVN processes
        systemctl stop --now ovsdb-server.service 2>/dev/null || true
        for pname in conmon pause ovn-controller ovn-northd ; do
            pkill -9 --exact ${pname} || true
        done
    fi
}

function clean_data() {
    if ${FULL_CLEAN} ; then
        echo Removing MicroShift configuration
        rm -rf /var/lib/microshift
        # Best effort - it contains Pod volumes which sometimes cannot be removed (Device or resource busy)
        rm -rf /var/lib/kubelet 2>/dev/null || true
    elif ${CERT_CLEAN} ; then
        echo Removing MicroShift certificates
        rm -rf /var/lib/microshift/certs
    fi

    if ${FULL_CLEAN} || ${OVN_CLEAN} ; then
        echo Removing OVN configuration
        rm -rf /var/run/ovn
        rm -rf /var/run/ovn-kubernetes
        rm -f /etc/cni/net.d/10-ovn-kubernetes.conf
        rm -f /run/cni/bin/ovn-k8s-cni-overlay
    fi
}

function report_status() {
    if ${SERVICE_STOPPED} ; then
        echo "MicroShift service was stopped"
    fi
    if ${SERVICE_DISABLED} ; then
        echo "MicroShift service was disabled"
    fi
}

# Parse command line
[ $# -lt 1 ] && usage

while [ $# -gt 0 ] ; do
    case $1 in
    --all)
        FULL_CLEAN=true
        ;;
    --keep-images)
        KEEP_IMAGE=true
        ;;
    --ovn)
        OVN_CLEAN=true
        ;;
    --cert)
        CERT_CLEAN=true
        ;;
    *)
        usage
        ;;
    esac
    shift
done

# Verify valid option combination
! ${FULL_CLEAN} && ! ${OVN_CLEAN}  && ! ${CERT_CLEAN} && usage "Either --all, --ovn or --cert option must be specified"
! ${FULL_CLEAN} &&   ${KEEP_IMAGE}                    && usage "The --keep-images option can only be used with --all"
  ${FULL_CLEAN} &&   ${OVN_CLEAN}                     && usage "The --all and --ovn options are mutually exclusive"
  ${FULL_CLEAN} &&   ${CERT_CLEAN}                    && usage "The --all and --cert options are mutually exclusive"
  ${OVN_CLEAN}  &&   ${CERT_CLEAN}                    && usage "The --ovn and --cert options are mutually exclusive"

# Exit if the current user is not 'root'
if [ "$(id -u)" -ne 0 ] ; then
    echo "The '${SCRIPT_NAME}' script must be run with the 'root' user privileges"
    exit 1
fi

if ${FULL_CLEAN} ; then
    read -r -p \
        $'DATA LOSS WARNING: Do you wish to stop and clean ALL MicroShift data AND cri-o container workloads?\n1) Yes\n2) No\n#? ' \
        answer
    case "${answer,,}" in
        yes|y|1)
            ;;
        *)
            echo "Aborting cleanup"
            exit 0
            ;;
    esac
fi

stop_disable_services
stop_clean_pods
clean_processes
clean_data

report_status
echo Cleanup succeeded
