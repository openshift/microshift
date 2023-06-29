#!/bin/bash
#
# This script cancels any composer jobs and deletes failed and completed jobs.

cancel() {
    local status="$1"

    for id in $(sudo composer-cli compose list | grep "${status}" | cut -f1 -d' '); do
        echo "Cancelling ${id}"
        sudo composer-cli compose cancel "${id}"
    done
}

do_delete() {
    for id in $(sudo composer-cli compose list | grep -v "^ID" | cut -f1 -d' '); do
        echo "Deleting ${id}"
        sudo composer-cli compose delete "${id}"
    done
}

cancel WAITING
cancel RUNNING
do_delete
