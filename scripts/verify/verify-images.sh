#!/bin/bash
#
# Verify that all image references embedded in the release data are
# for valid registries.
set -euo pipefail

verbose=false
if [ $# -gt 0 ] && [ "$1" = "-v" ]; then
    verbose=true
fi

function debug() {
    if ${verbose} ; then
        echo "$*"
    fi
}

approved=true
image_list=$(mktemp)
trap 'rm -f ${image_list}' EXIT

jq -r '.images | .[] | (input_filename) + " " + (.)' assets/release/release-*.json > "${image_list}"

while read -r source_file image; do
    case ${image} in
        quay.io/microshift/*)
            debug "${image} OK";;
        quay.io/openshift-release-dev/*)
            debug "${image} OK";;
        quay.io/lvms_dev/*) # This is a registry for LVMS image clones from CPaaS candidates
            debug "${image} OK";;
        registry.redhat.io/openshift4/*)
            debug "${image} OK";;
        registry.redhat.io/lvms4/*)
            debug "${image} OK";;
        registry.redhat.io/ubi9*)
            debug "${image} OK";;
        *)
            echo "${image} used in ${source_file} is not from an approved location" 1>&2
            approved=false;;
    esac
done < "${image_list}"

if ! ${approved}; then
    echo "Invalid image reference found" 1>&2
    exit 1
fi
exit 0
