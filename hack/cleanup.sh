#!/bin/sh
set -eu

echo "DATA LOSS WARNING: Do you wish to stop and cleanup ALL MicroShift data AND cri-o container workloads?"
select yn in "Yes" "No"; do
    case "${yn}" in
        Yes ) break ;;
        * ) echo "aborting cleanup; " ; exit;;
    esac
done

# crictl redirect STDOUT.  When no objects (pod, image, container) are present, crictl dump the help menu instead.  This may be confusing to users.
sudo bash -c "
    echo Stopping MicroShift
    set +e
    systemctl stop --now microshift 2>/dev/null
    systemctl disable microshift 2>/dev/null
    systemctl stop --now microshift-containerized 2>/dev/null
    systemctl disable microshift-containerized 2>/dev/null
    podman stop microshift 2>/dev/null
    podman stop microshift-aio 2>/dev/null

    echo Removing non-OVN crio pods
    
    crictl pods | tail -n +2 | grep -vE openshift-ovn-kubernetes | awk '{print \$1}' | xargs crictl stopp
    crictl pods | tail -n +2 | grep -vE openshift-ovn-kubernetes | awk '{print \$1}' | xargs crictl rmp
    

    echo Removing all crio pods
    until crictl rmp --all --force 1>/dev/null; do sleep 1; done

    echo Removing crio containers
    crictl rm --all --force 1>/dev/null

    echo Removing crio images
    crictl rmi --all --prune 1>/dev/null

    echo Killing conmon, pause processes
    pkill -9 conmon
    pkill -9 pause
    pkill -9 ovn-controller
    pkill -9 ovn-northd
    pkill -9 ovsdb-server

    echo Removing /var/lib/microshift
    crio wipe -f
    systemctl restart crio
    echo Reverting microshift-ovs-init.service configuration
    /usr/bin/configure-ovs.sh OpenShiftSDN
    rm -rf /var/lib/microshift
    rm -rf /var/lib/ovn
    rm -rf /var/run/ovn

    echo Cleanup succeeded
"