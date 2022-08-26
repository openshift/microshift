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
    pkill -9 microshift

    echo Removing crio container and image storage
    crio wipe -f &>/dev/null || true
    systemctl restart crio

    echo Killing conmon, pause and OVN processes
    ovs-vsctl del-br br-int
    pkill -9 conmon
    pkill -9 pause
    pkill -9 ovn-controller
    pkill -9 ovn-northd
    pkill -9 ovsdb-server

    echo Removing MicroShift and OVN configuration
    rm -rf /var/lib/{microshift,ovnk}
    rm -rf /var/run/ovn
    rm -f /etc/cni/net.d/10-ovn-kubernetes.conf
    rm -f /opt/cni/bin/ovn-k8s-cni-overlay

    echo Cleanup succeeded
"
