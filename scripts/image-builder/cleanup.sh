#!/bin/bash
set -e -o pipefail

IMGNAME=microshift
ROOTDIR=$(git rev-parse --show-toplevel)/scripts/image-builder

title() {
    echo -e "\E[34m\n# $1\E[00m";
}

# Parse command line
if [ $# -ge 1 ] ; then
    case "$1" in
    -full)
        FULL_CLEAN=1
        ;;
    *)
        echo "Usage: $(basename $0) [-full]"
        exit 0
        ;;
    esac
fi

if [ "$FULL_CLEAN" = 1 ] ; then
    title "Cleaning the build directory"
    rm -rf ${ROOTDIR}/_builds
fi

title "Cleaning up local ostree container server"
sudo podman rm -f ${IMGNAME}-server 2>/dev/null || true

title "Cancelling composer jobs"
for uid in $(sudo composer-cli compose list | awk '{print $1}') ; do
    sudo composer-cli compose cancel $uid 2>/dev/null || true
    [ "$FULL_CLEAN" = 1 ] && sudo composer-cli compose delete $uid || true
done

if [ "$FULL_CLEAN" = 1 ] ; then
    title "Deleting composer jobs"
    for uid in $(sudo composer-cli compose list | awk '{print $1}') ; do
        sudo composer-cli compose delete $uid || true
    done
fi

title "Cleaning up composer sources"
sudo composer-cli sources delete openshift-local  2>/dev/null || true
sudo composer-cli sources delete microshift-local 2>/dev/null || true

title "Clean osbuild worker cache"
sudo systemctl stop osbuild-composer.socket osbuild-composer.service osbuild-worker@1.service
sleep 3
sudo rm -rf /var/cache/osbuild-worker/*
sudo systemctl start osbuild-composer.socket
