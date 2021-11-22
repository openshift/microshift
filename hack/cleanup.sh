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
sudo bash -c '
    echo "Stopping microshift"
    set +e
    systemctl stop --now microshift 2>/dev/null
    systemctl disable microshift 2>/dev/null
    systemctl stop --now microshift-containerized 2>/dev/null
    systemctl disable microshift-containerized 2>/dev/null
    podman stop microshift 2>/dev/null
    podman stop microshift-aio 2>/dev/null

    echo "Removing crio pods"
    until crictl rmp --all --force 1>/dev/null; do sleep 1; done

    echo "Removing crio containers"
    crictl rm --all --force 1>/dev/null

    echo "Removing crio images"
    crictl rmi --all --prune 1>/dev/null

    echo "Killing conmon, pause processes"
    pkill -9 conmon
    pkill -9 pause



    echo "Removing /var/lib/microshift"
    rm -rf /var/lib/microshift

    echo "Cleanup succeeded"
'
