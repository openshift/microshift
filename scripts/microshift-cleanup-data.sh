#!/bin/bash

set -euo pipefail

SCRIPT_NAME=$(basename "$0")
# The 'openshift-ovn-kubernetes' namespace must be in the end of the list
# to allow for pod deletion when MicroShift is stopped
PODS_NS_LIST=(openshift-service-ca openshift-ingress openshift-dns openshift-storage kube-system openshift-ovn-kubernetes)
FULL_CLEAN=false
KEEP_IMAGE=false
OVN_CLEAN=false

function usage() {
    if [ $# -gt 0 ] ; then
        echo "Error: $*"
    else
        echo "Stop all MicroShift services, also cleaning their data"
    fi
    echo ""
    echo "Usage: ${SCRIPT_NAME} <--all [--keep-images] | --ovn>"
    echo "   --all         Clean all MicroShift and OVN data"
    echo "   --keep-images Keep container images when cleaning all data"
    echo "   --ovn         Clean OVN data only"
    exit 1
}

function stop_disable_services() {
    echo Stopping MicroShift services
    for service in microshift microshift-etcd ; do
        systemctl stop --now   ${service} 2>/dev/null || true
        systemctl reset-failed ${service} 2>/dev/null || true
    done

    if ${FULL_CLEAN} ; then
        echo Disabling MicroShift services
        systemctl disable microshift 2>/dev/null
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
    echo Removing MicroShift pods
    for i in "${!PODS_NS_LIST[@]}"; do
        local ns
        ns=${PODS_NS_LIST[${i}]}
        retries=5
        while [ ${retries} -gt 0 ] ; do
            local ocp_pods
            ocp_pods=$(crictl pods --namespace "${ns}" -q)
            if [ "$(echo "${ocp_pods}" | wc -w)" -eq 0 ] ; then
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

    echo Killing conmon, pause and OVN processes
    for pname in conmon pause ovn-controller ovn-northd ovsdb-server ; do
        pkill -9 --exact ${pname} || true
    done
}

function clean_data() {
    if ${FULL_CLEAN} ; then
        echo Removing MicroShift configuration
        rm -rf /var/lib/microshift
    fi

    echo Removing OVN configuration
    rm -rf /var/lib/ovnk
    rm -rf /var/run/ovn
    rm -f /etc/cni/net.d/10-ovn-kubernetes.conf
    rm -f /opt/cni/bin/ovn-k8s-cni-overlay
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
    *)
        usage
        ;;
    esac
    shift
done

# Verify valid option combination
! ${FULL_CLEAN} && ! ${OVN_CLEAN}  && usage "Either --all or --ovn option must be specified"
! ${FULL_CLEAN} &&   ${KEEP_IMAGE} && usage "The --keep-images option can only be used with --all"
  ${FULL_CLEAN} &&   ${OVN_CLEAN}  && usage "The --all and --ovn options are mutually exclusive"

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

echo Cleanup succeeded
