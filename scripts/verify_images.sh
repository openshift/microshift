#!/bin/bash
#
# Verify that all image references embedded in the release data are
# for valid registries.

verbose=false
if [ "$1" = "-v" ]; then
    verbose=true
fi

function debug() {
    $verbose && echo "$*"
}

ROOTDIR=$(git rev-parse --show-toplevel)

RC=0
approved=true
image_list=$(mktemp)

function cleanup() {
    rm -f $image_list
}
trap cleanup EXIT
jq -r '.images | .[] | (input_filename) + " " + (.)' assets/release/release-*.json > $image_list

while read source_file image; do
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
        registry.redhat.io/lvms4/*)
            debug "$image OK";;
        *)
            echo "$image used in $source_file is not from an approved location" 1>&2
            approved=false;;
    esac
done < $image_list

if ! $approved; then
    echo "Invalid image reference found" 1>&2
    exit 1
fi
exit 0
