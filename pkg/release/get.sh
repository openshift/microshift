#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

function get_base {
    grep "var Base" "${SCRIPT_DIR}/release.go"  | cut -d\" -f 2
}

function add_bases {
    base=$(get_base)
    sed "s/:$/:${base}/g" # some lines have "xxxxx:" + Base  like flannel
}

function get_image_list {

    cat $1 | grep "Image = map\[string\]string" -A 100 | grep '":' | cut -d\" -f4 | \
        add_bases
}

function get_images {
    arch=$1
    case $arch in
        x86_64|amd64) get_image_list "${SCRIPT_DIR}/release_amd64.go" ;;
        *) get_image_list "${SCRIPT_DIR}/release.go"              ;;
    esac
}

function usage {
    echo "usage:"
    echo "   get.sh base                  : prints the OKD base version for this MicroShift codebase"
    echo "   get.sh images <architecture> : prints image list used by this MicroShift codebase and architecture"
    exit 1
}

case $1 in
    base) get_base        ;;
    images) get_images $2 ;;
    *) usage
esac


