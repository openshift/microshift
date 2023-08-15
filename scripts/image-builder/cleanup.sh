#!/bin/bash
set -e -o pipefail

ROOTDIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../../" && pwd )"
BUILDDIR="${ROOTDIR}/_output/image-builder/"

title() {
    echo -e "\E[34m\n# $1\E[00m";
}

clean_podman_images() {
    if ! which podman &>/dev/null ; then
        return
    fi

    title "Cleaning up container images"
    for id in $(sudo podman ps -a | grep microshift | awk '{print $1}') ; do
        sudo podman rm -f "${id}"
    done

    if [ "${FULL_CLEAN}" = 1 ] ; then
        sudo podman rmi -af
    fi
}

clean_composer_jobs() {
    if ! which composer-cli &>/dev/null ; then
        return
    fi

    title "Cancelling composer jobs"
    for id in $(sudo composer-cli compose list | grep -v "^ID"  | grep -E "WAITING|RUNNING" | cut -f1 -d' '); do
        echo "Cancelling ${id}"
        sudo composer-cli compose cancel "${id}"
    done

    if [ "${FULL_CLEAN}" = 1 ] ; then
        title "Deleting composer jobs"
        for id in $(sudo composer-cli compose list | grep -v "^ID" | cut -f1 -d' '); do
            echo "Deleting ${id}"
            sudo composer-cli compose delete "${id}"
        done

        title "Cleaning up composer sources"
        for src in $(sudo composer-cli sources list | grep -Ev "appstream|baseos") ; do
            echo "Removing source ${src}"
            sudo composer-cli sources delete "${src}"
        done
    fi
}

clean_osbuilder_services() {
    if ! sudo systemctl is-enabled osbuild-composer.socket &>/dev/null ; then
        return
    fi

    title "Stopping osbuild services"
    sudo systemctl stop --now osbuild-composer.socket osbuild-composer.service
    for n in $(seq 100) ; do
        worker=osbuild-worker@${n}.service
        if sudo systemctl status "${worker}" &>/dev/null ; then
            sudo systemctl stop --now "osbuild-worker@${n}.service"
        else
            break
        fi
    done

    title "Cleaning osbuild worker cache"
    sleep 5
    sudo rm -rf /var/cache/osbuild-worker/* /var/lib/osbuild-composer/*

    title "Starting osbuild services"
    sudo systemctl start osbuild-composer.socket
    sudo systemctl start osbuild-worker@1.service
}

# Parse command line
if [ $# -ge 1 ] ; then
    case "$1" in
    -full)
        FULL_CLEAN=1
        ;;
    *)
        echo "Usage: $(basename "$0") [-full]"
        exit 0
        ;;
    esac
fi

clean_podman_images
clean_composer_jobs
clean_osbuilder_services

if [ "${FULL_CLEAN}" = 1 ] ; then
    title "Cleaning the build directory"
    rm -rf "${BUILDDIR}"

    title "Cleaning up user cache"
    rm -rf ~/.cache 2>/dev/null || true
    sudo rm -rf /tmp/containers/* 2>/dev/null || true
fi
