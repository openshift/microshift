#!/bin/sh

set -eu

# crictl redirect STDOUT.  When no objects (pod, image, container) are present, crictl dump the help menu instead.  This may be confusing to users.
sudo bash -c '
    echo "Removing crio pods"
    until crictl rmp --all --force 1>/dev/null; do sleep 1; done

    echo "Removing crio containers"
    crictl rm --all --force 1>/dev/null

    echo "Removing crio images"
    crictl rmi --all --prune 1>/dev/null

    echo "Killing conmoni, pause processes"
    pkill -9 conmon
    pkill -9 pause

    echo "Removing /var/lib/microshift"
    rm -rf /var/lib/microshift

    echo "Cleanup succeeded"
'
