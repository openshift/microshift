#!/bin/bash
#
# Verify that all image references embedded in the release data are
# for valid registries.

verbose=false
if [ "$1" = "-v" ]; then
    verbose=true
fi

function debug() {
    if ! $verbose; then
        return
    fi
    echo "$*"
}

approved=true
for image in $(./pkg/release/get.sh images all); do
    case $image in
        quay.io/microshift/*)
            debug "$image OK";;
        quay.io/openshift-release-dev/*)
            debug "$image OK";;
        registry.redhat.io/openshift4/*)
            debug "$image OK";;
        registry.access.redhat.com/*)
            debug "$image OK";;
        registry.redhat.io/odf4/*)
            debug "$image OK";;
        *)
            echo "$image is not from an approved location" 1>&2
            approved=false;;
    esac
done

if ! $approved; then
    echo "Invalid image reference found" 1>&2
    exit 1
fi
