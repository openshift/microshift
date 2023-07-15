#!/bin/bash
#
# This script cleans up osbuild-composer. It cancels any running
# builds, deletes failed and completed builds, and removes package
# sources other than the defaults.

cancel_build() {
    local status="$1"

    for id in $(sudo composer-cli compose list | grep "${status}" | cut -f1 -d' '); do
        echo "Cancelling ${id}"
        sudo composer-cli compose cancel "${id}"
    done
}

delete_builds() {
    for id in $(sudo composer-cli compose list | grep -v "^ID" | cut -f1 -d' '); do
        echo "Deleting ${id}"
        sudo composer-cli compose delete "${id}"
    done
}

remove_sources() {
    for src in $(sudo composer-cli sources list | grep -v appstream | grep -v baseos); do
        echo "Removing source ${src}"
        sudo composer-cli sources delete "${src}"
    done
}

cancel_build WAITING
cancel_build RUNNING
delete_builds
remove_sources
